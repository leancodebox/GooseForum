package userservice

import (
	"fmt"
	"testing"
)

func TestGenerateName(t *testing.T) {
	fmt.Println(GenerateGooseNickname())
	fmt.Println(GenerateGooseNickname())
	fmt.Println(GenerateGooseNickname())
	fmt.Println(GenerateGooseNickname())
}
