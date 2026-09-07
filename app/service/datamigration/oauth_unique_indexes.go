package datamigration

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"gorm.io/gorm"
)

type OAuthUniqueIndexMigrationResult struct {
	DuplicatesRemoved int64
	Failed            int
	LastFailed        string
}

// EnforceOAuthBindingUniqueness keeps the oldest binding for each external
// identity and each user/provider pair, then upgrades the legacy indexes.
func EnforceOAuthBindingUniqueness() OAuthUniqueIndexMigrationResult {
	return EnforceOAuthBindingUniquenessWithDB(dbconnect.Connect())
}

func EnforceOAuthBindingUniquenessWithDB(db *gorm.DB) OAuthUniqueIndexMigrationResult {
	result := OAuthUniqueIndexMigrationResult{}
	err := db.Transaction(func(tx *gorm.DB) error {
		var bindings []userOAuth.Entity
		if err := tx.Order("id ASC").Find(&bindings).Error; err != nil {
			return err
		}
		seenIdentity := make(map[string]bool, len(bindings))
		seenUserProvider := make(map[string]bool, len(bindings))
		duplicateIDs := make([]uint64, 0)
		for _, binding := range bindings {
			identityKey := binding.Provider + "\x00" + binding.ProviderUid
			userProviderKey := fmt.Sprintf("%d\x00%s", binding.UserId, binding.Provider)
			if seenIdentity[identityKey] || seenUserProvider[userProviderKey] {
				duplicateIDs = append(duplicateIDs, binding.Id)
				continue
			}
			seenIdentity[identityKey] = true
			seenUserProvider[userProviderKey] = true
		}
		if len(duplicateIDs) > 0 {
			deleteResult := tx.Delete(&userOAuth.Entity{}, duplicateIDs)
			if deleteResult.Error != nil {
				return deleteResult.Error
			}
			result.DuplicatesRemoved = deleteResult.RowsAffected
		}
		for _, index := range []string{"idx_provider_uid", "idx_user_provider"} {
			if tx.Migrator().HasIndex(&userOAuth.Entity{}, index) {
				if err := tx.Migrator().DropIndex(&userOAuth.Entity{}, index); err != nil {
					return fmt.Errorf("drop %s: %w", index, err)
				}
			}
		}
		for _, statement := range []string{
			"CREATE UNIQUE INDEX idx_provider_uid ON user_o_auth(provider, provider_uid)",
			"CREATE UNIQUE INDEX idx_user_provider ON user_o_auth(user_id, provider)",
		} {
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("create OAuth unique index: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		result.Failed = 1
		result.LastFailed = err.Error()
	}
	return result
}

func ClearStoredOAuthTokens() error {
	return dbconnect.Connect().Model(&userOAuth.Entity{}).Where("1 = 1").Updates(map[string]any{
		"access_token": "", "refresh_token": "", "scopes": "", "raw_user_data": "",
	}).Error
}
