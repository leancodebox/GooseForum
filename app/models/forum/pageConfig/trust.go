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
	MaxDailyUploads         *int             `json:"maxDailyUploads"`
	MaxDailyReplies         *int             `json:"maxDailyReplies"`
	MaxDailyTopics          *int             `json:"maxDailyTopics"`
	Level                   trustlevel.Level `json:"level"`
	AllowImageUpload        bool             `json:"allowImageUpload"`
	MaxImagesPerPost        *int             `json:"maxImagesPerPost"`
	MaxExternalLinksPerPost *int             `json:"maxExternalLinksPerPost"`
	MaxMentionsPerPost      *int             `json:"maxMentionsPerPost"`
}

type RegistrationCooldown struct {
	MinPostAccountAgeMinutes   int `json:"minPostAccountAgeMinutes"`
	MinUploadAccountAgeMinutes int `json:"minUploadAccountAgeMinutes"`
}

func (c RegistrationCooldown) Valid() bool {
	return c.MinPostAccountAgeMinutes >= 0 && c.MinPostAccountAgeMinutes <= 5256000 && c.MinUploadAccountAgeMinutes >= 0 && c.MinUploadAccountAgeMinutes <= 5256000
}

type TrustConfig struct {
	RegistrationCooldown RegistrationCooldown `json:"registrationCooldown"`
	SchemaVersion        int                  `json:"schemaVersion"`
	PromotionRules       []PromotionRule      `json:"promotionRules"`
	LevelLimits          []TrustLimits        `json:"levelLimits"`
}

func DefaultTrustSettings() TrustConfig {
	c := TrustConfig{SchemaVersion: 3, RegistrationCooldown: RegistrationCooldown{MinUploadAccountAgeMinutes: 1440}, PromotionRules: []PromotionRule{{trustlevel.Member, true, 24, 2, 5}, {trustlevel.Regular, true, 336, 7, 20}}}
	for _, d := range trustlevel.Definitions() {
		l := TrustLimits{Level: d.Level, AllowImageUpload: true, MaxDailyTopics: new(10), MaxDailyUploads: new(10)}
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
	if c.SchemaVersion != 3 || !c.RegistrationCooldown.Valid() || len(c.PromotionRules) != 2 || len(c.LevelLimits) != 4 {
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

type TrustLimitOrderError struct {
	Field       string
	LowerLevel  trustlevel.Level
	HigherLevel trustlevel.Level
}

func (e *TrustLimitOrderError) Error() string {
	return "higher trust level has a stricter limit: " + e.Field
}

// Validate on save, so older policies remain readable and can be corrected.
func (c TrustConfig) ValidateLimitOrder() *TrustLimitOrderError {
	limits := make(map[trustlevel.Level]TrustLimits, len(c.LevelLimits))
	for _, limit := range c.LevelLimits {
		limits[limit.Level] = limit
	}
	definitions := trustlevel.Definitions()
	for i := 1; i < len(definitions); i++ {
		lower, lowerOK := limits[definitions[i-1].Level]
		higher, higherOK := limits[definitions[i].Level]
		if !lowerOK || !higherOK {
			continue
		}
		if lower.AllowImageUpload && !higher.AllowImageUpload {
			return &TrustLimitOrderError{"allowImageUpload", lower.Level, higher.Level}
		}
		for _, pair := range []struct {
			field         string
			lower, higher *int
		}{
			{"maxDailyTopics", lower.MaxDailyTopics, higher.MaxDailyTopics},
			{"maxDailyReplies", lower.MaxDailyReplies, higher.MaxDailyReplies},
			{"maxDailyUploads", lower.MaxDailyUploads, higher.MaxDailyUploads},
			{"maxImagesPerPost", lower.MaxImagesPerPost, higher.MaxImagesPerPost},
			{"maxExternalLinksPerPost", lower.MaxExternalLinksPerPost, higher.MaxExternalLinksPerPost},
			{"maxMentionsPerPost", lower.MaxMentionsPerPost, higher.MaxMentionsPerPost},
		} {
			if pair.higher != nil && (pair.lower == nil || *pair.higher < *pair.lower) {
				return &TrustLimitOrderError{pair.field, lower.Level, higher.Level}
			}
		}
	}
	return nil
}

func DecodeTrustSettings(raw []byte) (TrustConfig, error) {
	var header struct {
		SchemaVersion int `json:"schemaVersion"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return TrustConfig{}, err
	}
	if header.SchemaVersion == 2 {
		var root map[string]json.RawMessage
		if err := json.Unmarshal(raw, &root); err != nil {
			return TrustConfig{}, err
		}
		var limits []map[string]json.RawMessage
		if err := json.Unmarshal(root["levelLimits"], &limits); err != nil {
			return TrustConfig{}, err
		}
		cooldown := RegistrationCooldown{}
		for _, limit := range limits {
			var old RegistrationCooldown
			encoded, _ := json.Marshal(limit)
			if err := json.Unmarshal(encoded, &old); err != nil {
				return TrustConfig{}, err
			}
			if !old.Valid() {
				return TrustConfig{}, errors.New("invalid legacy registration cooldown")
			}
			cooldown.MinPostAccountAgeMinutes = max(cooldown.MinPostAccountAgeMinutes, old.MinPostAccountAgeMinutes)
			cooldown.MinUploadAccountAgeMinutes = max(cooldown.MinUploadAccountAgeMinutes, old.MinUploadAccountAgeMinutes)
			delete(limit, "minPostAccountAgeMinutes")
			delete(limit, "minUploadAccountAgeMinutes")
		}
		root["schemaVersion"] = json.RawMessage(`3`)
		root["levelLimits"], _ = json.Marshal(limits)
		root["registrationCooldown"], _ = json.Marshal(cooldown)
		raw, _ = json.Marshal(root)
	}
	// Accept the retired switch in stored configurations without returning it.
	var legacy struct {
		TrustConfig
		AutoPromotionEnabled *bool `json:"autoPromotionEnabled"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&legacy); err != nil {
		return legacy.TrustConfig, err
	}
	c := legacy.TrustConfig
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
