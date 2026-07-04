// Package i18n provides lightweight server-side translation for the
// server-rendered surfaces (SEO / no-JS templates, the activation page and
// e-mail bodies). The Vue SPA has its own vue-i18n dictionaries under
// resource/src/locales; this package covers only what the Go side renders.
//
// Supported locales mirror the frontend (resource/src/runtime/i18n.ts):
// zh (source / fallback), en and ja.
package i18n

import (
	"embed"
	"encoding/json"
	"log/slog"
	"strings"
)

//go:embed locales/*.json
var localeFS embed.FS

// SupportedLocales lists the locales the server can render, in preference order.
var SupportedLocales = []string{"zh", "en", "ja"}

// Fallback is used when a locale or a key is missing. It matches the frontend
// fallbackLocale so the two stay consistent.
const Fallback = "zh"

var dictionaries = map[string]map[string]string{}

func init() {
	for _, lang := range SupportedLocales {
		data, err := localeFS.ReadFile("locales/" + lang + ".json")
		if err != nil {
			slog.Error("i18n: failed to read locale file", "lang", lang, "err", err)
			continue
		}
		messages := map[string]string{}
		if err := json.Unmarshal(data, &messages); err != nil {
			slog.Error("i18n: failed to parse locale file", "lang", lang, "err", err)
			continue
		}
		dictionaries[lang] = messages
	}
}

// Normalize maps an arbitrary language value (a ?lang query value, the "lang"
// cookie, or a raw Accept-Language header) to one of SupportedLocales. It takes
// the primary subtag, e.g. "en-US,en;q=0.9" -> "en", and returns Fallback when
// nothing matches. It mirrors normalizeLocale/detectLocale on the frontend.
func Normalize(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return Fallback
	}
	primary := strings.FieldsFunc(value, func(r rune) bool {
		return r == '-' || r == '_' || r == ',' || r == ';' || r == ' '
	})
	if len(primary) == 0 {
		return Fallback
	}
	short := primary[0]
	for _, lang := range SupportedLocales {
		if short == lang {
			return lang
		}
	}
	return Fallback
}

// T returns the translation for key in the given language. lang may be any raw
// value accepted by Normalize. Missing keys fall back to the Fallback locale and
// finally to the key itself, so a missing translation is visible but never
// blank. Optional args are treated as alternating name/value pairs and replace
// {name} placeholders in the template, matching the frontend's {name} syntax.
func T(lang, key string, args ...any) string {
	lang = Normalize(lang)

	template, ok := lookup(lang, key)
	if !ok {
		template, ok = lookup(Fallback, key)
	}
	if !ok {
		template = key
	}
	if len(args) == 0 {
		return template
	}
	return interpolate(template, args)
}

// Func returns a translator bound to lang, suitable for passing into a template
// as a func value (e.g. {{call .T "key"}}).
func Func(lang string) func(key string, args ...any) string {
	normalized := Normalize(lang)
	return func(key string, args ...any) string {
		return T(normalized, key, args...)
	}
}

func lookup(lang, key string) (string, bool) {
	if dict, ok := dictionaries[lang]; ok {
		if value, ok := dict[key]; ok {
			return value, true
		}
	}
	return "", false
}

func interpolate(template string, args []any) string {
	replacements := make([]string, 0, len(args))
	for i := 0; i+1 < len(args); i += 2 {
		name, ok := args[i].(string)
		if !ok {
			continue
		}
		replacements = append(replacements, "{"+name+"}", toString(args[i+1]))
	}
	return strings.NewReplacer(replacements...).Replace(template)
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return strings.TrimSpace(jsonScalar(v))
	}
}

// jsonScalar renders numbers/bools without importing fmt for the common path.
func jsonScalar(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return strings.Trim(string(data), `"`)
}
