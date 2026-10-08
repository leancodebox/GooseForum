package agenttokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"gorm.io/gorm"
)

type Entity struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      uint64         `gorm:"index;not null" json:"-"`
	UserVersion uint64         `gorm:"not null" json:"-"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Hash        string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"-"`
	Prefix      string         `gorm:"type:varchar(20);not null" json:"prefix"`
	Scopes      []string       `gorm:"type:text;serializer:json;not null" json:"scopes"`
	CreatedAt   time.Time      `json:"createdAt"`
	ExpiresAt   time.Time      `gorm:"not null" json:"expiresAt"`
	RevokedAt   *time.Time     `json:"revokedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Entity) TableName() string { return "agent_tokens" }

func Digest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func ValidScopes(scopes []string) bool {
	if !slices.Contains(scopes, core.ScopeForumRead) {
		return false
	}
	seen := map[string]bool{}
	for _, scope := range scopes {
		if seen[scope] || scope != core.ScopeForumRead && scope != core.ScopeTopicsCreate && scope != core.ScopePostsCreate {
			return false
		}
		seen[scope] = true
	}
	return true
}

func Create(db *gorm.DB, userID, version uint64, name string, scopes []string, days int) (Entity, string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 100 || !ValidScopes(scopes) || days != 7 && days != 30 && days != 90 {
		return Entity{}, "", errors.New("invalid token parameters")
	}
	var entropy [32]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return Entity{}, "", err
	}
	raw := "gf_agent_" + base64.RawURLEncoding.EncodeToString(entropy[:])
	now := time.Now().UTC()
	entity := Entity{ID: uuid.NewString(), UserID: userID, UserVersion: version, Name: name, Hash: Digest(raw), Prefix: raw[:16], Scopes: scopes, CreatedAt: now, ExpiresAt: now.Add(time.Duration(days) * 24 * time.Hour)}
	err := db.Create(&entity).Error
	return entity, raw, err
}
