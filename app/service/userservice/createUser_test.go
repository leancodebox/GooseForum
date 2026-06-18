package userservice

import "testing"

func TestGenerateName(t *testing.T) {
	for range 4 {
		if name := GenerateGooseNickname(); name == "" {
			t.Fatal("expected generated nickname")
		}
	}
}
