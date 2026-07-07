package searchservice

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
)

// SearchRequest is a topic search request.
type SearchRequest struct {
	Query      string   `json:"query"`
	Categories []uint64 `json:"categories"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

// SearchResult is one search hit.
type SearchResult struct {
	ID    uint64 `json:"id"`
	Title string `json:"title"`
}

// SearchResponse is the topic search response.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int64          `json:"total"`
}

// SearchTopics returns topic IDs and titles directly from Meilisearch.
func SearchTopics(req SearchRequest) (*SearchResponse, error) {
	if !meiliconnect.IsAvailable() {
		return &SearchResponse{
			Results: []SearchResult{},
			Total:   0,
		}, nil
	}

	client := meiliconnect.GetClient()
	index := client.Index(TopicIndex)

	searchReq := &meilisearch.SearchRequest{
		Query:  req.Query,
		Limit:  int64(req.Limit),
		Offset: int64(req.Offset),
	}

	if len(req.Categories) > 0 {
		filters := lo.Map(req.Categories, func(categoryID uint64, _ int) string {
			return fmt.Sprintf("category = %d", categoryID)
		})
		filterStr := fmt.Sprintf("(%s)", strings.Join(filters, " OR "))
		searchReq.Filter = filterStr
	}

	searchReq.AttributesToRetrieve = []string{"id", "title"}

	searchResp, err := index.Search(req.Query, searchReq)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	results := lo.FilterMap(searchResp.Hits, func(hit meilisearch.Hit, _ int) (SearchResult, bool) {
		itemResult := SearchResult{}
		if err := hit.Decode(&itemResult); err != nil {
			slog.Error("failed to decode search hit", "err", err)
			return SearchResult{}, false
		}
		return itemResult, itemResult.ID > 0
	})

	return &SearchResponse{
		Results: results,
		Total:   searchResp.EstimatedTotalHits,
	}, nil
}
