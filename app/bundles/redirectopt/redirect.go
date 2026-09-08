// Package redirectopt validates browser redirects that must remain within the
// current origin.
package redirectopt

import (
	"net/url"
	"strings"
	"unicode"
)

func Local(value string) string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") || strings.Contains(value, "\\") {
		return ""
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return ""
		}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.User != nil {
		return ""
	}
	return value
}
