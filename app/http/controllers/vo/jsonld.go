package vo

type Person struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type Organization struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type InteractionCounter struct {
	Type                 string `json:"@type"`
	InteractionType      string `json:"interactionType"`
	UserInteractionCount uint64 `json:"userInteractionCount"`
}

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
