package outbound

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const Path = "/outbound"

// Parse accepts only unambiguous HTTP destinations, never executable schemes.
func Parse(raw string) (*url.URL, error) {
	if len(raw) > 8192 || strings.Contains(raw, "\\") || strings.ContainsFunc(raw, unicode.IsControl) {
		return nil, fmt.Errorf("invalid destination")
	}
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || strings.ContainsAny(u.Host, " %") {
		return nil, fmt.Errorf("invalid destination")
	}
	return u, nil
}

func Link(raw string) string {
	// Keep invalid destinations behind the gateway too; it will reject them.
	return Path + "?url=" + url.QueryEscape(raw)
}

func Allowed(u *url.URL, domains []string) bool {
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	for _, domain := range domains {
		domain = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
		if domain != "" && domain == host {
			return true
		}
	}
	return false
}

func NormalizeDomains(domains []string) ([]string, error) {
	if len(domains) > 200 {
		return nil, fmt.Errorf("too many domains")
	}
	result := make([]string, 0, len(domains))
	seen := make(map[string]bool)
	for _, value := range domains {
		value = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
		if value == "" {
			continue
		}
		if strings.ContainsAny(value, "/:@?#*\\") {
			return nil, fmt.Errorf("expected hostname")
		}
		u, err := Parse("https://" + value)
		if err != nil || u.Hostname() != value {
			return nil, fmt.Errorf("invalid hostname")
		}
		if !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result, nil
}
