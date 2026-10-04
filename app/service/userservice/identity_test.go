package userservice

import (
	"errors"
	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"gorm.io/gorm"
	"sync"
	"testing"
)

func TestConcurrentCreateRejectsCaseInsensitiveIdentityDuplicates(t *testing.T) {
	setupCreateUserTestDB(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	created := 0
	for i := range 8 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name, email := "identity-race", "race@example.com"
			if i%2 == 0 {
				name, email = "IDENTITY-RACE", "RACE@EXAMPLE.COM"
			}
			_, err := CreateUser(name, "password", email, false)
			if err == nil {
				mu.Lock()
				created++
				mu.Unlock()
			} else if !errors.Is(err, users.ErrUsernameExists) && !errors.Is(err, users.ErrEmailExists) {
				t.Errorf("create: %v", err)
			}
		}(i)
	}
	wg.Wait()
	if created != 1 {
		t.Fatalf("created %d accounts", created)
	}
}

func TestOAuthBindingFailureCanRecoverSameAccount(t *testing.T) {
	setupCreateUserTestDB(t)
	conn := db.Connect()
	if err := conn.AutoMigrate(&userOAuth.Entity{}); err != nil {
		t.Fatal(err)
	}
	const callback = "test:registration-binding-failure"
	if err := conn.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "user_o_auth" {
			tx.AddError(errors.New("binding failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Callback().Create().Remove(callback) })
	_, err := CreateUserWithBinding("failed-binding", "password", "binding@example.com", false, &userOAuth.Entity{Provider: "test", ProviderUid: "test-user"})
	if err == nil {
		t.Fatal("binding failure was ignored")
	}
	if !users.ExistUsername("failed-binding") {
		t.Fatal("account should remain recoverable")
	}
	pending, err := users.FindOAuthRegistration(users.OAuthRegistrationKey("test", "test-user"))
	if err != nil || pending == nil {
		t.Fatalf("pending account: %v", err)
	}
	_ = conn.Callback().Create().Remove(callback)
	recovered, err := CreateUserWithBinding("different-name", "password", "changed@example.com", false, &userOAuth.Entity{Provider: "test", ProviderUid: "test-user"})
	if err != nil || recovered.Id != pending.Id || recovered.Email != pending.Email {
		t.Fatalf("recovery user=%+v error=%v", recovered, err)
	}
	if binding := userOAuth.GetByProviderAndUID("test", "test-user"); binding == nil || binding.UserId != pending.Id {
		t.Fatal("binding was not recovered")
	}
	remaining, err := users.FindOAuthRegistration(users.OAuthRegistrationKey("test", "test-user"))
	if err != nil || remaining != nil {
		t.Fatal("completed registration retained recovery marker")
	}
}

func TestOAuthCompletionFailureRetriesExistingBinding(t *testing.T) {
	setupCreateUserTestDB(t)
	conn := db.Connect()
	if err := conn.AutoMigrate(&userOAuth.Entity{}); err != nil {
		t.Fatal(err)
	}
	const callback = "test:registration-completion-failure"
	if err := conn.Callback().Update().Before("gorm:update").Register(callback, func(tx *gorm.DB) {
		values, ok := tx.Statement.Dest.(map[string]any)
		if ok && tx.Statement.Table == "users" {
			if _, completing := values["oauth_registration_key"]; completing {
				tx.AddError(errors.New("completion failure"))
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Callback().Update().Remove(callback) })
	_, err := CreateUserWithBinding("completion-retry", "password", "completion@example.com", false, &userOAuth.Entity{Provider: "retry", ProviderUid: "completion-user"})
	if err == nil {
		t.Fatal("completion failure was ignored")
	}
	pending, err := users.FindOAuthRegistration(users.OAuthRegistrationKey("retry", "completion-user"))
	if err != nil || pending == nil {
		t.Fatalf("pending account: %v", err)
	}
	if err := conn.Model(&userStatistics.Entity{}).Where("user_id = ?", pending.Id).Update("topic_count", 7).Error; err != nil {
		t.Fatal(err)
	}
	_ = conn.Callback().Update().Remove(callback)
	recovered, err := CreateUserWithBinding("ignored", "password", "ignored@example.com", false, &userOAuth.Entity{Provider: "retry", ProviderUid: "completion-user"})
	if err != nil || recovered.Id != pending.Id || recovered.OAuthRegistrationKey != nil {
		t.Fatalf("completion retry user=%+v error=%v", recovered, err)
	}
	var count int64
	conn.Model(&userOAuth.Entity{}).Where("provider = ? AND provider_uid = ?", "retry", "completion-user").Count(&count)
	if count != 1 {
		t.Fatalf("retry created %d bindings", count)
	}
	if stats := userStatistics.Get(recovered.Id); stats.TopicCount != 7 {
		t.Fatalf("retry reset existing statistics: %+v", stats)
	}
}
