package pageConfig

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAnnouncementTimestampCompatibility(t *testing.T) {
	date := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	localDate, _ := time.ParseInLocation(time.DateTime, "2026-10-03 10:00:00", time.Local)
	for _, test := range []struct {
		name string
		json string
		want int64
	}{
		{"numeric", `{"enabled":true,"content":"Notice","publishedAt":1791021600000}`, date.UnixMilli()},
		{"RFC3339", `{"enabled":true,"content":"Notice","publishedAt":"2026-10-03T10:00:00Z"}`, date.UnixMilli()},
		{"local date", `{"enabled":true,"content":"Notice","publishedAt":"2026-10-03 10:00:00"}`, localDate.UnixMilli()},
		{"empty", `{"enabled":true,"content":"Notice","publishedAt":""}`, 0},
		{"absent", `{"enabled":true,"content":"Notice"}`, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			var config AnnouncementConfig
			if err := json.Unmarshal([]byte(test.json), &config); err != nil {
				t.Fatal(err)
			}
			if config.PublishedAt != test.want || !config.Enabled || config.Content != "Notice" {
				t.Fatalf("wrong decoded config: %+v", config)
			}
			encoded, err := json.Marshal(config)
			if err != nil {
				t.Fatal(err)
			}
			var restored AnnouncementConfig
			if err := json.Unmarshal(encoded, &restored); err != nil || restored.PublishedAt != test.want {
				t.Fatalf("timestamp round trip failed: %s, %v", encoded, err)
			}
		})
	}
}

func TestAnnouncementLegacyAndEmptyList(t *testing.T) {
	legacy := AnnouncementConfig{Enabled: true, Content: "**Legacy notice**"}
	legacy.PrepareHTML()
	items := legacy.ActiveItems()
	if len(items) != 1 || items[0].ID != "legacy" || items[0].HTML == "" {
		t.Fatalf("legacy was not migrated: %+v", items)
	}
	legacy.Items = []AnnouncementItem{}
	legacy.Content = "old content must not return"
	if got := legacy.GetHtmlContent(); got != "" {
		t.Fatalf("deleted notice returned: %s", got)
	}
	encoded, err := json.Marshal(legacy.Canonical())
	if err != nil {
		t.Fatal(err)
	}
	var restored AnnouncementConfig
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Items == nil || restored.GetHtmlContent() != "" {
		t.Fatalf("empty list lost its authority: %s", encoded)
	}
}

func TestAnnouncementActiveOrderAndCacheIsolation(t *testing.T) {
	config := AnnouncementConfig{Enabled: true, Items: []AnnouncementItem{
		{ID: "second", Content: "Second", Enabled: true},
		{ID: "disabled", Content: "Disabled", Enabled: false},
		{ID: "empty", Content: "  ", Enabled: true},
		{ID: "first", Content: "First", Enabled: true},
	}}
	config.PrepareHTML()
	active := config.ActiveItems()
	if len(active) != 2 || active[0].ID != "second" || active[1].ID != "first" {
		t.Fatalf("wrong order: %+v", active)
	}
	copy := config.Canonical()
	copy.Items[0].Content = "Changed"
	if config.Items[0].Content != "Second" {
		t.Fatal("editing modified the cached slice")
	}
	config.Enabled = false
	if len(config.ActiveItems()) != 0 {
		t.Fatal("globally disabled announcements are visible")
	}
	config.Enabled = true
	config.Items[0].Enabled = false
	config.Items[3].Enabled = false
	if config.GetHtmlContent() != "" {
		t.Fatal("disabled notices returned through compatibility HTML")
	}
}

func TestAnnouncementTotalUpdateTime(t *testing.T) {
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	previous := AnnouncementConfig{Enabled: true, PublishedAt: now.UnixMilli(), Items: []AnnouncementItem{
		{ID: "a", Content: "A", Enabled: true}, {ID: "b", Content: "B", Enabled: true},
	}}
	same := previous.Canonical()
	same.PublishedAt = now.Add(24 * time.Hour).UnixMilli()
	same.SetUpdateTime(previous, now.Add(time.Hour))
	if same.PublishedAt != previous.PublishedAt {
		t.Fatal("unchanged save advanced the timestamp")
	}
	changed := previous.Canonical()
	changed.Items[0], changed.Items[1] = changed.Items[1], changed.Items[0]
	changed.SetUpdateTime(previous, now)
	if changed.PublishedAt == previous.PublishedAt {
		t.Fatal("sorting did not advance the timestamp")
	}
	next := changed.Canonical()
	next.Items[0].Title = "Updated title"
	next.SetUpdateTime(changed, now)
	if next.PublishedAt == changed.PublishedAt {
		t.Fatal("changes in the same clock tick were not detected")
	}
}
