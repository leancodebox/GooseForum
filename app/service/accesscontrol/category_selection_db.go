package accesscontrol

import (
	"errors"
	"fmt"

	"github.com/leancodebox/GooseForum/app/models/forum/accessGroups"
	"github.com/leancodebox/GooseForum/app/models/forum/categoryGroupPermissions"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ValidateRestrictedCategorySelectionWithDB rechecks the single-restricted-
// category invariant against current database grants. Callers must run it in
// the same transaction that persists the topic. On databases that support row
// locking, the everyone grants are locked so a concurrent restriction change
// cannot pass between validation and the topic write.
func ValidateRestrictedCategorySelectionWithDB(tx *gorm.DB, categoryIDs []uint64) error {
	categoryIDs = uniqueNonZeroPreservingOrder(categoryIDs)
	if len(categoryIDs) <= 1 {
		return nil
	}
	if tx == nil {
		return errors.New("access control database is unavailable")
	}

	_, public, err := LockPublicCategoryStateWithDB(tx, categoryIDs)
	if err != nil {
		return err
	}
	for _, categoryID := range categoryIDs {
		if _, ok := public[categoryID]; !ok {
			return ErrRestrictedCategorySingle
		}
	}
	return nil
}

// LockPublicCategoryStateWithDB returns the everyone group and the requested
// categories that are currently public while locking the relevant rows when
// supported by the database.
func LockPublicCategoryStateWithDB(tx *gorm.DB, categoryIDs []uint64) (uint64, map[uint64]struct{}, error) {
	if tx == nil {
		return 0, nil, errors.New("access control database is unavailable")
	}
	categoryIDs = uniqueNonZeroPreservingOrder(categoryIDs)
	var everyone accessGroups.Entity
	// System groups are immutable, so locking this shared row would only
	// serialize otherwise unrelated multi-category topic writes.
	if err := tx.
		Where("system_key = ? AND status = ?", accessGroups.SystemKeyEveryone, accessGroups.StatusEnabled).
		First(&everyone).Error; err != nil {
		return 0, nil, fmt.Errorf("load everyone access group: %w", err)
	}

	var grants []categoryGroupPermissions.Entity
	grantQuery := tx.
		Where("access_group_id = ? AND category_id IN ?", everyone.Id, categoryIDs).
		Order("category_id ASC")
	grantQuery = withUpdateLock(grantQuery)
	if err := grantQuery.Find(&grants).Error; err != nil {
		return 0, nil, fmt.Errorf("load public category grants: %w", err)
	}
	public := make(map[uint64]struct{}, len(grants))
	for _, grant := range grants {
		if grant.Status == categoryGroupPermissions.StatusEnabled && grant.PermissionLevel >= categoryGroupPermissions.PermissionRead {
			public[grant.CategoryId] = struct{}{}
		}
	}
	return everyone.Id, public, nil
}

func withUpdateLock(query *gorm.DB) *gorm.DB {
	if query.Dialector.Name() == "sqlite" {
		return query
	}
	return query.Clauses(clause.Locking{Strength: "UPDATE"})
}
