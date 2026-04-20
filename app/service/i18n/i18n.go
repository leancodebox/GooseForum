package i18n

import (
	"io/fs"
	"log/slog"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	bundle *i18n.Bundle
	once   sync.Once
)

// Init initializes the i18n bundle using the provided fs.FS
func Init(fsys fs.FS) {
	once.Do(func() {
		bundle = i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

		// Load all translation files from the fs.FS
		// Assuming files are in "locales" directory
		_, err := bundle.LoadMessageFileFS(fsys, "locales/active.en.yaml")
		if err != nil {
			slog.Error("Failed to load en translations", "err", err)
		}
		_, err = bundle.LoadMessageFileFS(fsys, "locales/active.zh.yaml")
		if err != nil {
			slog.Error("Failed to load zh translations", "err", err)
		}
		_, err = bundle.LoadMessageFileFS(fsys, "locales/active.ja.yaml")
		if err != nil {
			slog.Error("Failed to load ja translations", "err", err)
		}
	})
}

// GetLocalizer returns a new localizer for the given language
func GetLocalizer(lang string) *i18n.Localizer {
	if bundle == nil {
		// Fallback if not initialized (though Init should be called)
		return nil
	}
	return i18n.NewLocalizer(bundle, lang)
}
