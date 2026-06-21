package middleware

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
)

func TestCheckWritableAccountRequiresLogin(t *testing.T) {
	recorder := requestWithMiddleware(CheckWritableAccount, http.MethodPost)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	var body component.ResultStruct
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.MessageCode != component.MessageAuthRequired {
		t.Fatalf("messageCode = %q, want %q", body.MessageCode, component.MessageAuthRequired)
	}
}
