package registrationservice

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"golang.org/x/net/idna"
)

var (
	ErrClosed          = errors.New("registration is disabled")
	ErrEmailRequired   = errors.New("a valid email is required for registration")
	ErrDomain          = errors.New("email domain is not allowed")
	ErrRateLimited     = errors.New("too many requests; please try again later")
	ErrMailUnavailable = errors.New("email verification is unavailable")
)

func CheckVerificationMail(required bool) error {
	if required {
		if err := mailservice.CheckConfigured(); err != nil {
			return fmt.Errorf("%w: %v", ErrMailUnavailable, err)
		}
	}
	return nil
}

func NormalizeDomain(domain string) (string, error) {
	domain = strings.TrimSpace(strings.ToLower(domain))
	if domain == "" || strings.ContainsAny(domain, "@/:* ") || strings.HasSuffix(domain, ".") {
		return "", ErrDomain
	}
	value, err := idna.Lookup.ToASCII(domain)
	if err != nil || !strings.Contains(value, ".") || len(value) > 253 {
		return "", ErrDomain
	}
	for label := range strings.SplitSeq(value, ".") {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return "", ErrDomain
		}
		for _, char := range label {
			if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-') {
				return "", ErrDomain
			}
		}
	}
	return value, nil
}

func NormalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 128 {
		return "", ErrEmailRequired
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", ErrEmailRequired
	}
	domain, err := NormalizeDomain(parts[1])
	if err != nil {
		return "", ErrEmailRequired
	}
	value := parts[0] + "@" + domain
	if len(value) > 128 {
		return "", ErrEmailRequired
	}
	return value, nil
}

func ValidateEmail(email string, config pageConfig.SecurityAndRegistration) (string, error) {
	value, err := NormalizeEmail(email)
	if err != nil {
		return "", err
	}
	if len(config.AllowedDomains) == 0 {
		return value, nil
	}
	domain := strings.Split(value, "@")[1]
	for _, allowed := range config.AllowedDomains {
		candidate, err := NormalizeDomain(allowed)
		if err == nil && candidate == domain {
			return value, nil
		}
	}
	return "", ErrDomain
}

func CheckPolicy(email string, oauth bool) (string, pageConfig.SecurityAndRegistration, error) {
	config := hotdataserve.GetSecuritySettingsConfigCache()
	if !config.EnableSignup {
		return "", config, ErrClosed
	}
	value := ""
	var err error
	if email != "" || !oauth || config.EnableEmailVerification || len(config.AllowedDomains) > 0 {
		value, err = ValidateEmail(email, config)
		if err != nil {
			return "", config, err
		}
	}
	return value, config, nil
}

func ConsumeSignup(email, ip string, config pageConfig.SecurityAndRegistration) error {
	return signupLimits.consume(time.Now(), []quota{
		{"ip:" + ip, config.RegistrationIPLimit}, {"email:" + email, config.RegistrationEmailLimit}, {"global", config.RegistrationGlobalLimit},
	})
}

func Check(email string, ip string, oauth bool) (string, pageConfig.SecurityAndRegistration, error) {
	value, config, err := CheckPolicy(email, oauth)
	if err != nil {
		return "", config, err
	}
	err = ConsumeSignup(value, ip, config)
	return value, config, err
}

func ValidateSettings(config *pageConfig.SecurityAndRegistration) error {
	if len(config.AllowedDomains) > 100 {
		return errors.New("too many allowed email domains")
	}
	for _, limit := range []int{config.RegistrationIPLimit, config.RegistrationEmailLimit, config.RegistrationGlobalLimit} {
		if limit < 0 || limit > 100000 {
			return errors.New("registration limits must be between 0 and 100000")
		}
	}
	domains := make([]string, 0, len(config.AllowedDomains))
	seen := make(map[string]bool)
	for _, domain := range config.AllowedDomains {
		value, err := NormalizeDomain(domain)
		if err != nil {
			return err
		}
		if !seen[value] {
			domains = append(domains, value)
			seen[value] = true
		}
	}
	config.AllowedDomains = domains
	return nil
}

type quota struct {
	key   string
	limit int
}
type counter struct {
	count   int
	expires time.Time
}
type limiter struct {
	mu      sync.Mutex
	entries map[string]counter
}

var signupLimits = limiter{entries: make(map[string]counter)}
var mailLimits = limiter{entries: make(map[string]counter)}

func (l *limiter) consume(now time.Time, quotas []quota) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, entry := range l.entries {
		if !now.Before(entry.expires) {
			delete(l.entries, key)
		}
	}
	needed := 0
	for _, q := range quotas {
		if q.limit == 0 || strings.HasSuffix(q.key, ":") {
			continue
		}
		entry, exists := l.entries[q.key]
		if exists && entry.count >= q.limit {
			return ErrRateLimited
		}
		if !exists {
			needed++
		}
	}
	// Reject new buckets when full so churn cannot reset an existing quota.
	if len(l.entries)+needed > 10000 {
		return ErrRateLimited
	}
	for _, q := range quotas {
		if q.limit == 0 || strings.HasSuffix(q.key, ":") {
			continue
		}
		entry := l.entries[q.key]
		if entry.count == 0 {
			entry.expires = now.Add(time.Hour)
		}
		entry.count++
		l.entries[q.key] = entry
	}
	return nil
}

// AllowMail limits reset and verification mail independently from registration.
func AllowMail(ip, email string) error {
	if len(email) > 128 {
		return ErrRateLimited
	}
	return mailLimits.consume(time.Now(), []quota{{"ip:" + ip, 20}, {"email:" + strings.ToLower(strings.TrimSpace(email)), 3}, {"global", 200}})
}
