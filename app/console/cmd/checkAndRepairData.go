package cmd

import (
	"fmt"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCollection"
	"github.com/leancodebox/GooseForum/app/models/forum/articleLike"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/reply"
	"github.com/leancodebox/GooseForum/app/models/forum/userFollow"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/meilisearch/meilisearch-go"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	cmd := &cobra.Command{
		Use:   "checkAndRepairData",
		Short: "检查和修复数据",
		Run:   runCheckAndRepairData,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	cmd.Flags().BoolP("meilisearch", "m", false, "构建 Meilisearch 文章索引")
	cmd.Flags().BoolP("repair", "r", false, "修复用户和文章数据")
	appendCommand(cmd)
}

func runCheckAndRepairData(cmd *cobra.Command, args []string) {
	meilisearchFlag, _ := cmd.Flags().GetBool("meilisearch")
	repairFlag, _ := cmd.Flags().GetBool("repair")

	// 如果没有指定任何标志，默认执行修复操作
	if !meilisearchFlag && !repairFlag {
		repairFlag = true
	}

	if meilisearchFlag {
		fmt.Println("=== 开始构建 Meilisearch 索引 ===")
		buildMeilisearch()
		fmt.Println("=== Meilisearch 索引构建完成 ===")
	}

	if repairFlag {
		fmt.Println("=== 开始数据修复 ===")
		repairUserData()
		repairArticleDescriptions()
		fmt.Println("=== 数据修复完成 ===")
	}
}

// repairUserData 修复用户数据
func repairUserData() {
	fmt.Println("检查用户")
	var userStartId uint64 = 0
	limit := 333
	for {
		userList := users.QueryById(userStartId, limit)
		for _, userItem := range userList {
			if userStartId < userItem.Id {
				userStartId = userItem.Id
			}
			if userItem.AvatarUrl == "" {
				userItem.AvatarUrl = users.RandAvatarUrl()
				users.Save(userItem)
			}
			userSt := userStatistics.Get(userItem.Id)
			if userSt.UserId == 0 {
				fmt.Println("用户统计信息不存在，开始初始化")
				userSt.UserId = userItem.Id
				userSt.LastActiveTime = userItem.UpdatedAt
			}
			userSt.ArticleCount = cast.ToUint(articles.GetUserCount(userItem.Id))
			fmt.Println("获取文章总量", userSt.ArticleCount)
			userSt.ReplyCount = cast.ToUint(reply.GetUserCount(userItem.Id))
			fmt.Println("获取评论总量", userSt.ReplyCount)
			userSt.LikeReceivedCount = cast.ToUint(articleLike.GetLikeReceivedCount(userItem.Id))
			fmt.Println("获取接收到的点赞", userSt.LikeReceivedCount)
			userSt.LikeGivenCount = cast.ToUint(articleLike.GetLikeGivenCount(userItem.Id))
			fmt.Println("获送出点赞数量", userSt.LikeGivenCount)
			userSt.FollowingCount = cast.ToUint(userFollow.GetFollowingCount(userItem.Id)) //关注数
			fmt.Println("关注数", userSt.LikeGivenCount)
			userSt.FollowerCount = cast.ToUint(userFollow.GetFollowerCount(userItem.Id)) //粉丝数
			fmt.Println("粉丝数", userSt.LikeGivenCount)
			userSt.CollectionCount = cast.ToUint(articleCollection.GetCollectionCount(userItem.Id))
			fmt.Println("获取收藏数量", userSt.CollectionCount)
			userStatistics.SaveOrCreateById(&userSt)
			fmt.Println(userSt.UserId, "保存成功")
		}
		if len(userList) < limit {
			break
		}
	}
	fmt.Println("用户数据计算完毕")
}

// repairArticleDescriptions 修复所有文章的描述
func repairArticleDescriptions() {
	var articleStartId uint64 = 0
	limit := 100
	updatedCount := 0

	for {
		articleList := articles.QueryById(articleStartId, limit)
		for _, article := range articleList {
			if articleStartId < article.Id {
				articleStartId = article.Id
			}
			article.LikeCount = cast.ToUint64(articleLike.GetArticleLikeByArticleId(article.Id))

			// 如果描述为空或者很短，重新生成
			if article.Description == "" || len(strings.TrimSpace(article.Description)) < 10 {
				newDescription := markdown2html.ExtractDescription(article.Content, 200)
				if newDescription != "" && newDescription != article.Description {
					article.Description = newDescription
					err := articles.Save(article)
					if err != nil {
						fmt.Printf("更新文章 %d 描述失败: %v\n", article.Id, err)
					} else {
						updatedCount++
						fmt.Printf("已更新文章 %d 的描述: %s\n", article.Id, newDescription[:min(50, len(newDescription))])
					}
				}
			}
		}
		if len(articleList) < limit {
			break
		}
	}

	fmt.Printf("共更新了 %d 篇文章的描述\n", updatedCount)
}

// ArticleSearchDocument 文章搜索文档结构
type ArticleSearchDocument struct {
	ID            uint64   `json:"id"`
	Title         string   `json:"title"`         // 主要搜索字段
	SearchContent string   `json:"searchContent"` // 优化后的搜索文本
	Type          int8     `json:"type"`          // 可过滤字段
	ArticleStatus int8     `json:"articleStatus"` // 可过滤字段
	ProcessStatus int8     `json:"processStatus"` // 可过滤字段
	Category      []uint64 `json:"category"`
	CreatedAt     int64    `json:"createdAt"` // 时间戳(Unix)
	UpdatedAt     int64    `json:"updatedAt"` // 时间戳(Unix)
}

// convertToSearchDocument 转换文章实体为搜索文档
func convertToSearchDocument(article *articles.Entity) ArticleSearchDocument {
	// 提取优化的搜索内容
	searchContent := markdown2html.ExtractSearchContent(article.Content)
	categoryIds := array.Map(articleCategoryRs.GetByArticleIdsEffective([]uint64{article.Id}), func(rs *articleCategoryRs.Entity) uint64 {
		return cast.ToUint64(rs.Id)
	})
	return ArticleSearchDocument{
		ID:            article.Id,
		Title:         article.Title,
		SearchContent: searchContent,
		Type:          article.Type,
		Category:      categoryIds,
		ArticleStatus: article.ArticleStatus,
		ProcessStatus: article.ProcessStatus,
		CreatedAt:     article.CreatedAt.Unix(),
		UpdatedAt:     article.UpdatedAt.Unix(),
	}
}

func buildMeilisearch() {
	fmt.Println("开始构建 Meilisearch 文章索引...")

	// 获取 Meilisearch 客户端
	client := meiliconnect.GetClient()
	indexName := "articles"
	index := client.Index(indexName)

	// 配置索引设置
	fmt.Println("配置索引设置...")
	err := configureIndex(index)
	if err != nil {
		fmt.Printf("配置索引失败: %v\n", err)
		return
	}

	var articleStartId uint64 = 0
	limit := 100
	processedCount := 0
	failedCount := 0
	totalBatches := 0

	for {
		articleList := articles.QueryById(articleStartId, limit)
		if len(articleList) == 0 {
			break
		}

		// 转换为搜索文档
		var documents []ArticleSearchDocument
		for _, article := range articleList {
			if articleStartId < article.Id {
				articleStartId = article.Id
			}

			// 只索引已发布且正常状态的文章
			if article.ArticleStatus == 1 && article.ProcessStatus == 0 {
				doc := convertToSearchDocument(article)
				task, err := index.AddDocuments(doc, "id")
				fmt.Println("task", task, err)
				if err != nil {
					fmt.Printf("批次 %d 添加文档失败: %v\n", totalBatches+1, err)
					failedCount += len(documents)
				} else {
					fmt.Printf("批次 %d: 成功添加 %d 篇文章到索引 (TaskUID: %d)\n",
						totalBatches+1, len(documents), task.TaskUID)
					processedCount += len(documents)
				}
			}
		}

		totalBatches++

		if len(articleList) < limit {
			break
		}
	}

	fmt.Printf("\n=== Meilisearch 索引构建完成 ===\n")
	fmt.Printf("处理批次: %d\n", totalBatches)
	fmt.Printf("成功索引: %d 篇文章\n", processedCount)
	fmt.Printf("失败数量: %d 篇文章\n", failedCount)
	fmt.Printf("索引名称: %s\n", indexName)
}

// configureIndex 配置 Meilisearch 索引设置
func configureIndex(index meilisearch.IndexManager) error {
	// 设置可搜索字段（按权重排序）
	searchableAttributes := []string{
		"title",         // 权重最高
		"searchContent", // 优化后的搜索内容
	}
	_, err := index.UpdateSearchableAttributes(&searchableAttributes)
	if err != nil {
		return fmt.Errorf("设置可搜索字段失败: %v", err)
	}

	// 设置可过滤字段
	filterableAttributes := []string{
		"type",
		"userId",
		"articleStatus",
		"processStatus",
	}
	_, err = index.UpdateFilterableAttributes(&filterableAttributes)
	if err != nil {
		return fmt.Errorf("设置可过滤字段失败: %v", err)
	}

	// 设置可排序字段
	sortableAttributes := []string{
		"createdAt",
		"updatedAt",
	}
	_, err = index.UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		return fmt.Errorf("设置可排序字段失败: %v", err)
	}

	// 设置显示字段（返回所有字段）
	displayedAttributes := []string{"*"}
	_, err = index.UpdateDisplayedAttributes(&displayedAttributes)
	if err != nil {
		return fmt.Errorf("设置显示字段失败: %v", err)
	}

	fmt.Println("索引配置完成:")
	fmt.Printf("- 可搜索字段: %v\n", searchableAttributes)
	fmt.Printf("- 可过滤字段: %v\n", filterableAttributes)
	fmt.Printf("- 可排序字段: %v\n", sortableAttributes)

	return nil
}
