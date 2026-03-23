package cmd

import (
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "fixDailyStats",
		Short: "修复每日统计数据 (daily_stats)",
		Run:   runFixDailyStats,
	}
	appendCommand(cmd)
}

func runFixDailyStats(cmd *cobra.Command, args []string) {
	fmt.Println("开始修复每日统计数据...")

	minDate := time.Now()

	// 1. 修复注册人数
	fmt.Println("同步注册人数统计...")
	regResults, _ := users.GetCountGroupByDay()
	for _, res := range regResults {
		date := cast.ToTime(res["date"])
		if date.Before(minDate) {
			minDate = date
		}
		count := cast.ToInt64(res["count"])
		_ = dailyStats.SetValue(date, dailyStats.StatTypeRegCount, count)
	}

	// 2. 修复发帖数
	fmt.Println("同步发帖数统计...")
	articleResults, _ := articles.GetCountGroupByDay()
	for _, res := range articleResults {
		date := cast.ToTime(res["date"])
		if date.Before(minDate) {
			minDate = date
		}
		count := cast.ToInt64(res["count"])
		_ = dailyStats.SetValue(date, dailyStats.StatTypeArticleCount, count)
	}

	// 3. 修复回复数
	fmt.Println("同步回复数统计...")
	replyResults, _ := reply.GetCountGroupByDay()
	for _, res := range replyResults {
		date := cast.ToTime(res["date"])
		if date.Before(minDate) {
			minDate = date
		}
		count := cast.ToInt64(res["count"])
		_ = dailyStats.SetValue(date, dailyStats.StatTypeReplyCount, count)
	}

	// 4. 补充缺失的日期记录（从最小日期至今）
	// 如果 minDate 是今天，也补充一下
	fmt.Printf("补充从 %s 至今的缺失统计项...\n", minDate.Format("2006-01-02"))
	now := time.Now()
	keys := []dailyStats.StatType{
		dailyStats.StatTypeRegCount,
		dailyStats.StatTypeArticleCount,
		dailyStats.StatTypeReplyCount,
	}

	// 确保至少包含过去 7 天，以防数据库里什么都没有
	startDate := minDate
	sevenDaysAgo := now.AddDate(0, 0, -7)
	if startDate.After(sevenDaysAgo) {
		startDate = sevenDaysAgo
	}

	for d := startDate; !d.After(now); d = d.AddDate(0, 0, 1) {
		for _, key := range keys {
			_ = dailyStats.InitStats(d, key)
		}
	}

	fmt.Println("每日统计数据修复完成！")
}
