package users

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"time"

	"gorm.io/gorm"
)

func GetAccountState(userID uint64) (AccountState, error) {
	var state AccountState
	err := dbconnect.Connect().Model(&EntityComplete{}).Where("id = ?", userID).First(&state).Error
	return state, err
}

type RestrictionAccount struct {
	Id                        uint64
	RoleId                    uint64
	IsFrozen                  int8
	RestrictionStatus         string
	RestrictionUntil          *time.Time
	RestrictionReason         string
	RestrictionNote           string
	IsActivated               int8
	RequiresEmailVerification bool
}

func (user RestrictionAccount) EffectiveRestriction(now time.Time) string {
	return effectiveRestriction(user.RestrictionStatus, user.IsFrozen, user.RestrictionUntil, now)
}

func GetRestrictionAccount(userID uint64) (RestrictionAccount, error) {
	var account RestrictionAccount
	err := dbconnect.Connect().Model(&EntityComplete{}).Where("id = ?", userID).First(&account).Error
	return account, err
}

func AccountsInRoles(roleIDs []uint64) ([]AccountState, error) {
	var accounts []AccountState
	if len(roleIDs) == 0 {
		return accounts, nil
	}
	err := dbconnect.Connect().Model(&EntityComplete{}).Where("role_id IN ?", roleIDs).Order("id").Find(&accounts).Error
	return accounts, err
}

func HasNormalAdministrator(roleIDs []uint64, excludeUserID, excludeRoleID uint64, now time.Time, verificationEnabled bool) (bool, error) {
	return hasNormalAdministrator(dbconnect.Connect(), roleIDs, excludeUserID, excludeRoleID, now, verificationEnabled)
}

func hasNormalAdministrator(db *gorm.DB, roleIDs []uint64, excludeUserID, excludeRoleID uint64, now time.Time, verificationEnabled bool) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}
	query := db.Model(&EntityComplete{}).Where("role_id IN ?", roleIDs).
		Where("restriction_status = ? OR (restriction_status = '' AND is_frozen = ?) OR (restriction_status <> '' AND restriction_until <= ?)", RestrictionNormal, StatusNormal, now)
	if verificationEnabled {
		query = query.Where("is_activated <> ?", ActivationPending)
	} else {
		query = query.Where("requires_email_verification = ? OR is_activated <> ?", false, ActivationPending)
	}
	if excludeUserID != 0 {
		query = query.Where("id <> ?", excludeUserID)
	}
	if excludeRoleID != 0 {
		query = query.Where("role_id <> ?", excludeRoleID)
	}
	var account struct{ Id uint64 }
	result := query.Limit(1).Find(&account)
	return result.RowsAffected > 0, result.Error
}

type RestrictionUpdate struct {
	RoleId         uint64
	Activation     int8
	Status         string
	Until          *time.Time
	Reason         string
	Note           string
	RevokeSessions bool
}

func UpdateRestriction(userID uint64, change RestrictionUpdate) error {
	frozen := StatusNormal
	if change.Status != RestrictionNormal {
		frozen = StatusFrozen
	}
	fields := map[string]any{"role_id": change.RoleId, "is_activated": change.Activation, "restriction_status": change.Status, "restriction_until": change.Until, "restriction_reason": change.Reason, "restriction_note": change.Note, "is_frozen": frozen}
	if change.RevokeSessions {
		fields["token_version"] = gorm.Expr("token_version + 1")
	}
	return dbconnect.Connect().Model(&EntityComplete{}).Where("id = ?", userID).Updates(fields).Error
}

func RoleMemberIDs(roleID uint64) ([]uint64, error) {
	var members []struct{ Id uint64 }
	if err := dbconnect.Connect().Model(&EntityComplete{}).Where("role_id = ?", roleID).Find(&members).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(members))
	for _, member := range members {
		ids = append(ids, member.Id)
	}
	return ids, nil
}

func RevokeRoleMembers(roleID uint64, removeRole bool) error {
	fields := map[string]any{"token_version": gorm.Expr("token_version + 1")}
	if removeRole {
		fields["role_id"] = 0
	}
	return dbconnect.Connect().Model(&EntityComplete{}).Where("role_id = ?", roleID).Updates(fields).Error
}
