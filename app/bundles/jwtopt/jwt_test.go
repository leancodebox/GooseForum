package jwtopt

import (
	"testing"
	"time"
)

func TestCreateNewToken(t *testing.T) {
	const userID uint64 = 123456
	token, err := CreateNewToken(userID, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	userId, newToken, err := VerifyTokenWithFresh(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != userID {
		t.Fatalf("VerifyTokenWithFresh userId = %d, want %d", userId, userID)
	}
	if newToken == "" {
		t.Fatal("expected refreshed token")
	}

	userId, err = VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != userID {
		t.Fatalf("VerifyToken userId = %d, want %d", userId, userID)
	}

	time.Sleep(1100 * time.Millisecond)
	if _, err = VerifyToken(token); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestVerifyTokenWithFresh(t *testing.T) {
	const userID uint64 = 123456
	token, err := CreateNewToken(123456, 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	userId, newToken, err := VerifyTokenWithFresh(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != userID {
		t.Fatalf("VerifyTokenWithFresh userId = %d, want %d", userId, userID)
	}
	if newToken != token {
		token = newToken
	}

	time.Sleep(time.Second * 2)
	userId, err = VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userId != userID {
		t.Fatalf("VerifyToken userId = %d, want %d", userId, userID)
	}
}
