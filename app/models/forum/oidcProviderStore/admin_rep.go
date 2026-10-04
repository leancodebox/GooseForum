package oidcProviderStore

import (
	"context"
	"errors"
	"time"

	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"gorm.io/gorm"
)

func (s *Store) UpdateClient(ctx context.Context, client *core.Client) error {
	if client == nil {
		return errors.New("client is required")
	}
	if err := core.ValidateClient(*client); err != nil {
		return err
	}
	entity := clientToEntity(client)
	result := s.db.WithContext(ctx).Model(&ClientEntity{}).Where("client_id = ?", client.ID).
		Select("*").Omit("client_id", "created_at").Updates(&entity)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		existing, err := s.GetClient(ctx, client.ID)
		if err != nil {
			return err
		}
		if existing == nil {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

type AdminGrant struct {
	UserID    string    `json:"userId"`
	Scopes    []string  `json:"scopes" gorm:"serializer:json"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (s *Store) ListClientGrants(ctx context.Context, clientID, after, userID string) ([]AdminGrant, bool, error) {
	query := s.db.WithContext(ctx).Model(&ConsentEntity{}).Where("client_id = ? AND user_id > ?", clientID, after)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	var rows []AdminGrant
	if err := query.Order("user_id ASC").Limit(21).Find(&rows).Error; err != nil {
		return nil, false, err
	}
	more := len(rows) > 20
	if more {
		rows = rows[:20]
	}
	return rows, more, nil
}

func (s *Store) RevokeClientGrants(ctx context.Context, clientID string, now time.Time) error {
	db := s.db.WithContext(ctx)
	if err := db.Model(&AuthorizationCodeEntity{}).Where("client_id = ? AND used = ?", clientID, false).Update("used", true).Error; err != nil {
		return err
	}
	if err := db.Model(&TokenEntity{}).Where("client_id = ? AND revoked_at IS NULL", clientID).Update("revoked_at", now).Error; err != nil {
		return err
	}
	return db.Where("client_id = ?", clientID).Delete(&ConsentEntity{}).Error
}

func (s *Store) DeleteClient(ctx context.Context, clientID string, now time.Time) error {
	// Disable first so partial cleanup failures leave the registration unusable.
	if err := s.db.WithContext(ctx).Model(&ClientEntity{}).Where("client_id = ?", clientID).Update("enabled", false).Error; err != nil {
		return err
	}
	if err := s.RevokeClientGrants(ctx, clientID, now); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Where("client_id = ?", clientID).Delete(&ClientEntity{}).Error
}

func (s *Store) ResetSigningKey(ctx context.Context, key *SigningKeyEntity, now time.Time) error {
	db := s.db.WithContext(ctx)
	if err := db.Model(&AuthorizationCodeEntity{}).Where("used = ?", false).Update("used", true).Error; err != nil {
		return err
	}
	if err := db.Model(&TokenEntity{}).Where("revoked_at IS NULL").Update("revoked_at", now).Error; err != nil {
		return err
	}
	if err := db.Model(&InteractionEntity{}).Where("used = ?", false).Update("used", true).Error; err != nil {
		return err
	}
	if err := db.Where("1 = 1").Delete(&ConsentEntity{}).Error; err != nil {
		return err
	}
	if err := db.Where("1 = 1").Delete(&SigningKeyHistoryEntity{}).Error; err != nil {
		return err
	}
	key.ID = 1
	return db.Save(key).Error
}
