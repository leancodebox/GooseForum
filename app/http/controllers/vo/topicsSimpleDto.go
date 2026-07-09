package vo

// TopicsSimpleVo is the compact topic payload used by list views and feed-like responses.
type TopicsSimpleVo struct {
	Id             uint64     `json:"id"`
	Title          string     `json:"title,omitempty"`
	Description    string     `json:"description,omitempty"`
	FirstImageURL  string     `json:"firstImageUrl,omitempty"`
	Content        string     `json:"content,omitempty"`
	CreateTime     string     `json:"createTime,omitempty"`
	LastUpdateTime string     `json:"lastUpdateTime,omitempty"`
	Username       string     `json:"username,omitempty"`
	AuthorId       uint64     `json:"authorId,omitempty"`
	ViewCount      uint64     `json:"viewCount"`
	CommentCount   uint64     `json:"commentCount"`
	PinWeight      int        `json:"pinWeight"`
	ProcessStatus  int8       `json:"processStatus"`
	Categories     []string   `json:"categories,omitempty"`
	CategoriesId   []uint64   `json:"categoriesId,omitempty"`
	AvatarUrl      string     `json:"avatarUrl,omitempty"`
	Posters        []PosterVo `json:"posters,omitempty"`
}

// PosterVo is a lightweight user summary attached to compact topic responses.
type PosterVo struct {
	Id        uint64 `json:"id"`
	Username  string `json:"username"`
	AvatarUrl string `json:"avatarUrl"`
}
