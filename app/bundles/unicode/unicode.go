// Package unicode decodes escaped Unicode sequences in strings.
package unicode

import (
	"regexp"
	"strconv"
)

// Decode replaces JavaScript-style \uXXXX escapes with UTF-8 characters.
func Decode(str string) string {
	re := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)

	return re.ReplaceAllStringFunc(str, func(m string) string {
		code, _ := strconv.ParseInt(m[2:], 16, 32)
		return string(rune(code))
	})
}
