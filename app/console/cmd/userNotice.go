package cmd

import (
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/eventNotification"
	"github.com/leancodebox/GooseForum/app/models/forum/kvstore"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "userNotice",
		Short: "",
		Run:   runUserNotice,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	// cmd.Flags().StringP("param", "p", "value", "--param=x | -p x")
	appendCommand(cmd)
}

func runUserNotice(cmd *cobra.Command, args []string) {
	fmt.Println("开始执行用户通知任务...")

	processedCount := 0
	notificationSentCount := 0
	batchSize := 50 // 每批处理50个用户
	lastUserId := uint64(0)

	// 分批处理用户
	for {
		// 获取一批用户
		userBatch := getUserBatch(lastUserId, batchSize)
		if len(userBatch) == 0 {
			break // 没有更多用户了
		}

		fmt.Printf("\n处理第 %d-%d 个用户...\n", processedCount+1, processedCount+len(userBatch))

		// 根据用户循环遍历
		for _, user := range userBatch {
			lastUserId = user.Id // 更新最后处理的用户ID
			processedCount++
			fmt.Printf("\n[%d] 处理用户: %s (ID: %d)\n", processedCount, user.Username, user.Id)

			// 判断 userStatistics 的最后活跃时间
			userStats := userStatistics.Get(user.Id)
			if !isUserActiveRecently(&userStats) {
				fmt.Printf("  ❌ 用户最近未活跃，跳过\n")
				continue
			}

			// 根据 eventNotification，判断用户是否存在7天内未读的通知
			unreadNotifications := getUnreadNotificationsWithin7Days(user.Id)
			if len(unreadNotifications) == 0 {
				fmt.Printf("  ❌ 用户无7天内未读通知，跳过\n")
				continue
			}

			// 判断 kvstore 记录的上次发送时间
			if !shouldSendNotification(user.Id) {
				fmt.Printf("  ❌ 距离上次发送时间不足24小时，跳过\n")
				continue
			}

			// 综合判断后合并需要提醒的内容，发送邮件通知
			emailContent := buildEmailContent(user, unreadNotifications)

			// 这里不用真发送。打印需要发送的内容即可
			printEmailContent(user, emailContent)

			// 更新发送记录
			updateLastSendTime(user.Id)

			notificationSentCount++
		}
	}

	fmt.Printf("\n=== 任务完成 ===\n")
	fmt.Printf("处理用户总数: %d\n", processedCount)
	fmt.Printf("发送通知数量: %d\n", notificationSentCount)
}

// 分批获取用户列表
func getUserBatch(startId uint64, limit int) []*users.EntityComplete {
	// 使用 QueryById 方法分批获取用户，避免一次性加载所有数据
	userList := users.QueryById(startId, limit)
	return userList
}

// 判断用户是否最近活跃（7天内）
func isUserActiveRecently(userStats *userStatistics.Entity) bool {
	// 检查最后活跃时间是否在7天内
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	return userStats.LastActiveTime.After(sevenDaysAgo)
}

// 获取用户7天内的未读通知
func getUnreadNotificationsWithin7Days(userId uint64) []*eventNotification.Entity {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	// 查询7天内创建且未读的通知
	notifications, err := eventNotification.QueryByUserId(userId, 100, 0, true)
	if err != nil {
		fmt.Printf("  ⚠️ 查询通知失败: %v\n", err)
		return nil
	}

	// 过滤出7天内的通知
	var recentNotifications []*eventNotification.Entity
	for _, notification := range notifications {
		if notification.CreatedAt.After(sevenDaysAgo) {
			recentNotifications = append(recentNotifications, notification)
		}
	}

	return recentNotifications
}

// 判断是否应该发送通知（检查上次发送时间）
func shouldSendNotification(userId uint64) bool {
	key := fmt.Sprintf("user_notice_last_send:%d", userId)

	lastSendTimeStr, err := kvstore.Get(key)
	if err != nil {
		fmt.Printf("  ⚠️ 查询上次发送时间失败: %v\n", err)
		return true // 如果查询失败，允许发送
	}

	if lastSendTimeStr == "" {
		return true // 从未发送过，允许发送
	}

	// 解析上次发送时间
	lastSendTime, err := time.Parse(time.RFC3339, lastSendTimeStr)
	if err != nil {
		fmt.Printf("  ⚠️ 解析上次发送时间失败: %v\n", err)
		return true
	}

	// 检查是否超过24小时
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	return lastSendTime.Before(twentyFourHoursAgo)
}

// 构建邮件内容
func buildEmailContent(user *users.EntityComplete, notifications []*eventNotification.Entity) map[string]interface{} {
	// 统计不同类型的通知数量
	notificationStats := make(map[string]int)
	for _, notification := range notifications {
		notificationStats[notification.EventType]++
	}

	// 构建邮件内容
	emailContent := map[string]interface{}{
		"to":       user.Email,
		"username": user.Username,
		"subject":  "GooseForum - 您有新的未读通知",
		"stats":    notificationStats,
		"total":    len(notifications),
		"details":  notifications[:min(len(notifications), 5)], // 最多显示5条详细通知
	}

	return emailContent
}

// 打印邮件内容（模拟发送）
func printEmailContent(user *users.EntityComplete, emailContent map[string]interface{}) {
	fmt.Printf("  ✅ 准备发送邮件通知\n")
	fmt.Printf("  📧 收件人: %s <%s>\n", emailContent["username"], emailContent["to"])
	fmt.Printf("  📋 主题: %s\n", emailContent["subject"])
	fmt.Printf("  📊 通知统计:\n")

	stats := emailContent["stats"].(map[string]int)
	for eventType, count := range stats {
		typeName := getEventTypeName(eventType)
		fmt.Printf("     - %s: %d条\n", typeName, count)
	}

	fmt.Printf("  📝 邮件内容预览:\n")
	fmt.Printf("     亲爱的 %s，\n", user.Username)
	fmt.Printf("     您在 GooseForum 有 %d 条未读通知，请及时查看。\n", emailContent["total"])
	fmt.Printf("     访问链接: https://forum.example.com/notifications\n")
	fmt.Printf("  ⏰ 发送时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

// 更新上次发送时间
func updateLastSendTime(userId uint64) {
	key := fmt.Sprintf("user_notice_last_send:%d", userId)
	currentTime := time.Now().Format(time.RFC3339)

	err := kvstore.Set(key, currentTime, 30*24*time.Hour) // 30天过期
	if err != nil {
		fmt.Printf("  ⚠️ 更新发送记录失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ 已更新发送记录\n")
	}
}

// 获取事件类型的中文名称
func getEventTypeName(eventType string) string {
	switch eventType {
	case eventNotification.EventTypeComment:
		return "评论通知"
	case eventNotification.EventTypeReply:
		return "回复通知"
	case eventNotification.EventTypeSystem:
		return "系统通知"
	case eventNotification.EventTypeFollow:
		return "关注通知"
	default:
		return "其他通知"
	}
}
