package controllers

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/mojocn/base64Captcha"
)

// 验证码存储结构
type captchaStore struct {
	sync.RWMutex
	data map[string]captchaInfo
}

type captchaInfo struct {
	code      string    // 存储验证码答案
	expiredAt time.Time // 过期时间
}

// 实现 base64Captcha.Store 接口
type customStore struct {
	*captchaStore
}

func (s *customStore) Set(id string, value string) error {
	s.Lock()
	s.data[id] = captchaInfo{
		code:      value,
		expiredAt: time.Now().Add(captchaExpiration),
	}
	s.Unlock()
	return nil
}

func (s *customStore) Get(id string, clear bool) string {
	s.RLock()
	info, exists := s.data[id]
	s.RUnlock()

	if !exists {
		return ""
	}

	if time.Now().After(info.expiredAt) {
		if clear {
			s.Lock()
			delete(s.data, id)
			s.Unlock()
		}
		return ""
	}

	if clear {
		s.Lock()
		delete(s.data, id)
		s.Unlock()
	}

	return info.code
}

func (s *customStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return strings.EqualFold(v, answer)
}

// 全局验证码存储
var (
	store = &captchaStore{
		data: make(map[string]captchaInfo),
	}
	customCaptchaStore = &customStore{store}
	captchaExpiration  = time.Minute * 5 // 验证码5分钟过期
)

// 定期清理过期验证码
func init() {
	go func() {
		ticker := time.NewTicker(time.Minute) // 每分钟清理一次
		for range ticker.C {
			store.cleanup()
		}
	}()
}

// 清理过期验证码
func (s *captchaStore) cleanup() {
	s.Lock()
	defer s.Unlock()

	now := time.Now()
	for id, info := range s.data {
		if now.After(info.expiredAt) {
			delete(s.data, id)
		}
	}
}

// 生成验证码
func GenerateCaptcha() (string, string) {
	// 配置验证码参数
	driver := base64Captcha.NewDriverDigit(
		80,  // 高度
		240, // 宽度
		6,   // 验证码长度
		0.7, // 干扰强度
		80,  // 行数
	)

	// 生成验证码
	c := base64Captcha.NewCaptcha(driver, customCaptchaStore)
	id, b64s, _, err := c.Generate()
	if err != nil {
		slog.Error("generate captcha failed", "error", err)
		return "", ""
	}

	return id, b64s // 返回验证码ID和base64图片字符串
}

// 验证验证码
func VerifyCaptcha(captchaId, captchaCode string) bool {
	if captchaId == "" || captchaCode == "" {
		return false
	}
	return customCaptchaStore.Verify(captchaId, captchaCode, true)
}
