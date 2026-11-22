package pageConfig

import (
	"fmt"
	"slices"
)

var builtinAvatarFiles = func() []string {
	files := make([]string, 12)
	for i := range files {
		files[i] = fmt.Sprintf("/static/pic/%d.webp", i+1)
	}
	return files
}()

// BuiltinAvatarFiles is the complete catalog, including disabled presets.
func BuiltinAvatarFiles() []string {
	return slices.Clone(builtinAvatarFiles)
}

// EnabledAvatarFiles is shared by manual selection and automatic assignment.
func (c SiteSettingsConfig) EnabledAvatarFiles() []string {
	if c.AvatarPool == nil {
		if c.DefaultAvatarFile != "" {
			return []string{c.DefaultAvatarFile}
		}
		return BuiltinAvatarFiles()
	}
	files := make([]string, 0, len(c.AvatarPool))
	for _, option := range c.AvatarPool {
		if option.Enabled {
			files = append(files, option.File)
		}
	}
	return files
}
