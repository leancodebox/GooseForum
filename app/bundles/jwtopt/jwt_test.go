package jwtopt

import (
	"fmt"
	"testing"
	"time"
)

func TestCreateNewToken(t *testing.T) {
	token, err := CreateNewToken(123456, time.Second*2)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(token)
	time.Sleep(time.Second)
	userId, newToken, err := VerifyTokenWithFresh(token)
	fmt.Printf("userId %v newToken %v %v", userId, newToken, err)
	if err != nil {
		t.Error(err)
		return
	}

	userId, err = VerifyToken(token)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(userId)
	fmt.Println("token还可以用")

	time.Sleep(time.Second * 2)
	fmt.Println(token)
	userId, err = VerifyToken(token)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(userId)
}

func TestVerifyTokenWithFresh(t *testing.T) {
	token, err := CreateNewToken(123456, 15*time.Second)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("token created : ", token)
	userId, newToken, err := VerifyTokenWithFresh(token)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("token can use")
	if newToken != token {
		token = newToken
	}
	fmt.Println(userId)

	time.Sleep(time.Second * 2)
	fmt.Println(token)
	userId, err = VerifyToken(token)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(userId)
	fmt.Println("success")
}
