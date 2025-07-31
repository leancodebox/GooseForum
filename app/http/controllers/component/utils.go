package component

import (
	"cmp"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"math/rand"
	"strings"
	"time"
)

var (
	minPageSize int = 10
	maxPageSize int = 30
)

func BoundPageSize(pageSize int) int {
	return BoundPageSizeWithRange(pageSize, minPageSize, maxPageSize)
}

func BoundPageSizeWithRange[T cmp.Ordered](pageSize T, minN, maxN T) T {
	return min(max(pageSize, minN), maxN)
}

func BuildCanonicalHref(c *gin.Context) string {
	return GetBaseUri(c) + c.Request.URL.String()
}

func GetBaseUri(c *gin.Context) string {
	scheme := "https"
	if strings.HasPrefix(c.Request.Host, "localhost") {
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	return preferences.Get("server.url", host)
}

func GetHost(c *gin.Context) string {
	scheme := "https"
	if strings.HasPrefix(c.Request.Host, "localhost") {
		scheme = "http"
	}
	host := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	return preferences.Get("server.url", host)
}

// GenerateGooseNickname 新增生成鹅相关昵称的函数
func GenerateGooseNickname() string {
	prefixes := []string{
		"鹅", "大白鹅", "灰鹅", "小鹅", "鹅宝",
		"Goose", "Gander", "Gosling", "Honker",
	}
	prefix := prefixes[rand.Intn(len(prefixes))]
	// 使用纳秒级时间戳+随机数确保唯一性
	now := time.Now()
	timestamp := now.UnixNano()
	randomPart := rand.Intn(1000)
	// 组合成16进制字符串
	return fmt.Sprintf("%s%x%03d", prefix, timestamp, randomPart)
}
