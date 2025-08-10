package controllers

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
)

func TestGName(t *testing.T) {
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
}
