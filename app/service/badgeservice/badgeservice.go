package badgeservice

import (
	"sort"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/localcache"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/service/eventnotice"
	"github.com/samber/lo"
)

const definitionsTTL = 10 * time.Minute

var adminBadgesCache = localcache.Cache[[]AdminBadge]{MaxEntries: 4}

type Badge struct {
	Code        string `json:"code"`
	Type        string `json:"type"`
	GrantMode   string `json:"grantMode"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconType    string `json:"iconType"`
	IconKey     string `json:"iconKey"`
	IconURL     string `json:"iconUrl"`
	Color       string `json:"color"`
	Level       string `json:"level"`
	IsEnabled   bool   `json:"isEnabled"`
	SortOrder   int    `json:"sortOrder"`
}

type AdminBadge struct {
	Badge
	IsSystem  bool `json:"isSystem"`
	CanDelete bool `json:"canDelete"`
}

type UserBadge struct {
	Badge
	Source    string `json:"source"`
	Reason    string `json:"reason"`
	GrantedAt string `json:"grantedAt"`
}

func AllForAdmin() []AdminBadge {
	items := adminBadgesCache.GetOrLoad("adminBadges", func() ([]AdminBadge, error) {
		return buildAllForAdmin(), nil
	}, definitionsTTL)
	return cloneAdminBadges(items)
}

func InvalidateDefinitions() {
	adminBadgesCache.Clear()
}

func buildAllForAdmin() []AdminBadge {
	systemDefs := systemDefinitions()
	overrides := lo.KeyBy(badges.All(), func(item *badges.Entity) string { return item.Code })
	seen := map[string]bool{}
	result := make([]AdminBadge, 0, len(systemDefs)+len(overrides))

	for _, def := range systemDefs {
		seen[def.Code] = true
		badge := applyOverride(def, overrides[def.Code])
		result = append(result, AdminBadge{Badge: badge, IsSystem: true, CanDelete: false})
	}

	for _, entity := range overrides {
		if entity == nil || seen[entity.Code] || entity.Type != badges.TypeCustom {
			continue
		}
		result = append(result, AdminBadge{Badge: fromEntity(entity), IsSystem: false, CanDelete: true})
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].SortOrder == result[j].SortOrder {
			return result[i].Code < result[j].Code
		}
		return result[i].SortOrder < result[j].SortOrder
	})
	return result
}

func cloneAdminBadges(items []AdminBadge) []AdminBadge {
	if len(items) == 0 {
		return []AdminBadge{}
	}
	result := make([]AdminBadge, len(items))
	copy(result, items)
	return result
}

func ManualGrantBadgesForAdmin() []Badge {
	items := AllForAdmin()
	result := make([]Badge, 0, len(items))
	for _, item := range items {
		if !item.IsEnabled || item.GrantMode != badges.GrantModeManual {
			continue
		}
		result = append(result, item.Badge)
	}
	return result
}

func GetUserBadges(userID uint64) []UserBadge {
	records := userBadges.GetActiveByUserID(userID)
	if len(records) == 0 {
		return []UserBadge{}
	}
	codes := lo.Map(records, func(item *userBadges.Entity, _ int) string { return item.BadgeCode })
	badgeMap := lo.KeyBy(ResolveByCodes(codes), func(item Badge) string { return item.Code })

	result := make([]UserBadge, 0, len(records))
	for _, record := range records {
		if record == nil {
			continue
		}
		badge, ok := badgeMap[record.BadgeCode]
		if !ok || !badge.IsEnabled {
			continue
		}
		result = append(result, UserBadge{
			Badge:     badge,
			Source:    record.Source,
			Reason:    record.Reason,
			GrantedAt: record.GrantedAt.Format(time.DateTime),
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].SortOrder == result[j].SortOrder {
			return result[i].GrantedAt > result[j].GrantedAt
		}
		return result[i].SortOrder < result[j].SortOrder
	})
	return result
}

func ResolveByCodes(codes []string) []Badge {
	if len(codes) == 0 {
		return []Badge{}
	}
	uniqueCodes := lo.Uniq(codes)
	allBadges := lo.KeyBy(AllForAdmin(), func(item AdminBadge) string { return item.Code })

	result := make([]Badge, 0, len(uniqueCodes))
	for _, code := range uniqueCodes {
		if item, ok := allBadges[code]; ok {
			result = append(result, item.Badge)
		}
	}
	return result
}

func Grant(userID uint64, code string, source string, reason string, grantedBy uint64) (bool, error) {
	badge := ResolveOne(code)
	if badge.Code == "" || !badge.IsEnabled {
		return false, nil
	}
	if source == "" {
		source = userBadges.SourceAuto
	}
	if reason == "" {
		reason = badge.Name
	}
	hadRecord := userBadges.Exists(userID, code)
	granted, err := userBadges.Grant(userID, code, source, reason, grantedBy, nil)
	if err == nil && granted && !hadRecord && source != userBadges.SourceMigration {
		_ = eventnotice.SendBadgeNotification(userID, badge.Code, badge.Name, badge.IconURL)
	}
	return granted, err
}

func ResolveOne(code string) Badge {
	items := ResolveByCodes([]string{code})
	if len(items) == 0 {
		return Badge{}
	}
	return items[0]
}

func applyOverride(def Definition, override *badges.Entity) Badge {
	item := Badge(def)
	if override == nil {
		return item
	}
	if override.Name != "" {
		item.Name = override.Name
	}
	if override.Description != "" {
		item.Description = override.Description
	}
	if override.IconType != "" {
		item.IconType = override.IconType
	}
	if override.IconKey != "" {
		item.IconKey = override.IconKey
	}
	if override.IconURL != "" {
		item.IconURL = override.IconURL
	}
	if override.Color != "" {
		item.Color = override.Color
	}
	if override.Level != "" {
		item.Level = override.Level
	}
	item.IsEnabled = override.IsEnabled
	if override.SortOrder != 0 {
		item.SortOrder = override.SortOrder
	}
	return item
}

func fromEntity(entity *badges.Entity) Badge {
	return Badge{
		Code:        entity.Code,
		Type:        entity.Type,
		GrantMode:   entity.GrantMode,
		Name:        entity.Name,
		Description: entity.Description,
		IconType:    entity.IconType,
		IconKey:     entity.IconKey,
		IconURL:     entity.IconURL,
		Color:       entity.Color,
		Level:       entity.Level,
		IsEnabled:   entity.IsEnabled,
		SortOrder:   entity.SortOrder,
	}
}
