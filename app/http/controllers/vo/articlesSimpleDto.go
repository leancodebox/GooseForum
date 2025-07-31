package vo

type ArticlesSimpleDto struct {
	Id             uint64   `json:"id"`
	Title          string   `json:"title,omitempty"`
	Content        string   `json:"content,omitempty"`
	CreateTime     string   `json:"createTime,omitempty"`
	LastUpdateTime string   `json:"lastUpdateTime,omitempty"`
	Username       string   `json:"username,omitempty"`
	AuthorId       uint64   `json:"authorId,omitempty"`
	ViewCount      uint64   `json:"viewCount,omitempty"`
	CommentCount   uint64   `json:"commentCount"`
	Type           int8     `json:"type,omitempty"`
	TypeStr        string   `json:"typeStr,omitempty"`
	Category       string   `json:"category,omitempty"`
	Categories     []string `json:"categories,omitempty"`
	CategoriesId   []uint64 `json:"categoriesId,omitempty"`
	AvatarUrl      string   `json:"avatarUrl,omitempty"`
}
