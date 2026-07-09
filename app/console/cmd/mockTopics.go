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
		Use:   "mock-topics",
		Short: "Mock topics for testing",
		Run:   runMockTopics,
	}
	appendCommand(cmd)
}

func runMockTopics(cmd *cobra.Command, args []string) {
	fmt.Println("Starting to mock topics...")

	allUsers := users.All()
	if len(allUsers) == 0 {
		fmt.Println("No users found, please create users first.")
		return
	}

	totalTopics := 50
	for i := 1; i <= totalTopics; i++ {
		// Pick a user randomly or sequentially
		user := allUsers[i%len(allUsers)]

		req := component.BetterRequest[controllers.WriteTopicReq]{
			UserId: user.Id,
			Params: controllers.WriteTopicReq{
				Title:      fmt.Sprintf("测试文章 %03d - %s", i, time.Now().Format("15:04:05")),
				Content:    fmt.Sprintf("这是第 %d 篇自动生成的测试文章内容。\n\n生成时间: %s\n作者: %s", i, time.Now().Format(time.RFC3339), user.Username),
				CategoryId: []uint64{1}, // Default to first category
			},
		}

		resp := controllers.WriteTopic(req)
		if resp.Data.Code == component.SUCCESS {
			fmt.Printf("[%d/%d] Created topic: %s (User: %s)\n", i, totalTopics, req.Params.Title, user.Username)
		} else {
			fmt.Printf("[%d/%d] Failed to create topic: %s\n", i, totalTopics, resp.Data.MessageCode)
		}

		// Small delay to ensure sequence and timestamp difference
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("Mock topics generation completed.")
}
