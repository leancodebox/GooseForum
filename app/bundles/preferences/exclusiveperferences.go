package preferences

import (
	"fmt"

	"github.com/spf13/cast"
)

// GetExclusivePreferences returns a namespaced preferences reader.
func GetExclusivePreferences(prefix string) ExclusivePreferences {
	return ExclusivePreferences{root: prefix}
}

// ExclusivePreferences reads settings below a fixed root path.
type ExclusivePreferences struct {
	root string
}

func (itself *ExclusivePreferences) realPath(path string) string {
	return fmt.Sprintf("%v.%v", itself.root, path)
}

// Get returns a string setting below the namespace.
func (itself *ExclusivePreferences) Get(path string, defaultValue ...any) string {
	return GetString(itself.realPath(path), defaultValue...)
}

func (itself *ExclusivePreferences) internalGet(path string, defaultValue ...any) any {
	path = itself.realPath(path)
	if !v.IsSet(path) || v.Get(path) == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	return v.Get(path)
}

// GetString returns a string setting below the namespace.
func (itself *ExclusivePreferences) GetString(path string, defaultValue ...any) string {
	return cast.ToString(internalGet(itself.realPath(path), defaultValue...))
}

// GetInt returns an int setting below the namespace.
func (itself *ExclusivePreferences) GetInt(path string, defaultValue ...any) int {
	return cast.ToInt(internalGet(itself.realPath(path), defaultValue...))
}

// GetFloat64 returns a float64 setting below the namespace.
func (itself *ExclusivePreferences) GetFloat64(path string, defaultValue ...any) float64 {
	return cast.ToFloat64(internalGet(itself.realPath(path), defaultValue...))
}

// GetInt64 returns an int64 setting below the namespace.
func (itself *ExclusivePreferences) GetInt64(path string, defaultValue ...any) int64 {
	return cast.ToInt64(internalGet(itself.realPath(path), defaultValue...))
}

// GetUint returns a uint setting below the namespace.
func (itself *ExclusivePreferences) GetUint(path string, defaultValue ...any) uint {
	return cast.ToUint(internalGet(itself.realPath(path), defaultValue...))
}

// GetBool returns a bool setting below the namespace.
func (itself *ExclusivePreferences) GetBool(path string, defaultValue ...any) bool {
	return cast.ToBool(internalGet(itself.realPath(path), defaultValue...))
}

// GetStringMapString returns a string map setting below the namespace.
func (itself *ExclusivePreferences) GetStringMapString(path string) map[string]string {
	return v.GetStringMapString(itself.realPath(path))
}
