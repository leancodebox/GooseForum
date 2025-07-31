package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"testing"
)

func TestGName(t *testing.T) {
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
	println(component.GenerateGooseNickname())
}
