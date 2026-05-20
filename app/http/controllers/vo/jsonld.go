package vo

// Person represents a schema.org Person node in JSON-LD output.
type Person struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Organization represents a schema.org Organization node in JSON-LD output.
type Organization struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Comment represents a schema.org Comment node attached to a forum post.
type Comment struct {
	Type          string `json:"@type"`
	Text          string `json:"text"`
	Author        Person `json:"author"`
	DatePublished string `json:"datePublished,omitempty"`
	URL           string `json:"url,omitempty"`
}

// InteractionCounter represents schema.org interaction statistics for an article.
type InteractionCounter struct {
	Type                 string `json:"@type"`
	InteractionType      string `json:"interactionType"`
	UserInteractionCount uint64 `json:"userInteractionCount"`
}

// ArticleJSONLD is the structured data payload embedded on article detail pages.
type ArticleJSONLD struct {
	Context              string               `json:"@context"`
	Type                 string               `json:"@type"`
	Headline             string               `json:"headline"`
	Description          string               `json:"description,omitempty"`
	Text                 string               `json:"text,omitempty"`
	Author               Person               `json:"author"`
	Publisher            Organization         `json:"publisher"`
	DatePublished        string               `json:"datePublished"`
	DateModified         string               `json:"dateModified,omitempty"`
	URL                  string               `json:"url"`
	MainEntityOfPage     string               `json:"mainEntityOfPage,omitempty"`
	ArticleSection       string               `json:"articleSection,omitempty"`
	Keywords             []string             `json:"keywords,omitempty"`
	CommentCount         uint64               `json:"commentCount,omitempty"`
	Comment              []Comment            `json:"comment,omitempty"`
	InteractionStatistic []InteractionCounter `json:"interactionStatistic,omitempty"`
}
