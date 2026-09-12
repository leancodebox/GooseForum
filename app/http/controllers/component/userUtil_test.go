package component

import "testing"

func TestValidateUsernameAllowsFourToThirtyTwoSafeCharacters(t *testing.T) {
	for _, username := range []string{"abcd", "user-name", "user_name", "1234"} {
		if !ValidateUsername(username) {
			t.Errorf("ValidateUsername(%q) = false", username)
		}
	}
	for _, username := range []string{"abc", "user name", "用户", "abcdefghijklmnopqrstuvwxyz1234567"} {
		if ValidateUsername(username) {
			t.Errorf("ValidateUsername(%q) = true", username)
		}
	}
}
