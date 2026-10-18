package pageConfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"gorm.io/gorm"
)

const TrustSettings = "trustSettings"

type PromotionRule struct {
	Level              trustlevel.Level `json:"level"`
	Enabled            bool             `json:"enabled"`
	MinAccountAgeHours int              `json:"minAccountAgeHours"`
	MinVisitDays       uint32           `json:"minVisitDays"`
	MinTopicsVisited   int              `json:"minTopicsVisited"`
}

type TrustLimits struct {
	MaxDailyUploads            *int             `json:"maxDailyUploads"`
	MinPostAccountAgeMinutes   int              `json:"minPostAccountAgeMinutes"`
	MinUploadAccountAgeMinutes int              `json:"minUploadAccountAgeMinutes"`
	MaxDailyReplies            *int             `json:"maxDailyReplies"`
	MaxDailyTopics             *int             `json:"maxDailyTopics"`
	Level                      trustlevel.Level `json:"level"`
	AllowImageUpload           bool             `json:"allowImageUpload"`
	MaxImagesPerPost           *int             `json:"maxImagesPerPost"`
	MaxExternalLinksPerPost    *int             `json:"maxExternalLinksPerPost"`
	MaxMentionsPerPost         *int             `json:"maxMentionsPerPost"`
}

type TrustConfig struct {
	SchemaVersion        int             `json:"schemaVersion"`
	AutoPromotionEnabled bool            `json:"autoPromotionEnabled"`
	PromotionRules       []PromotionRule `json:"promotionRules"`
	LevelLimits          []TrustLimits   `json:"levelLimits"`
}

func DefaultTrustSettings() TrustConfig {
	c := TrustConfig{SchemaVersion: 2, AutoPromotionEnabled: true, PromotionRules: []PromotionRule{{trustlevel.Member, true, 24, 2, 5}, {trustlevel.Regular, true, 336, 7, 20}}}
	for _, d := range trustlevel.Definitions() {
		l := TrustLimits{Level: d.Level, AllowImageUpload: true, MaxDailyTopics: new(10), MaxDailyUploads: new(10), MinUploadAccountAgeMinutes: 1440}
		if d.Level == trustlevel.Newcomer {
			l.MaxDailyTopics = new(3)
			l.MaxDailyReplies = new(10)
			l.MaxImagesPerPost = new(1)
			l.MaxExternalLinksPerPost = new(2)
			l.MaxMentionsPerPost = new(2)
		}
		c.LevelLimits = append(c.LevelLimits, l)
	}
	return c
}

func (c TrustConfig) Valid() bool {
	if c.SchemaVersion != 2 || len(c.PromotionRules) != 2 || len(c.LevelLimits) != 4 {
		return false
	}
	rules := map[trustlevel.Level]PromotionRule{}
	for _, r := range c.PromotionRules {
		if (r.Level != trustlevel.Member && r.Level != trustlevel.Regular) || r.MinAccountAgeHours < 0 || r.MinAccountAgeHours > 87600 || r.MinVisitDays > 36500 || r.MinTopicsVisited < 0 || r.MinTopicsVisited > trustlevel.MaxTopics {
			return false
		}
		if _, ok := rules[r.Level]; ok {
			return false
		}
		rules[r.Level] = r
		if r.Enabled && r.MinAccountAgeHours == 0 && r.MinVisitDays == 0 && r.MinTopicsVisited == 0 {
			return false
		}
	}
	a, b := rules[trustlevel.Member], rules[trustlevel.Regular]
	if b.MinAccountAgeHours < a.MinAccountAgeHours || b.MinVisitDays < a.MinVisitDays || b.MinTopicsVisited < a.MinTopicsVisited {
		return false
	}
	seen := map[trustlevel.Level]bool{}
	for _, l := range c.LevelLimits {
		if l.MinPostAccountAgeMinutes < 0 || l.MinPostAccountAgeMinutes > 5256000 || l.MinUploadAccountAgeMinutes < 0 || l.MinUploadAccountAgeMinutes > 5256000 {
			return false
		}
		if l.MaxDailyUploads != nil && (*l.MaxDailyUploads < 0 || *l.MaxDailyUploads > 100000) {
			return false
		}
		if l.MaxDailyReplies != nil && (*l.MaxDailyReplies < 0 || *l.MaxDailyReplies > 100000) {
			return false
		}
		if l.MaxDailyTopics != nil && (*l.MaxDailyTopics < 0 || *l.MaxDailyTopics > 100000) {
			return false
		}
		if !trustlevel.Valid(l.Level) || seen[l.Level] {
			return false
		}
		seen[l.Level] = true
		for _, n := range []*int{l.MaxImagesPerPost, l.MaxExternalLinksPerPost, l.MaxMentionsPerPost} {
			if n != nil && (*n < 0 || *n > 100) {
				return false
			}
		}
	}
	return true
}

func DecodeTrustSettings(raw []byte) (TrustConfig, error) {
	var c TrustConfig
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&c); err != nil {
		return c, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return c, errors.New("invalid trust settings JSON")
	}
	if !c.Valid() {
		return c, errors.New("invalid trust settings")
	}
	return c, nil
}

func GetTrustSettings() (TrustConfig, error) {
	var row Entity
	err := builder().Where("page_type = ?", TrustSettings).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultTrustSettings(), nil
	}
	if err != nil {
		return TrustConfig{}, err
	}
	return DecodeTrustSettings([]byte(row.Config))
}
