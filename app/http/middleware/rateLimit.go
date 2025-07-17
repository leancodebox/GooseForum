package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

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

// checkUserRegistrationTime 检查用户注册时间是否超过指定天数
func checkUserRegistrationTime(userId uint64, minDays int) (bool, error) {
	user, err := users.Get(userId)
	if err != nil {
		return false, err
	}
	if user.Status == 1 {
		return false, nil
	}
	// 计算用户注册时间到现在的天数
	now := time.Now()
	registrationTime := user.CreatedAt
	daysSinceRegistration := int(now.Sub(registrationTime).Hours() / 24)

	return daysSinceRegistration >= minDays, nil
}

// FileUploadRateLimit 文件上传频率限制中间件
// maxUploads: 最大上传次数
// window: 时间窗口
func FileUploadRateLimit(maxUploads int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint64("userId")
		if userId == 0 {
			c.JSON(http.StatusUnauthorized, component.FailData("未登陆"))
			c.Abort()
			return
		}
		// 检查用户注册时间是否超过3天
		isEligible, err := checkUserRegistrationTime(userId, 3)
		if err != nil {
			c.JSON(http.StatusInternalServerError, component.FailData("获取用户信息失败"))
			c.Abort()
			return
		}

		if !isEligible {
			c.JSON(http.StatusForbidden, component.FailData("注册未满3天的用户暂时无法上传图片"))
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
			c.JSON(http.StatusTooManyRequests, component.FailData(msg))
			c.Abort()
			return
		}

		c.Next()
	}
}
