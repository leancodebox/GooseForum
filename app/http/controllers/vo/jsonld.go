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

// InteractionCounter represents schema.org interaction statistics for an article.
type InteractionCounter struct {
	Type                 string `json:"@type"`
	InteractionType      string `json:"interactionType"`
	UserInteractionCount uint64 `json:"userInteractionCount"`
}

// ArticleJSONLD is the structured data payload embedded on article detail pages.
type ArticleJSONLD struct {
	Context              string             `json:"@context"`
	Type                 string             `json:"@type"`
	Headline             string             `json:"headline"`
	Author               Person             `json:"author"`
	Publisher            Organization       `json:"publisher"`
	DatePublished        string             `json:"datePublished"`
	URL                  string             `json:"url"`
	InteractionStatistic InteractionCounter `json:"interactionStatistic"`
}
