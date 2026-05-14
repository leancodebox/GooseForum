package searchservice

import (
	"fmt"
	"strings"

	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/meilisearch/meilisearch-go"
	"github.com/samber/lo"
)

// SearchRequest is an article search request.
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

// SearchResponse is the article search response.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int64          `json:"total"`
}

// SearchArticles returns article IDs and titles directly from Meilisearch.
func SearchArticles(req SearchRequest) (*SearchResponse, error) {
	if !meiliconnect.IsAvailable() {
		return &SearchResponse{
			Results: []SearchResult{},
			Total:   0,
		}, nil
	}

	client := meiliconnect.GetClient()
	index := client.Index(Index)

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
		return nil, fmt.Errorf("搜索失败: %v", err)
	}

	results := lo.FilterMap(searchResp.Hits, func(hit meilisearch.Hit, _ int) (SearchResult, bool) {
		itemResult := SearchResult{}
		hit.Decode(&itemResult)
		return itemResult, itemResult.ID > 0
	})

	return &SearchResponse{
		Results: results,
		Total:   searchResp.EstimatedTotalHits,
	}, nil
}
