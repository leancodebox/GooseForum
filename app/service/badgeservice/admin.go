package badgeservice

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/leancodebox/GooseForum/app/bundles/randopt"
	"github.com/leancodebox/GooseForum/app/models/forum/badgeStats"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/notificationservice"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"gorm.io/gorm"
)

var (
	ErrBadgeNotFound     = errors.New("badge not found")
	ErrBadgeInUse        = errors.New("badge has grant history")
	ErrSystemBadge       = errors.New("system badge cannot be deleted")
	ErrRecipientNotFound = errors.New("badge recipient not found")
)

type InvalidFieldError struct{ Field string }

func (e *InvalidFieldError) Error() string { return "invalid badge field: " + e.Field }
func invalidField(field string) error      { return &InvalidFieldError{Field: field} }

var badgeCodePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func systemDefinition(code string) (Badge, bool) {
	for _, badge := range systemDefinitions() {
		if badge.Code == code {
			return badge, true
		}
	}
	return Badge{}, false
}

// NormalizeDefinition validates the supported presentation and grant settings.
func NormalizeDefinition(badge *Badge) error {
	badge.Code = strings.TrimSpace(badge.Code)
	badge.Name = strings.TrimSpace(badge.Name)
	badge.Description = strings.TrimSpace(badge.Description)
	badge.IconURL = strings.TrimSpace(badge.IconURL)
	if badge.Name == "" || utf8.RuneCountInString(badge.Name) > 64 {
		return invalidField("name")
	}
	if utf8.RuneCountInString(badge.Description) > 255 {
		return invalidField("description")
	}
	if badge.Type == "" {
		badge.Type = badges.TypeCustom
	}
	if badge.Type != badges.TypeCustom && badge.Type != badges.TypeSystem {
		return invalidField("type")
	}
	if badge.GrantMode == "" {
		badge.GrantMode = badges.GrantModeManual
	}
	if def, system := systemDefinition(badge.Code); system {
		if badge.Type != badges.TypeSystem {
			return invalidField("type")
		}
		badge.GrantMode = def.GrantMode
	} else if badge.Type == badges.TypeSystem {
		return ErrBadgeNotFound
	} else if badge.GrantMode != badges.GrantModeManual {
		return invalidField("grantMode")
	}
	if badge.Level == "" {
		badge.Level = LevelBronze
	}
	switch badge.Level {
	case LevelBronze, LevelSilver, LevelGold, LevelSpecial:
	default:
		return invalidField("level")
	}
	if badge.Color == "" {
		badge.Color = "blue"
	}
	switch badge.Color {
	case "blue", "emerald", "teal", "sky", "cyan", "rose", "violet", "purple", "fuchsia", "indigo", "amber", "orange", "yellow", "slate":
	default:
		return invalidField("color")
	}
	if badge.IconType == "" {
		badge.IconType = badges.IconTypeAsset
	}
	if badge.IconType != badges.IconTypeAsset {
		return invalidField("iconType")
	}
	if !validIconURL(badge.IconURL) {
		return invalidField("iconUrl")
	}
	badge.IconKey = ""
	return nil
}

func validIconURL(value string) bool {
	if value == "" || len(value) > 255 || strings.ContainsAny(value, "\\") {
		return false
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.User != nil || parsed.Fragment != "" || parsed.Opaque != "" {
		return false
	}
	if parsed.Scheme == "" {
		return strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && parsed.Host == ""
	}
	return (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Hostname() != ""
}

func SaveDefinition(badge Badge) error {
	if err := NormalizeDefinition(&badge); err != nil {
		return err
	}
	if badge.Code == "" {
		// Codes are internal identifiers; the creation form does not need to ask for one.
		badge.Code = "custom_" + strings.ToLower(randopt.RandomString(16))
	}
	entity, err := badges.GetByCodeWithError(badge.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// Older versions allowed arbitrary codes. They are read-only in the editor,
	// so apply the new identifier format only when creating a definition.
	if entity.Id == 0 && !badgeCodePattern.MatchString(badge.Code) {
		return invalidField("code")
	}
	if entity.Id != 0 && entity.Type != badge.Type {
		return invalidField("type")
	}
	entity.Code, entity.Type, entity.GrantMode = badge.Code, badge.Type, badge.GrantMode
	entity.Name, entity.Description = badge.Name, badge.Description
	entity.IconType, entity.IconKey, entity.IconURL = badge.IconType, badge.IconKey, badge.IconURL
	entity.Color, entity.Level = badge.Color, badge.Level
	entity.IsEnabled, entity.IsWearable, entity.SortOrder = badge.IsEnabled, badge.IsWearable, badge.SortOrder
	if err := badges.Save(&entity); err != nil {
		return err
	}
	InvalidateDefinitions()
	return nil
}

// ListForAdmin reads persisted grant statistics independently of cached definitions.
func ListForAdmin() ([]AdminBadge, error) {
	items := AllForAdmin()
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
	}
	stats, err := badgeStats.GetByCodes(codes)
	if err != nil {
		return nil, err
	}
	for i := range items {
		stat, ok := stats[items[i].Code]
		if !ok {
			return nil, fmt.Errorf("badge statistics missing for %s: %w", items[i].Code, gorm.ErrRecordNotFound)
		}
		items[i].GrantedCount = stat.GrantedCount
		items[i].CanDelete = !items[i].IsSystem && !stat.HasGrantHistory
	}
	return items, nil
}

func DeleteDefinition(code string) error {
	if _, system := systemDefinition(code); system {
		return ErrSystemBadge
	}
	entity, err := badges.GetByCodeWithError(code)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrBadgeNotFound
	}
	if err != nil {
		return err
	}
	if entity.Type != badges.TypeCustom {
		return ErrSystemBadge
	}
	deleted, err := badges.DeleteUnusedCustom(code)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrBadgeInUse
	}
	InvalidateDefinitions()
	return nil
}

func UpdateRecipient(code string, userID uint64, action, reason string, actor uint64) error {
	reason = strings.TrimSpace(reason)
	if reason == "" || utf8.RuneCountInString(reason) > 255 {
		return invalidField("reason")
	}
	if action != "grant" && action != "revoke" {
		return invalidField("action")
	}
	if userID == 0 {
		return ErrRecipientNotFound
	}
	if _, err := users.GetIdentity(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecipientNotFound
		}
		return err
	}
	badge := ResolveOne(code)
	if badge.Code == "" {
		return ErrBadgeNotFound
	}
	if action == "grant" && (!badge.IsEnabled || badge.GrantMode != badges.GrantModeManual) {
		return invalidField("grantMode")
	}
	changed, err := userBadges.UpdateRecipient(userID, code, action, reason, actor)
	if err != nil || !changed {
		return err
	}
	if action == "grant" {
		optlogger.UserOptCode(actor, optlogger.EditUser, userID, "admin.opt.badge.granted", optlogger.MessageParams{"badgeCode": code, "badgeName": badge.Name, "reason": reason})
		// A successful re-grant is also visible to the recipient.
		if err := notificationservice.SendBadgeNotification(userID, badge.Code, badge.Name, badge.IconURL); err != nil {
			// The persisted grant remains successful when notification delivery fails.
			slog.Warn("badge notification failed", "user_id", userID, "badge", code, "err", err)
		}
	} else {
		optlogger.UserOptCode(actor, optlogger.EditUser, userID, "admin.opt.badge.revoked", optlogger.MessageParams{"badgeCode": code, "badgeName": badge.Name, "reason": reason})
	}
	return nil
}

// SaveManualSelection preserves automatic and disabled awards when editing a user's manual selection.
func SaveManualSelection(userID, actor uint64, codes []string) error {
	if userID == 0 {
		return ErrRecipientNotFound
	}
	if _, err := users.GetIdentity(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecipientNotFound
		}
		return err
	}
	allowed := make(map[string]bool)
	for _, badge := range ManualGrantBadgesForAdmin() {
		allowed[badge.Code] = true
	}
	next := make(map[string]bool)
	for _, code := range codes {
		if !allowed[code] {
			return invalidField("badgeCodes")
		}
		next[code] = true
	}
	active, err := userBadges.GetActiveByUserIDWithError(userID)
	if err != nil {
		return err
	}
	for code := range next {
		if err := UpdateRecipient(code, userID, "grant", "管理员手动授予", actor); err != nil {
			return fmt.Errorf("grant badge: %w", err)
		}
	}
	for _, record := range active {
		if record.Source == userBadges.SourceManual && allowed[record.BadgeCode] && !next[record.BadgeCode] {
			if err := UpdateRecipient(record.BadgeCode, userID, "revoke", "管理员取消授予", actor); err != nil {
				return fmt.Errorf("revoke badge: %w", err)
			}
		}
	}
	return nil
}
