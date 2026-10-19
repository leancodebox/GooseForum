package pageConfig

import (
	"slices"
	"testing"
)

func TestTrustLimitOrder(t *testing.T) {
	fields := []struct {
		name string
		set  func(*TrustLimits, *int)
	}{
		{"maxDailyTopics", func(l *TrustLimits, n *int) { l.MaxDailyTopics = n }},
		{"maxDailyReplies", func(l *TrustLimits, n *int) { l.MaxDailyReplies = n }},
		{"maxDailyUploads", func(l *TrustLimits, n *int) { l.MaxDailyUploads = n }},
		{"maxImagesPerPost", func(l *TrustLimits, n *int) { l.MaxImagesPerPost = n }},
		{"maxExternalLinksPerPost", func(l *TrustLimits, n *int) { l.MaxExternalLinksPerPost = n }},
		{"maxMentionsPerPost", func(l *TrustLimits, n *int) { l.MaxMentionsPerPost = n }},
	}
	for _, field := range fields {
		for _, tc := range []struct {
			name          string
			lower, higher *int
			invalid       bool
		}{
			{"decrease", new(10), new(9), true},
			{"unlimited-to-finite", nil, new(10), true},
			{"zero-after-positive", new(1), new(0), true},
			{"equal", new(0), new(0), false},
			{"increase", new(0), new(1), false},
			{"finite-to-unlimited", new(10), nil, false},
			{"both-unlimited", nil, nil, false},
		} {
			t.Run(field.name+"/"+tc.name, func(t *testing.T) {
				c := DefaultTrustSettings()
				for i := range c.LevelLimits {
					field.set(&c.LevelLimits[i], nil)
				}
				field.set(&c.LevelLimits[0], tc.lower)
				field.set(&c.LevelLimits[1], tc.higher)
				slices.Reverse(c.LevelLimits)
				issue := c.ValidateLimitOrder()
				if (issue != nil) != tc.invalid {
					t.Fatalf("issue=%+v", issue)
				}
				if issue != nil && (issue.Field != field.name || issue.LowerLevel != 0 || issue.HigherLevel != 100) {
					t.Fatalf("wrong conflict=%+v", issue)
				}
			})
		}
	}
	for i := 1; i < 4; i++ {
		c := DefaultTrustSettings()
		c.LevelLimits[i].AllowImageUpload = false
		if issue := c.ValidateLimitOrder(); issue == nil || issue.Field != "allowImageUpload" || issue.HigherLevel != c.LevelLimits[i].Level {
			t.Fatalf("upload reversal accepted: %+v", issue)
		}
	}
	c := DefaultTrustSettings()
	c.LevelLimits[0].AllowImageUpload = false
	if issue := c.ValidateLimitOrder(); issue != nil {
		t.Fatal("permission expansion rejected", issue)
	}
}
