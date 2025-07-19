package searchservice

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/leancodebox/GooseForum/app/models/meilisearchmodel"
	"github.com/meilisearch/meilisearch-go"
	"github.com/spf13/cast"
)

// SearchRequest 搜索请求结构
type SearchRequest struct {
	Query      string   `json:"query"`      // 搜索关键词
	Categories []uint64 `json:"categories"` // 分类ID列表
	Limit      int      `json:"limit"`      // 返回结果数量限制
	Offset     int      `json:"offset"`     // 偏移量
}

// SearchResult 搜索结果结构
type SearchResult struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
}

// SearchResponse 搜索响应结构
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int64          `json:"total"`
}

// SearchArticles 通过名字和类别搜索文章
// 直接从Meilisearch搜索结果中返回ID和标题，无需查询数据库
func SearchArticles(req SearchRequest) (*SearchResponse, error) {
	// 获取 Meilisearch 客户端
	client := meiliconnect.GetClient()
	index := client.Index(meilisearchmodel.Index)

	// 构建搜索请求
	searchReq := &meilisearch.SearchRequest{
		Query:  req.Query,
		Limit:  int64(req.Limit),
		Offset: int64(req.Offset),
	}

	// 如果指定了分类，添加过滤条件
	if len(req.Categories) > 0 {
		filters := make([]string, len(req.Categories))
		for i, categoryID := range req.Categories {
			filters[i] = fmt.Sprintf("category = %d", categoryID)
		}
		filterStr := fmt.Sprintf("(%s)", joinFilters(filters, " OR "))
		searchReq.Filter = filterStr
	}

	// 只返回需要的字段
	searchReq.AttributesToRetrieve = []string{"id", "title"}

	// 执行搜索
	searchResp, err := index.Search(req.Query, searchReq)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %v", err)
	}

	// 直接从搜索结果中提取ID和标题
	results := make([]SearchResult, 0, len(searchResp.Hits))
	for _, hit := range searchResp.Hits {
		if hitMap, ok := hit.(map[string]interface{}); ok {
			id := cast.ToUint64(hitMap["id"])
			title := cast.ToString(hitMap["title"])
			
			if id > 0 {
				results = append(results, SearchResult{
					ID:    id,
					Title: title,
				})
			}
		}
	}

	return &SearchResponse{
		Results: results,
		Total:   searchResp.EstimatedTotalHits,
	}, nil
}

// joinFilters 连接过滤条件
func joinFilters(filters []string, separator string) string {
	if len(filters) == 0 {
		return ""
	}
	if len(filters) == 1 {
		return filters[0]
	}

	result := filters[0]
	for i := 1; i < len(filters); i++ {
		result += separator + filters[i]
	}
	return result
}
