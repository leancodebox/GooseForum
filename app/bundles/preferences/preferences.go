package preferences

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/bundles/fileopt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// Viper 库实例
var v *viper.Viper

//go:embed config.templ.toml
var configTempl []byte

func GenerateConfig() []byte {
	var b bytes.Buffer
	t := template.New("config.templ.toml")
	t = template.Must(t.Parse(string(configTempl)))
	err := t.Execute(&b, map[string]any{
		"AppSigningKey": algorithm.SafeGenerateSigningKey(32),
		"SigningKey":    algorithm.SafeGenerateSigningKey(32),
	})

	if err != nil {
		fmt.Println(err)
	}
	return b.Bytes()
}

// 初始化配置信息，完成对环境变量以及 conf 信息的加载
func init() {
	cfgPath := "config.toml"
	wd, _ := os.Getwd()
	if isTestMode() {
		dir, err := findConfigDirTest(wd, 6)
		if err != nil {
			slog.Error("preferences.test.search", "err", err)
			dir = wd
		}
		cfgPath = filepath.Join(dir, "config.toml")
		if !fileopt.IsExist(cfgPath) {
			if e := fileopt.PutContents(cfgPath, GenerateConfig()); e != nil {
				slog.Error("preferences.test.init", "err", e)
			} else {
				slog.Info("preferences.test.init", "path", cfgPath)
			}
		}
	} else {
		if !fileopt.IsExist(cfgPath) {
			if err := fileopt.PutContents(cfgPath, GenerateConfig()); err != nil {
				panic(err)
			}
		}
	}
	v = viper.New()
	v.SetConfigType("toml")
	v.AddConfigPath(filepath.Dir(cfgPath))
	configFlag := flag.String("config", cfgPath, "path to config file")
	v.SetConfigFile(*configFlag)
	if err := v.ReadInConfig(); err != nil {
		slog.Warn("ReadInConfig", "err", err)
	}
	v.WatchConfig()
}

func internalGet(path string, defaultValue ...any) any {
	// conf 或者环境变量不存在的情况
	if !v.IsSet(path) || v.Get(path) == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return nil
	}
	return v.Get(path)
}

func IsSet(path string) bool {
	return v.IsSet(path) && v.Get(path) != nil
}

// OpenConfigChangeEvent 开启监控配置文件⌚️
func OpenConfigChangeEvent() {
	v.OnConfigChange(runEvent)
}

var eventManagerLock sync.Mutex

var eventList []func(e fsnotify.Event)

func AddWatch(event func(e fsnotify.Event)) {
	eventManagerLock.Lock()
	defer eventManagerLock.Unlock()
	eventList = append(eventList, event)
}

func runEvent(e fsnotify.Event) {
	eventManagerLock.Lock()
	defer func() {
		eventManagerLock.Unlock()
		if r := recover(); r != nil {
			slog.Error("recover", "r", r)
		}
	}()
	for _, item := range eventList {
		item(e)
	}
}

// Get 获取配置项 允许使用点式获取，如：app.name
func Get(path string, defaultValue ...any) string {
	return GetString(path, defaultValue...)
}

// GetString 获取 String 类型的配置信息
func GetString(path string, defaultValue ...any) string {
	return cast.ToString(internalGet(path, defaultValue...))
}

// GetInt 获取 Int 类型的配置信息
func GetInt(path string, defaultValue ...any) int {
	return cast.ToInt(internalGet(path, defaultValue...))
}

// GetFloat64 获取 float64 类型的配置信息
func GetFloat64(path string, defaultValue ...any) float64 {
	return cast.ToFloat64(internalGet(path, defaultValue...))
}

// GetInt64 获取 Int64 类型的配置信息
func GetInt64(path string, defaultValue ...any) int64 {
	return cast.ToInt64(internalGet(path, defaultValue...))
}

// GetUint 获取 Uint 类型的配置信息
func GetUint(path string, defaultValue ...any) uint {
	return cast.ToUint(internalGet(path, defaultValue...))
}

// GetBool 获取 Bool 类型的配置信息
func GetBool(path string, defaultValue ...any) bool {
	return cast.ToBool(internalGet(path, defaultValue...))
}

// GetStringMapString 获取结构数据
func GetStringMapString(path string) map[string]string {
	return v.GetStringMapString(path)
}

func GetStringSlice(path string) []string {
	return v.GetStringSlice(path)
}

func GetIntSlice(path string) []int {
	return v.GetIntSlice(path)
}

func All() map[string]any {
	return v.AllSettings()
}

func isTestMode() bool {
	if strings.HasSuffix(os.Args[0], ".test") {
		return true
	}
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-test.") {
			return true
		}
	}
	return false
}

func findConfigDirTest(start string, maxDepth int) (string, error) {
	if fileopt.IsExist(filepath.Join(start, "config.toml")) {
		return start, nil
	}
	if fileopt.IsExist(filepath.Join(start, "go.mod")) {
		return start, nil
	}
	cur := start
	for range maxDepth {
		next := filepath.Dir(cur)
		if next == cur {
			break
		}
		cur = next
		if fileopt.IsExist(filepath.Join(cur, "go.mod")) {
			return cur, nil
		}
	}
	return "", fmt.Errorf("preferences: test mode cannot find go.mod within %d levels from %s", maxDepth, start)
}
