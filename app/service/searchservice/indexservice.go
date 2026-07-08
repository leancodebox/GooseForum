package searchservice

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
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

// convertTopicToSearchDocument maps a topic and its first post to a search document.
func convertTopicToSearchDocument(topic *topics.Entity, firstPost *posts.Entity) TopicSearchDocument {
	searchContent := ""
	if firstPost != nil {
		searchContent = markdown2html.ExtractSearchContent(firstPost.Content)
	}
	return TopicSearchDocument{
		ID:            topic.Id,
		Title:         topic.Title,
		SearchContent: searchContent,
		Category:      topic.CategoryIds,
		TopicStatus:   topic.Status,
		ProcessStatus: topic.ProcessStatus,
		CreatedAt:     topic.CreatedAt.Unix(),
		UpdatedAt:     topic.UpdatedAt.Unix(),
	}
}

func BuildSingleTopicSearchDocument(topic *topics.Entity, firstPost *posts.Entity) (*meilisearch.TaskInfo, error) {
	if !meiliconnect.IsAvailable() {
		return nil, nil
	}
	if topic == nil {
		return nil, nil
	}

	client := meiliconnect.GetClient()
	index := client.Index(TopicIndex)
	var task *meilisearch.TaskInfo
	var err error
	pk := "id"
	if topic.Status == 1 && topic.ProcessStatus == 0 {
		doc := convertTopicToSearchDocument(topic, firstPost)
		task, err = index.AddDocuments(doc, &pk)
		if err != nil {
			slog.Warn(fmt.Sprintf("Meilisearch 处理主题 ID:%v 失败: %v\n", doc.ID, err))
			return nil, fmt.Errorf("add search document: %w", err)
		}
		slog.Info(fmt.Sprintf("处理主题 ID:%v, TaskUID: %v\n", doc.ID, getTaskUID(task)))
	} else {
		_, err = index.Delete(cast.ToString(topic.Id))
		if err != nil {
			slog.Warn(fmt.Sprintf("Meilisearch 删除文档失败: %v, Error: %v\n", topic.Id, err))
			return nil, fmt.Errorf("delete search document: %w", err)
		}
	}
	return task, nil
}

// BuildMeilisearchIndex rebuilds the Meilisearch topic index.
func BuildMeilisearchIndex() (*IndexBuildResult, error) {
	if !meiliconnect.IsAvailable() {
		return nil, errors.New("meilisearch 服务不可用，请检查配置或连接状态")
	}

	fmt.Println("开始构建 Meilisearch 主题索引...")

	client := meiliconnect.GetClient()
	indexName := TopicIndex
	index := client.Index(indexName)

	fmt.Println("配置索引设置...")
	if err := configureIndex(index); err != nil {
		return nil, fmt.Errorf("配置索引失败: %w", err)
	}

	var topicStartID uint64
	limit := 100
	processedCount := 0
	failedCount := 0
	totalBatches := 0

	for {
		topicList := topics.QueryById(topicStartID, limit)
		if len(topicList) == 0 {
			break
		}
		lo.ForEach(topicList, func(topic *topics.Entity, _ int) {
			firstPost := posts.Get(topic.FirstPostId)
			if firstPost.Id == 0 {
				firstPost, _ = posts.GetByTopicPostNoAtOrAfter(topic.Id, 1)
			}
			task, err := BuildSingleTopicSearchDocument(topic, &firstPost)
			if err != nil {
				failedCount++
				slog.Warn("failed to build topic search document", "topicId", topic.Id, "err", err)
				return
			}
			fmt.Printf("处理主题 ID:%v, TaskUID: %v\n", topic.Id, getTaskUID(task))
			processedCount++
		})
		topicStartID = topicList[len(topicList)-1].Id

		totalBatches++
		if len(topicList) < limit {
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
	fmt.Printf("成功索引: %d 个主题\n", result.ProcessedCount)
	fmt.Printf("失败数量: %d 个主题\n", result.FailedCount)
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
		return fmt.Errorf("设置可搜索字段失败: %w", err)
	}

	filterableAttributes := []any{
		"category",
	}
	_, err = index.UpdateFilterableAttributes(&filterableAttributes)
	if err != nil {
		return fmt.Errorf("设置可过滤字段失败: %w", err)
	}

	sortableAttributes := []string{
		"createdAt",
		"updatedAt",
	}
	_, err = index.UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		return fmt.Errorf("设置可排序字段失败: %w", err)
	}

	displayedAttributes := []string{"id", "title"}
	_, err = index.UpdateDisplayedAttributes(&displayedAttributes)
	if err != nil {
		return fmt.Errorf("设置显示字段失败: %w", err)
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
