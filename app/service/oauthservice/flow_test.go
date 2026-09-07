package oauthservice

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOAuthFlowRoundTrip(t *testing.T) {
	startRequest := httptest.NewRequest(http.MethodGet, "https://forum.example.com/api/auth/github", nil)
	startResponse := httptest.NewRecorder()
	if err := StartFlow(startResponse, startRequest, "github", 42, "bind", "/settings?tab=binding"); err != nil {
		t.Fatal(err)
	}

	callbackRequest := httptest.NewRequest(http.MethodGet, "https://forum.example.com/api/auth/github/callback", nil)
	for _, cookie := range startResponse.Result().Cookies() {
		callbackRequest.AddCookie(cookie)
	}
	callbackResponse := httptest.NewRecorder()
	flow, err := ConsumeFlow(callbackResponse, callbackRequest, "github")
	if err != nil {
		t.Fatal(err)
	}
	if flow.Mode != "bind" || flow.UserID != 42 || flow.Redirect != "/settings?tab=binding" {
		t.Fatalf("flow = %#v", flow)
	}
}

func TestStartFlowRequiresLoginForBinding(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://forum.example.com/api/auth/github", nil)
	if err := StartFlow(httptest.NewRecorder(), req, "github", 0, "bind", ""); err == nil {
		t.Fatal("expected unauthenticated bind to fail")
	}
}

func TestSafeRedirect(t *testing.T) {
	for _, unsafe := range []string{"https://evil.example", "//evil.example", "javascript:alert(1)"} {
		if got := safeRedirect(unsafe); got != "" {
			t.Errorf("safeRedirect(%q) = %q", unsafe, got)
		}
	}
	if got := safeRedirect("/topics?id=1"); got != "/topics?id=1" {
		t.Fatalf("local redirect = %q", got)
	}
}
