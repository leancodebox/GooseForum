package userservice

import "testing"

func TestGenerateName(t *testing.T) {
	for i := 0; i < 4; i++ {
		if name := GenerateGooseNickname(); name == "" {
			t.Fatal("expected generated nickname")
		}
	}
}
