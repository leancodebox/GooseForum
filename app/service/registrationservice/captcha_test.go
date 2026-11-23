package registrationservice

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

type captchaTransport func(*http.Request) (*http.Response, error)

func (fn captchaTransport) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestHostedAuthProof(t *testing.T) {
	original := captchaHTTP
	t.Cleanup(func() { captchaHTTP = original })
	for _, tc := range []struct {
		provider, action, response string
		want                       bool
	}{
		{"hcaptcha", "register", `{"success":true,"hostname":"forum.example"}`, true},
		{"recaptcha_v2", "register", `{"success":true,"hostname":"forum.example"}`, true},
		{"recaptcha_v3", "register", `{"success":true,"hostname":"forum.example","action":"register","score":0.5}`, true},
		{"recaptcha_v3", "register", `{"success":true,"hostname":"forum.example","action":"register","score":0.4}`, false},
		{"recaptcha_v3", "register", `{"success":true,"hostname":"forum.example","action":"login","score":1}`, false},
		{"recaptcha_v3", "register", `{"success":true,"hostname":"forum.example","action":"register"}`, false},
		{"recaptcha_v2", "register", `{"success":true,"hostname":"other.example"}`, false},
		{"hcaptcha", "register", `{"success":false,"hostname":"forum.example"}`, false},
		{"recaptcha_v2", "login", `{"success":true,"hostname":"forum.example"}broken`, false},
		{"recaptcha_v3", "register", `{"success":true,"hostname":"forum.example","action":"register","score":10}`, false},
		{"recaptcha_v3", "login", `{"success":true,"hostname":"forum.example","action":"login","score":0.9,"error-codes":["Over free quota."]}`, false},
		{"recaptcha_v3", "login", `{"success":true,"hostname":"forum.example","action":"register","score":1}`, false},
		{"recaptcha_v3", "login", `{"success":true,"hostname":"forum.example","action":"login","score":0.5}`, true},
		{"recaptcha_v2", "login", `{"success":false,"error-codes":["timeout-or-duplicate"]}`, false},
	} {
		captchaHTTP = &http.Client{Transport: captchaTransport(func(req *http.Request) (*http.Response, error) {
			if req.Method != "POST" || req.PostFormValue("response") != "token" || req.PostFormValue("secret") != "secret" {
				t.Fatal("invalid verification request")
			}
			if tc.provider == "hcaptcha" && (req.URL.Host != "api.hcaptcha.com" || req.PostFormValue("sitekey") != "key") {
				t.Fatal("invalid hCaptcha verifier")
			}
			if tc.provider != "hcaptcha" && req.URL.Host != "www.google.com" {
				t.Fatal("invalid Google verifier")
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(tc.response))}, nil
		})}
		threshold := 0.5
		cfg := pageConfig.RegistrationCaptcha{Provider: tc.provider, SiteKey: "key", Secret: "secret", ScoreThreshold: &threshold}
		if got := VerifyHostedCaptcha(t.Context(), cfg, "token", "https://forum.example", tc.action); got != tc.want {
			t.Errorf("%s %s: got %v", tc.provider, tc.response, got)
		}
	}
}

func TestHostedVerificationTimeout(t *testing.T) {
	original := captchaHTTP
	t.Cleanup(func() { captchaHTTP = original })
	captchaHTTP = &http.Client{Transport: captchaTransport(func(*http.Request) (*http.Response, error) { return nil, context.DeadlineExceeded })}
	for _, action := range []string{"login", "register"} {
		if VerifyHostedCaptcha(t.Context(), pageConfig.RegistrationCaptcha{Provider: "hcaptcha", SiteKey: "key", Secret: "secret"}, "token", "https://forum.example", action) {
			t.Fatal("timeout accepted")
		}
	}
}

func TestRegistrationCaptchaSettings(t *testing.T) {
	old := &pageConfig.RegistrationCaptcha{Provider: "recaptcha_v3", SiteKey: "site", Secret: "saved"}
	var settings pageConfig.SecurityAndRegistration
	if err := PrepareCaptchaSettings(&settings, old); err != nil || settings.RegistrationCaptcha != old {
		t.Fatal("older admin client lost settings")
	}
	cfg := &pageConfig.RegistrationCaptcha{Provider: "recaptcha_v3", SiteKey: "site"}
	settings.RegistrationCaptcha = cfg
	if err := PrepareCaptchaSettings(&settings, old); err != nil || cfg.Secret != "saved" || *cfg.ScoreThreshold != 0.5 {
		t.Fatal("failed to retain credentials or default score")
	}
	cfg.Provider = "hcaptcha"
	cfg.Secret = ""
	if PrepareCaptchaSettings(&settings, old) == nil {
		t.Fatal("reused credentials across providers")
	}
	cfg.Provider = "none"
	if err := PrepareCaptchaSettings(&settings, old); err != nil || cfg.SiteKey != "" || cfg.Secret != "" {
		t.Fatal("failed to clear credentials")
	}
}
