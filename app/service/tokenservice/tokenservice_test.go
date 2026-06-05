package tokenservice

import (
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func withJWTKey(t *testing.T, key string) {
	t.Helper()
	old := preferences.GetString("jwtopt.key", "")
	preferences.Set("jwtopt.key", key)
	t.Cleanup(func() {
		preferences.Set("jwtopt.key", old)
	})
}

func TestActivationTokenLifecycle(t *testing.T) {
	withJWTKey(t, "activation-test-key")

	token, err := GenerateActivationToken(12, "user@example.com")
	if err != nil {
		t.Fatalf("GenerateActivationToken failed: %v", err)
	}

	claims, err := ParseActivationToken(token)
	if err != nil {
		t.Fatalf("ParseActivationToken failed: %v", err)
	}
	if claims.UserId != 12 {
		t.Fatalf("UserId = %d, want 12", claims.UserId)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("Email = %q, want user@example.com", claims.Email)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 23*time.Hour {
		t.Fatalf("activation token expiry should be close to 24h, got %v", claims.ExpiresAt)
	}
}

func TestGenerateActivationTokenByUser(t *testing.T) {
	withJWTKey(t, "activation-user-key")

	token, err := GenerateActivationTokenByUser(users.EntityComplete{
		Id:    99,
		Email: "entity@example.com",
	})
	if err != nil {
		t.Fatalf("GenerateActivationTokenByUser failed: %v", err)
	}

	claims, err := ParseActivationToken(token)
	if err != nil {
		t.Fatalf("ParseActivationToken failed: %v", err)
	}
	if claims.UserId != 99 || claims.Email != "entity@example.com" {
		t.Fatalf("claims = {%d %q}, want entity user", claims.UserId, claims.Email)
	}
}

func TestPasswordResetTokenLifecycle(t *testing.T) {
	withJWTKey(t, "password-reset-test-key")

	token, err := GeneratePasswordResetToken(34, "reset@example.com")
	if err != nil {
		t.Fatalf("GeneratePasswordResetToken failed: %v", err)
	}

	claims, err := ParsePasswordResetToken(token)
	if err != nil {
		t.Fatalf("ParsePasswordResetToken failed: %v", err)
	}
	if claims.UserId != 34 {
		t.Fatalf("UserId = %d, want 34", claims.UserId)
	}
	if claims.Email != "reset@example.com" {
		t.Fatalf("Email = %q, want reset@example.com", claims.Email)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) <= 29*time.Minute {
		t.Fatalf("password reset token expiry should be close to 30m, got %v", claims.ExpiresAt)
	}
}

func TestTokenParsingRejectsInvalidInput(t *testing.T) {
	withJWTKey(t, "reject-test-key")

	if _, err := ParseActivationToken("not-a-token"); err == nil {
		t.Fatalf("expected invalid activation token error")
	}
	if _, err := ParsePasswordResetToken("not-a-token"); err == nil {
		t.Fatalf("expected invalid password reset token error")
	}
}
