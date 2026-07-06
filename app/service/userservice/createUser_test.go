package userservice

import (
	"testing"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func setupCreateUserTestDB(t *testing.T) {
	t.Helper()
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := db.Connect()
	if err := conn.AutoMigrate(&users.EntityComplete{}, &userStatistics.Entity{}); err != nil {
		t.Fatalf("migrate user tables: %v", err)
	}
	conn.Where("1 = 1").Delete(&userStatistics.Entity{})
	conn.Where("1 = 1").Delete(&users.EntityComplete{})
}

func TestCreateUserStoresNormalizedLocale(t *testing.T) {
	setupCreateUserTestDB(t)

	user, err := CreateUser("lang-user", "password", "lang@example.com", false, "en-US")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.Locale != "en" {
		t.Fatalf("Locale = %q, want en", user.Locale)
	}
}

func TestCreateUserKeepsLocaleEmptyWhenMissing(t *testing.T) {
	setupCreateUserTestDB(t)

	user, err := CreateUser("empty-locale", "password", "empty@example.com", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.Locale != "" {
		t.Fatalf("Locale = %q, want empty", user.Locale)
	}
}

func TestGenerateName(t *testing.T) {
	for range 4 {
		if name := GenerateGooseNickname(); name == "" {
			t.Fatal("expected generated nickname")
		}
	}
}
