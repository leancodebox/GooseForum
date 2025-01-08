package preferences

import (
	"fmt"

	"github.com/spf13/cast"
)

func GetExclusivePreferences(prefix string) ExclusivePreferences {
	return ExclusivePreferences{root: prefix}
}

type ExclusivePreferences struct {
	root string
}

func (itself *ExclusivePreferences) realPath(path string) string {
	return fmt.Sprintf("%v.%v", itself.root, path)
}

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

func (itself *ExclusivePreferences) GetString(path string, defaultValue ...any) string {
	return cast.ToString(internalGet(itself.realPath(path), defaultValue...))
}

func (itself *ExclusivePreferences) GetInt(path string, defaultValue ...any) int {
	return cast.ToInt(internalGet(itself.realPath(path), defaultValue...))
}

func (itself *ExclusivePreferences) GetFloat64(path string, defaultValue ...any) float64 {
	return cast.ToFloat64(internalGet(itself.realPath(path), defaultValue...))
}

func (itself *ExclusivePreferences) GetInt64(path string, defaultValue ...any) int64 {
	return cast.ToInt64(internalGet(itself.realPath(path), defaultValue...))
}

func (itself *ExclusivePreferences) GetUint(path string, defaultValue ...any) uint {
	return cast.ToUint(internalGet(itself.realPath(path), defaultValue...))
}

func (itself *ExclusivePreferences) GetBool(path string, defaultValue ...any) bool {
	return cast.ToBool(internalGet(itself.realPath(path), defaultValue...))
}

func (itself *ExclusivePreferences) GetStringMapString(path string) map[string]string {
	return v.GetStringMapString(path)
}
