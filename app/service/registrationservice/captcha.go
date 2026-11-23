package registrationservice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func PrepareCaptchaSettings(settings *pageConfig.SecurityAndRegistration, previous *pageConfig.RegistrationCaptcha) error {
	cfg := settings.RegistrationCaptcha
	if cfg == nil {
		// Older admin clients omit the registration captcha when editing other policies.
		settings.RegistrationCaptcha = previous
		return nil
	}
	switch cfg.Provider {
	case "altcha", "image", "none":
		*cfg = pageConfig.RegistrationCaptcha{Provider: cfg.Provider}
		return nil
	case "hcaptcha", "recaptcha_v2", "recaptcha_v3":
	default:
		return errors.New("unknown registration captcha")
	}
	cfg.SiteKey = strings.TrimSpace(cfg.SiteKey)
	cfg.Secret = strings.TrimSpace(cfg.Secret)
	if cfg.Secret == "" && previous != nil && previous.Provider == cfg.Provider && previous.SiteKey == cfg.SiteKey {
		cfg.Secret = previous.Secret
	}
	if cfg.Secret == "" || cfg.SiteKey == "" {
		return errors.New("captcha credentials required")
	}
	if cfg.Provider == "recaptcha_v3" {
		if cfg.ScoreThreshold == nil {
			threshold := 0.5
			cfg.ScoreThreshold = &threshold
		}
		if *cfg.ScoreThreshold < 0 || *cfg.ScoreThreshold > 1 {
			return errors.New("invalid captcha score threshold")
		}
	} else {
		cfg.ScoreThreshold = nil
	}
	return nil
}

var captchaHTTP = &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

func VerifyHostedCaptcha(ctx context.Context, cfg pageConfig.RegistrationCaptcha, token, siteURL, action string) bool {
	if token == "" || len(token) > 16384 || cfg.Secret == "" || cfg.SiteKey == "" {
		return false
	}
	site, err := url.Parse(siteURL)
	if err != nil || (site.Scheme != "https" && site.Scheme != "http") || site.User != nil || site.Hostname() == "" {
		return false
	}
	endpoint := "https://www.google.com/recaptcha/api/siteverify"
	form := url.Values{"secret": {cfg.Secret}, "response": {token}}
	switch cfg.Provider {
	case "hcaptcha":
		endpoint = "https://api.hcaptcha.com/siteverify"
		form.Set("sitekey", cfg.SiteKey)
	case "recaptcha_v2", "recaptcha_v3":
	default:
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := captchaHTTP.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var result struct {
		Success    bool
		Hostname   string
		Action     string
		Score      *float64
		ErrorCodes []string `json:"error-codes"`
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	const maxResponse = 64 << 10
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	if err != nil || len(body) > maxResponse || json.Unmarshal(body, &result) != nil {
		return false
	}
	if !result.Success || len(result.ErrorCodes) != 0 || !strings.EqualFold(result.Hostname, site.Hostname()) {
		return false
	}
	if cfg.Provider == "recaptcha_v3" {
		return cfg.ScoreThreshold != nil && result.Action == action && result.Score != nil && *result.Score >= *cfg.ScoreThreshold && *result.Score <= 1
	}
	return true
}
