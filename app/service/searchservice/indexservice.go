package searchservice

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/meilisearchmodel"
	"github.com/meilisearch/meilisearch-go"
	"github.com/spf13/cast"
	"log/slog"
)

// ArticleSearchDocument 文章搜索文档结构
type ArticleSearchDocument meilisearchmodel.Doc

// IndexBuildResult 索引构建结果
type IndexBuildResult struct {
	ProcessedCount int    `json:"processedCount"`
	FailedCount    int    `json:"failedCount"`
	TotalBatches   int    `json:"totalBatches"`
	IndexName      string `json:"indexName"`
}

// convertToSearchDocument 转换文章实体为搜索文档
func convertToSearchDocument(article *articles.Entity) ArticleSearchDocument {
	// 提取优化的搜索内容
	searchContent := markdown2html.ExtractSearchContent(article.Content)
	categoryIds := article.CategoryId
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

func BuildSingleArticleSearchDocument(article *articles.Entity) (*meilisearch.TaskInfo, error) {
	// 获取 Meilisearch 客户端
	client := meiliconnect.GetClient()
	indexName := meilisearchmodel.Index
	index := client.Index(indexName)
	var task *meilisearch.TaskInfo
	var err error
	// 只索引已发布且正常状态的文章
	if article.ArticleStatus == 1 && article.ProcessStatus == 0 {
		doc := convertToSearchDocument(article)
		task, err = index.AddDocuments(doc, "id")
		slog.Info(fmt.Sprintf("处理文章 ID:%v, TaskUID: %v, Error: %v\n", doc.ID, getTaskUID(task), err))
	} else {
		// 删除不符合条件的文章
		_, err = index.Delete(cast.ToString(article.Id))
		if err != nil {
			slog.Info(fmt.Sprintf("删除文档失败: %v, Error: %v\n", article.Id, err))
		}
	}
	return task, err
}

// BuildMeilisearchIndex 构建Meilisearch索引
func BuildMeilisearchIndex() (*IndexBuildResult, error) {
	fmt.Println("开始构建 Meilisearch 文章索引...")

	// 获取 Meilisearch 客户端
	client := meiliconnect.GetClient()
	indexName := meilisearchmodel.Index
	index := client.Index(indexName)

	// 配置索引设置
	fmt.Println("配置索引设置...")
	if err := configureIndex(index); err != nil {
		return nil, fmt.Errorf("配置索引失败: %v", err)
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
		for _, article := range articleList {
			if articleStartId < article.Id {
				articleStartId = article.Id
			}
			task, err := BuildSingleArticleSearchDocument(article)
			fmt.Printf("处理文章 ID:%v, TaskUID: %v, Error: %v\n", article.Id, getTaskUID(task), err)
			if err != nil {
				fmt.Printf("更新文档失败: %v %v\n", article.Id, err)
				failedCount++
			} else {
				processedCount++
			}
		}

		totalBatches++
		if len(articleList) < limit {
			break
		}
	}

	result := &IndexBuildResult{
		ProcessedCount: processedCount,
		FailedCount:    failedCount,
		TotalBatches:   totalBatches,
		IndexName:      indexName,
	}

	fmt.Printf("\n=== Meilisearch 索引构建完成 ===\n")
	fmt.Printf("处理批次: %d\n", result.TotalBatches)
	fmt.Printf("成功索引: %d 篇文章\n", result.ProcessedCount)
	fmt.Printf("失败数量: %d 篇文章\n", result.FailedCount)
	fmt.Printf("索引名称: %s\n", result.IndexName)

	return result, nil
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
		"category",
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
	displayedAttributes := []string{"id", "title"}
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

// getTaskUID 安全获取TaskUID
func getTaskUID(task *meilisearch.TaskInfo) interface{} {
	if task == nil {
		return nil
	}
	return task.TaskUID
}
