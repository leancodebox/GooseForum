package vo

// ArticlesSimpleVo is the compact article payload used by list views and feed-like responses.
type ArticlesSimpleVo struct {
	Id             uint64     `json:"id"`
	Title          string     `json:"title,omitempty"`
	Description    string     `json:"description,omitempty"`
	Content        string     `json:"content,omitempty"`
	CreateTime     string     `json:"createTime,omitempty"`
	LastUpdateTime string     `json:"lastUpdateTime,omitempty"`
	Username       string     `json:"username,omitempty"`
	AuthorId       uint64     `json:"authorId,omitempty"`
	ViewCount      uint64     `json:"viewCount"`
	CommentCount   uint64     `json:"commentCount"`
	Type           int8       `json:"type,omitempty"`
	TypeStr        string     `json:"typeStr,omitempty"`
	Categories     []string   `json:"categories,omitempty"`
	CategoriesId   []uint64   `json:"categoriesId,omitempty"`
	AvatarUrl      string     `json:"avatarUrl,omitempty"`
	Posters        []PosterVo `json:"posters,omitempty"`
}

// PosterVo is a lightweight user summary attached to compact article responses.
type PosterVo struct {
	Id        uint64 `json:"id"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatarUrl"`
}
