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
		Use:   "mock-replies",
		Short: "Mock replies for articles",
		Run:   runMockArticlesUserStat,
	}
	cmd.Flags().Uint64("article-id", 0, "Only mock replies for the specified article")
	cmd.Flags().Int("count", 0, "Reply count for --article-id mode")
	cmd.Flags().Int("articles", 1000, "How many latest articles to scan when --article-id is empty")
	cmd.Flags().Int("max-per-article", 5, "Random upper bound per article when --article-id is empty")
	appendCommand(cmd)
}

func runMockArticlesUserStat(cmd *cobra.Command, args []string) {

	fmt.Println("Starting to mock replies...")
	articleID, _ := cmd.Flags().GetUint64("article-id")
	count, _ := cmd.Flags().GetInt("count")
	articleLimit, _ := cmd.Flags().GetInt("articles")
	maxPerArticle, _ := cmd.Flags().GetInt("max-per-article")

	if articleLimit <= 0 {
		articleLimit = 1000
	}
	if maxPerArticle <= 0 {
		maxPerArticle = 5
	}

	// 1. Get last 1000 articles
	articleList := articles.GetLast(articleLimit)
	if articleID > 0 {
		article := articles.Get(articleID)
		if article.Id == 0 {
			fmt.Printf("Article %d not found.\n", articleID)
			return
		}
		articleList = []*articles.Entity{&article}
		if count <= 0 {
			count = 60
		}
	}
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
		replyCount := count
		if articleID == 0 {
			replyCount = rand.Intn(maxPerArticle)
		}
		if replyCount == 0 {
			continue
		}

		for j := range replyCount {
			// Pick random user
			userId := lo.Sample(userIds)

			req := component.BetterRequest[controllers.ArticleReplyId]{
				Params: controllers.ArticleReplyId{
					ArticleId: article.Id,
					Content:   fmt.Sprintf("Mock reply content %d-%d for article %d. This is generated for local reply timeline testing.", i, j, article.Id),
					ReplyId:   0,
				},
				UserId: userId,
			}
			resp := controllers.ArticleReply(req)

			if resp.Data.Code != component.SUCCESS {
				fmt.Printf("Error creating reply for article %d: %s\n", article.Id, resp.Data.MessageCode)
			} else {
				totalReplies++
			}
		}
	}

	fmt.Printf("Success: Created %d mock replies for %d articles.\n", totalReplies, len(articleList))
}
