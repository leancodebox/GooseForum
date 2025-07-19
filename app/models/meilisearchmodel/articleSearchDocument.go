package meilisearchmodel

type Doc struct {
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

const Index = "articles"
