package component

import (
	"encoding/json"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

func TestFailDataCodeSerializesMessageCodeAndParams(t *testing.T) {
	result := FailDataCode(
		MessageUploadCooldown,
		MessageParams{"minutes": 10, "availableAt": "2026-05-28 12:00:00"},
	)

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if payload["messageCode"] != string(MessageUploadCooldown) {
		t.Fatalf("messageCode = %v, want %s", payload["messageCode"], MessageUploadCooldown)
	}
	params, ok := payload["params"].(map[string]any)
	if !ok {
		t.Fatalf("params missing or wrong type: %T", payload["params"])
	}
	if params["minutes"].(float64) != 10 {
		t.Fatalf("params.minutes = %v, want 10", params["minutes"])
	}
}

func TestFailDataErrorUsesMessageError(t *testing.T) {
	err := NewMessageError(
		MessageAuthPasswordTooShort,
		"密码长度不能少于8位",
		MessageParams{"minLength": 8},
	)

	result := FailDataError(err)

	if result.MessageCode != MessageAuthPasswordTooShort {
		t.Fatalf("messageCode = %s, want %s", result.MessageCode, MessageAuthPasswordTooShort)
	}
	if result.Params["minLength"] != 8 {
		t.Fatalf("params.minLength = %v, want 8", result.Params["minLength"])
	}
}

func TestCheckUserPermissionIncludesActionCode(t *testing.T) {
	user := &users.EntityComplete{
		Id:       1,
		IsFrozen: users.StatusFrozen,
	}

	code, err := CheckUserPermission(user, PermissionActionPost)
	if code != 403 {
		t.Fatalf("code = %d, want 403", code)
	}

	result := FailDataError(err)
	if result.MessageCode != MessagePermissionUserFrozen {
		t.Fatalf("messageCode = %s, want %s", result.MessageCode, MessagePermissionUserFrozen)
	}
	if result.Params["actionCode"] != "post" {
		t.Fatalf("params.actionCode = %v, want post", result.Params["actionCode"])
	}
	if result.Params["action"] != "发帖" {
		t.Fatalf("params.action = %v, want 发帖", result.Params["action"])
	}
}
