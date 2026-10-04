package userservice

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
)

func CreateUser(username, password, email string, needValid bool, locale ...string) (*users.EntityComplete, error) {
	return CreateUserWithBinding(username, password, email, needValid, nil, locale...)
}

// CreateUserWithBinding retains a provider identity on the account so failed bindings can be retried.
func CreateUserWithBinding(username, password, email string, needValid bool, binding *userOAuth.Entity, locale ...string) (*users.EntityComplete, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.ToLower(strings.TrimSpace(email))
	userEntity := users.MakeUser(username, password, email)
	userEntity.Locale = normalizeUserLocale(locale...)
	userEntity.Nickname = GenerateGooseNickname()
	userEntity.RequiresEmailVerification = needValid
	if !needValid {
		userEntity.IsActivated = users.ActivationSuccess
	}
	if binding != nil {
		if binding.Provider == "" || binding.ProviderUid == "" {
			return nil, fmt.Errorf("OAuth identity is missing")
		}
		key := users.OAuthRegistrationKey(binding.Provider, binding.ProviderUid)
		userEntity.OAuthRegistrationKey = &key
	}
	err := func() error {
		if binding != nil {
			recovered, err := users.FindOAuthRegistration(*userEntity.OAuthRegistrationKey)
			if err != nil {
				return err
			}
			if recovered != nil {
				userEntity = recovered
			}
		}
		if userEntity.Id == 0 {
			if err := users.CheckIdentityAvailable(username, email, 0); err != nil {
				return err
			}
			if err := users.Create(userEntity); err != nil {
				return err
			}
		}
		if binding != nil {
			binding.UserId = userEntity.Id
			existing := userOAuth.GetByProviderAndUID(binding.Provider, binding.ProviderUid)
			if existing != nil {
				if existing.UserId != userEntity.Id {
					return fmt.Errorf("OAuth identity belongs to another account")
				}
			} else if err := userOAuth.Create(binding); err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}
	pointservice.InitUserPoints(userEntity.Id, 100)
	if err := userStatistics.EnsureInitialized(userEntity.Id); err != nil {
		return nil, err
	}
	if userEntity.Id == 1 {
		if err := FirstUserInit(userEntity); err != nil {
			return nil, err
		}
	}
	if binding != nil {
		if err := users.CompleteOAuthRegistration(userEntity.Id); err != nil {
			return nil, err
		}
		userEntity.OAuthRegistrationKey = nil
	}
	return userEntity, nil
}

func normalizeUserLocale(values ...string) string {
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return ""
	}
	return i18n.Normalize(values[0])
}

// GenerateGooseNickname creates a compact random default nickname.
func GenerateGooseNickname() string {
	prefixes := []string{
		"鹅", "大白鹅", "灰鹅", "小鹅", "鹅宝",
		"Goose", "Gander", "Gosling", "Honker",
	}
	prefix := prefixes[rand.Intn(len(prefixes))]

	now := time.Now()
	timestamp := now.UnixMilli()
	randomPart := rand.Intn(1296)

	timestamp36 := strconv.FormatInt(timestamp, 36)
	random36 := strconv.FormatInt(int64(randomPart), 36)

	if len(random36) < 2 {
		random36 = "0" + random36
	}

	return fmt.Sprintf("%s%s%s", prefix, timestamp36, random36)
}
