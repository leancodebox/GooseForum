package cmd

import (
	"fmt"
	"math/rand"

	"github.com/leancodebox/GooseForum/app/http/controllers"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "mockArticlesUserStat",
		Short: "Mock replies for the last 1000 articles",
		Run:   runMockArticlesUserStat,
	}
	appendCommand(cmd)
}

func runMockArticlesUserStat(cmd *cobra.Command, args []string) {

	fmt.Println("Starting to mock replies...")

	// 1. Get last 1000 articles
	articleList := articles.GetLast(1000)
	fmt.Printf("Found %d articles\n", len(articleList))

	if len(articleList) == 0 {
		fmt.Println("No articles found.")
		return
	}

	// 2. Get all users
	userList := users.All()
	userIds := lo.Map(userList, func(u *users.EntityComplete, _ int) uint64 {
		return u.Id
	})

	if len(userIds) == 0 {
		fmt.Println("No users found.")
		return
	}
	fmt.Printf("Found %d users\n", len(userIds))

	// 3. Loop articles and generate replies
	totalReplies := 0
	for i, article := range articleList {
		// Random number of replies: 0 to 4
		replyCount := rand.Intn(5)
		if replyCount == 0 {
			continue
		}

		for j := range replyCount {
			// Pick random user
			userId := lo.Sample(userIds)

			req := component.BetterRequest[controllers.ArticleReplyId]{
				Params: controllers.ArticleReplyId{
					ArticleId: article.Id,
					Content:   fmt.Sprintf("Mock reply content %d-%d for article %d", i, j, article.Id),
					ReplyId:   0,
				},
				UserId: userId,
			}
			resp := controllers.ArticleReply(req)

			if resp.Code != 200 { // Assuming 200 is success code in component.Status or http.StatusOK
				fmt.Printf("Error creating reply for article %d: %v\n", article.Id, resp.Data.Msg)
			} else {
				totalReplies++
			}
		}
	}

	fmt.Printf("Success: Created %d mock replies for %d articles.\n", totalReplies, len(articleList))
}
