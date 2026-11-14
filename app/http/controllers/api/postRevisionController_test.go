package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/leancodebox/GooseForum/app/service/postrevisionservice"
)

func TestPostRevisionErrorUsesHTTPVisibilityContract(t *testing.T) {
	if response := postRevisionError(postrevisionservice.ErrNotFound); response.Code != http.StatusNotFound {
		t.Fatalf("not found status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if response := postRevisionError(postrevisionservice.ErrForbidden); response.Code != http.StatusForbidden {
		t.Fatalf("forbidden status = %d, want %d", response.Code, http.StatusForbidden)
	}
	if response := postRevisionError(errors.New("database")); response.Code != http.StatusOK {
		t.Fatalf("internal status = %d, want %d", response.Code, http.StatusOK)
	}
}
