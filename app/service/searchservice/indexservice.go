package searchservice

import (
	"fmt"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

// IndexBuildResult summarizes a Meilisearch rebuild.
type IndexBuildResult struct {
	ProcessedCount int    `json:"processedCount"`
	FailedCount    int    `json:"failedCount"`
	TotalBatches   int    `json:"totalBatches"`
	IndexName      string `json:"indexName"`
}

// convertToSearchDocument maps an article entity to a search document.
func convertToSearchDocument(article *articles.Entity) ArticleSearchDocument {
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
	if !meiliconnect.IsAvailable() {
		return nil, nil
	}

	client := meiliconnect.GetClient()
	indexName := Index
	index := client.Index(indexName)
	var task *meilisearch.TaskInfo
	var err error
	pk := "id"
	if article.ArticleStatus == 1 && article.ProcessStatus == 0 {
		doc := convertToSearchDocument(article)
		task, err = index.AddDocuments(doc, &pk)
		if err != nil {
			slog.Warn(fmt.Sprintf("Meilisearch 处理文章 ID:%v 失败: %v\n", doc.ID, err))
			return nil, nil
		}
		slog.Info(fmt.Sprintf("处理文章 ID:%v, TaskUID: %v\n", doc.ID, getTaskUID(task)))
	} else {
		_, err = index.Delete(cast.ToString(article.Id))
		if err != nil {
			slog.Warn(fmt.Sprintf("Meilisearch 删除文档失败: %v, Error: %v\n", article.Id, err))
			return nil, nil
		}
	}
	return task, nil
}

// BuildMeilisearchIndex rebuilds the Meilisearch article index.
func BuildMeilisearchIndex() (*IndexBuildResult, error) {
	if !meiliconnect.IsAvailable() {
		return nil, fmt.Errorf("Meilisearch 服务不可用，请检查配置或连接状态")
	}

	fmt.Println("开始构建 Meilisearch 文章索引...")

	client := meiliconnect.GetClient()
	indexName := Index
	index := client.Index(indexName)

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
		lo.ForEach(articleList, func(article *articles.Entity, _ int) {
			task, _ := BuildSingleArticleSearchDocument(article)
			fmt.Printf("处理文章 ID:%v, TaskUID: %v\n", article.Id, getTaskUID(task))
			processedCount++
		})
		articleStartId = articleList[len(articleList)-1].Id

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

// configureIndex applies searchable, filterable, sortable and displayed fields.
func configureIndex(index meilisearch.IndexManager) error {
	searchableAttributes := []string{
		"title",
		"searchContent",
	}
	_, err := index.UpdateSearchableAttributes(&searchableAttributes)
	if err != nil {
		return fmt.Errorf("设置可搜索字段失败: %v", err)
	}

	filterableAttributes := []any{
		"type",
		"userId",
		"category",
	}
	_, err = index.UpdateFilterableAttributes(&filterableAttributes)
	if err != nil {
		return fmt.Errorf("设置可过滤字段失败: %v", err)
	}

	sortableAttributes := []string{
		"createdAt",
		"updatedAt",
	}
	_, err = index.UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		return fmt.Errorf("设置可排序字段失败: %v", err)
	}

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

// getTaskUID returns nil when no task was created.
func getTaskUID(task *meilisearch.TaskInfo) any {
	if task == nil {
		return nil
	}
	return task.TaskUID
}
