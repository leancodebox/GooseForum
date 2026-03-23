package cmd

import (
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "mockArticles",
		Short: "Mock articles for testing",
		Run:   runMockArticles,
	}
	appendCommand(cmd)
}

func runMockArticles(cmd *cobra.Command, args []string) {
	fmt.Println("Starting to mock articles...")

	allUsers := users.All()
	if len(allUsers) == 0 {
		fmt.Println("No users found, please create users first.")
		return
	}

	totalArticles := 50
	for i := 1; i <= totalArticles; i++ {
		// Pick a user randomly or sequentially
		user := allUsers[i%len(allUsers)]

		req := component.BetterRequest[controllers.WriteArticleReq]{
			UserId: user.Id,
			Params: controllers.WriteArticleReq{
				Title:      fmt.Sprintf("测试文章 %03d - %s", i, time.Now().Format("15:04:05")),
				Content:    fmt.Sprintf("这是第 %d 篇自动生成的测试文章内容。\n\n生成时间: %s\n作者: %s", i, time.Now().Format(time.RFC3339), user.Username),
				Type:       1,           // 0-3
				CategoryId: []uint64{1}, // Default to first category
			},
		}

		resp := controllers.WriteArticles(req)
		if resp.Data.Code == component.SUCCESS {
			fmt.Printf("[%d/%d] Created article: %s (User: %s)\n", i, totalArticles, req.Params.Title, user.Username)
		} else {
			fmt.Printf("[%d/%d] Failed to create article: %v\n", i, totalArticles, resp.Data.Msg)
		}

		// Small delay to ensure sequence and timestamp difference
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("Mock articles generation completed.")
}
