package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
)

// isAllowedByDatabase 基于数据库记录检查用户是否允许上传
func isAllowedByDatabase(userId uint64, maxCount int, window time.Duration) bool {
	// 根据时间窗口计算查询范围
	now := time.Now()
	var startTime time.Time
	
	if window == 24*time.Hour {
		// 如果是24小时窗口，使用今日统计方法
		count := filedata.CountUserUploadsToday(userId)
		return count < int64(maxCount)
	} else {
		// 其他时间窗口，使用通用方法
		startTime = now.Add(-window)
		count := filedata.CountUserUploadsInTimeRange(userId, startTime, now)
		return count < int64(maxCount)
	}
}

// getUserUploadCount 获取用户当前上传计数
func getUserUploadCount(userId uint64, window time.Duration) int64 {
	if window == 24*time.Hour {
		return filedata.CountUserUploadsToday(userId)
	} else {
		now := time.Now()
		startTime := now.Add(-window)
		return filedata.CountUserUploadsInTimeRange(userId, startTime, now)
	}
}

// FileUploadRateLimit 文件上传频率限制中间件
// maxUploads: 最大上传次数
// window: 时间窗口
func FileUploadRateLimit(maxUploads int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userIdData, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, component.FailData("未登录"))
			c.Abort()
			return
		}
		
		userId := cast.ToUint64(userIdData)
		if userId == 0 {
			c.JSON(http.StatusUnauthorized, component.FailData("无效的用户ID"))
			c.Abort()
			return
		}
		
		// 获取当前上传次数
		currentCount := getUserUploadCount(userId, window)
		
		// 检查频率限制
		if currentCount >= int64(maxUploads) {
			var timeDesc string
			if window == 24*time.Hour {
				timeDesc = "今日"
			} else {
				timeDesc = fmt.Sprintf("最近%v", window)
			}
			
			msg := fmt.Sprintf("上传频率超限，%s已上传%d张图片，最多允许%d张", timeDesc, currentCount, maxUploads)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  msg,
				"data": gin.H{
					"current_count": currentCount,
					"max_count":     maxUploads,
					"remaining":     maxUploads - int(currentCount),
					"window":        window.String(),
				},
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// GetUserUploadCount 获取用户当前上传计数（用于调试）
func GetUserUploadCount(userId uint64, window time.Duration) int64 {
	return getUserUploadCount(userId, window)
}