package cmd

import (
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/articlesUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "fixArticlesUserStat",
		Short: "Fix article user statistics and posters based on replies",
		Run:   runFixArticlesUserStat,
	}
	appendCommand(cmd)
}

func runFixArticlesUserStat(cmd *cobra.Command, args []string) {
	fmt.Println("Starting to fix article user stats and posters...")

	var lastId uint64 = 0
	batchSize := 100
	totalProcessed := 0

	for {
		batchArticles := articles.GetBatch(lastId, batchSize)
		if len(batchArticles) == 0 {
			break
		}

		fmt.Printf("Processing batch of %d articles, starting from ID %d\n", len(batchArticles), lastId+1)

		for _, article := range batchArticles {
			// 1. Get all replies for the article
			replies := reply.GetAllByArticleId(article.Id)

			// 2. Calculate stats in memory
			type UserStat struct {
				Count       uint32
				LastReplyAt time.Time
			}
			groupedReplies := lo.GroupBy(replies, func(r *reply.Entity) uint64 {
				return r.UserId
			})
			userStats := lo.MapValues(groupedReplies, func(userReplies []*reply.Entity, _ uint64) *UserStat {
				maxReply := lo.MaxBy(userReplies, func(a, b *reply.Entity) bool {
					return a.CreatedAt.After(b.CreatedAt)
				})
				return &UserStat{
					Count:       uint32(len(userReplies)),
					LastReplyAt: maxReply.CreatedAt,
				}
			})

			// 3. Update articlesUserStat for each user found in replies
			for userId, stat := range userStats {
				err := articlesUserStat.FixStat(article.Id, userId, stat.Count, stat.LastReplyAt)
				if err != nil {
					fmt.Printf("Error fixing stat for article %d user %d: %v\n", article.Id, userId, err)
				}
			}

			// 4. Update Posters in Articles table
			// Get top users from the updated stats
			topUserIds := articlesUserStat.SyncArticlePosters(article.Id)

			// Filter out author to avoid duplication (logic consistent with articleController)
			filteredList := lo.Filter(topUserIds, func(t uint64, _ int) bool {
				return t != article.UserId
			})

			// Prepend Author to be always the first
			finalList := append([]uint64{article.UserId}, filteredList...)

			// Map to Poster struct
			posters := lo.Map(finalList, func(t uint64, _ int) articles.Poster {
				return articles.Poster{
					UserID: t,
				}
			})

			// Save updated posters to article
			err := articles.UpdatePosters(article.Id, posters)
			if err != nil {
				fmt.Printf("Error updating posters for article %d: %v\n", article.Id, err)
			}

			lastId = article.Id
			totalProcessed++
		}
		fmt.Printf("Processed up to article ID: %d\n", lastId)
	}

	fmt.Printf("Success: All articles processed (%d total).\n", totalProcessed)
}
