package userBadges

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

type Recipient struct {
	UserId            uint64     `json:"userId"`
	Username          string     `json:"username"`
	AvatarUrl         string     `json:"avatarUrl"`
	Source            string     `json:"source"`
	Reason            string     `json:"reason"`
	GrantedBy         uint64     `json:"grantedBy"`
	GrantedByName     string     `json:"grantedByName"`
	GrantedAt         time.Time  `json:"grantedAt"`
	RestrictionStatus string     `json:"-"`
	RestrictionUntil  *time.Time `json:"-"`
}

func activeRecipients() *gorm.DB {
	return builder().
		Joins("JOIN users AS recipient ON recipient.id = user_badges.user_id AND recipient.deleted_at IS NULL").
		Where("user_badges.revoked_at IS NULL")
}

func ListRecipients(code string, page, pageSize int, search string) ([]Recipient, bool, error) {
	rows := make([]Recipient, 0)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	query := activeRecipients().Where("user_badges.badge_code = ?", code)
	search = strings.TrimSpace(search)
	if search != "" {
		if id, err := strconv.ParseUint(search, 10, 64); err == nil {
			query = query.Where("(recipient.id = ? OR recipient.username = ?)", id, search)
		} else {
			pattern := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(strings.ToLower(search))
			query = query.Where("LOWER(recipient.username) LIKE ? ESCAPE '!'", "%"+pattern+"%")
		}
	}
	err := query.
		Joins("LEFT JOIN users AS grantor ON grantor.id = user_badges.granted_by AND grantor.deleted_at IS NULL").
		Select("user_badges.user_id, recipient.username, recipient.avatar_url, recipient.restriction_status, recipient.restriction_until, user_badges.source, user_badges.reason, user_badges.granted_by, COALESCE(grantor.username, '') AS granted_by_name, user_badges.granted_at").
		Order("user_badges.granted_at DESC, user_badges.id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize + 1).Scan(&rows).Error
	if err != nil {
		return nil, false, err
	}
	hasNext := len(rows) > pageSize
	if hasNext {
		rows = rows[:pageSize]
	}
	return rows, hasNext, nil
}

func UpdateRecipient(userID uint64, code, action, reason string, actor uint64) (bool, error) {
	switch action {
	case "grant":
		return GrantManual(userID, code, SourceManual, reason, actor, nil)
	case "revoke":
		return revokeWithDB(dbconnect.Connect(), userID, code)
	default:
		return false, errors.New("invalid badge recipient action")
	}
}
