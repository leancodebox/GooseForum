package captchaOpt

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/closer"
	paniclog "github.com/leancodebox/GooseForum/app/bundles/recovery"
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
	captchaExpiration  = time.Minute * 3 // 验证码3分钟过期
	cleanupStopCh      = make(chan struct{})
	cleanupOnce        sync.Once
	cleanupWg          sync.WaitGroup
)

// StartCleanup starts the expired captcha cleanup worker.
func StartCleanup() {
	cleanupOnce.Do(func() {
		closer.RegisterPriority(closer.PriorityCache, StopCleanup)
		cleanupWg.Go(func() {
			defer paniclog.Recover("captcha_cleanup")
			ticker := time.NewTicker(time.Minute) // 每分钟清理一次
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					store.cleanup()
				case <-cleanupStopCh:
					return
				}
			}
		})
	})
}

// StopCleanup stops the expired captcha cleanup worker.
func StopCleanup() error {
	select {
	case <-cleanupStopCh:
	default:
		close(cleanupStopCh)
	}
	cleanupWg.Wait()
	return nil
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

// GenerateCaptcha 生成验证码
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

// VerifyCaptcha 验证验证码
func VerifyCaptcha(captchaId, captchaCode string) bool {
	if captchaId == "" || captchaCode == "" {
		return false
	}
	return customCaptchaStore.Verify(captchaId, captchaCode, true)
}
