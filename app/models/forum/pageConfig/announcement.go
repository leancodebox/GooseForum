package pageConfig

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
)

type AnnouncementItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Enabled bool   `json:"enabled"`
	HTML    string `json:"-"`
}

type AnnouncementConfig struct {
	Enabled     bool               `json:"enabled"`
	Content     string             `json:"content"` // Read compatibility for pre-list configurations.
	PublishedAt int64              `json:"publishedAt,omitempty"`
	Items       []AnnouncementItem `json:"items"`
	HtmlContent string             `json:"-"`
}

// Old configurations stored a date string; new configurations use Unix milliseconds.
func (c *AnnouncementConfig) UnmarshalJSON(data []byte) error {
	type config AnnouncementConfig
	var next AnnouncementConfig
	decoded := struct {
		*config
		PublishedAt json.RawMessage `json:"publishedAt"`
	}{config: (*config)(&next)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if len(decoded.PublishedAt) > 0 && string(decoded.PublishedAt) != "null" {
		if decoded.PublishedAt[0] == '"' {
			var old string
			if err := json.Unmarshal(decoded.PublishedAt, &old); err != nil {
				return err
			}
			if old != "" {
				date, err := time.Parse(time.RFC3339Nano, old)
				if err != nil {
					date, err = time.ParseInLocation(time.DateTime, old, time.Local)
				}
				if err != nil {
					return err
				}
				next.PublishedAt = date.UnixMilli()
			}
		} else if err := json.Unmarshal(decoded.PublishedAt, &next.PublishedAt); err != nil {
			return err
		}
	}
	*c = next
	return nil
}

// A present empty list is authoritative: deleted announcements cannot fall back to old content.
func (c AnnouncementConfig) Canonical() AnnouncementConfig {
	if c.Items == nil && strings.TrimSpace(c.Content) != "" {
		c.Items = []AnnouncementItem{{ID: "legacy", Content: c.Content, Enabled: true}}
	} else {
		c.Items = append([]AnnouncementItem{}, c.Items...)
	}
	c.Content = ""
	c.HtmlContent = ""
	for i := range c.Items {
		c.Items[i].ID = strings.TrimSpace(c.Items[i].ID)
		c.Items[i].Title = strings.TrimSpace(c.Items[i].Title)
	}
	return c
}

func (c *AnnouncementConfig) PrepareHTML() {
	if c == nil {
		return
	}
	*c = c.Canonical()
	for i := range c.Items {
		if c.Items[i].HTML == "" && c.Items[i].Enabled && strings.TrimSpace(c.Items[i].Content) != "" {
			c.Items[i].HTML = markdown2html.MarkdownToHTML(c.Items[i].Content)
		}
	}
}

func (c AnnouncementConfig) ActiveItems() []AnnouncementItem {
	items := make([]AnnouncementItem, 0, len(c.Items))
	if !c.Enabled {
		return items
	}
	for _, item := range c.Canonical().Items {
		if !item.Enabled || strings.TrimSpace(item.Content) == "" {
			continue
		}
		if item.HTML == "" {
			item.HTML = markdown2html.MarkdownToHTML(item.Content)
		}
		items = append(items, item)
	}
	return items
}

func (c AnnouncementConfig) GetHtmlContent() string {
	items := c.ActiveItems()
	if len(items) == 0 {
		return ""
	}
	return items[0].HTML
}

// The total timestamp changes only for meaningful configuration changes, including order.
func (c *AnnouncementConfig) SetUpdateTime(previous AnnouncementConfig, now time.Time) {
	old := previous.Canonical()
	equal := c.Enabled == old.Enabled && len(c.Items) == len(old.Items)
	if equal {
		for i, item := range c.Items {
			other := old.Items[i]
			if item.ID != other.ID || item.Title != other.Title || item.Content != other.Content || item.Enabled != other.Enabled {
				equal = false
				break
			}
		}
	}
	c.PublishedAt = previous.PublishedAt
	if equal {
		return
	}
	c.PublishedAt = now.UnixMilli()
	if c.PublishedAt <= previous.PublishedAt {
		c.PublishedAt = previous.PublishedAt + 1
	}
}
