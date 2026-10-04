package users

import (
	"strings"

	"gorm.io/gorm"
)

// Add nullable columns and populate them before AutoMigrate applies constraints.
type userIdentityColumns struct {
	UsernameLower   *string `gorm:"column:username_lower;type:varchar(64)"`
	EmailNormalized *string `gorm:"column:email_normalized;type:varchar(128)"`
}

func (userIdentityColumns) TableName() string { return "users" }

type userIdentityBackfill struct {
	Id       uint64
	Username string
	Email    string
}

func PrepareIdentitySchema() error {
	db := builder().Session(&gorm.Session{})
	if !db.Migrator().HasTable(&EntityComplete{}) {
		return nil
	}
	// Empty display emails are allowed; uniqueness belongs to EmailNormalized.
	for _, index := range []string{"ux_users_email", "ux_users_username"} {
		if db.Migrator().HasIndex(&EntityComplete{}, index) {
			if err := db.Migrator().DropIndex(&EntityComplete{}, index); err != nil {
				return err
			}
		}
	}
	if db.Migrator().HasColumn(&userIdentityColumns{}, "UsernameLower") && db.Migrator().HasColumn(&userIdentityColumns{}, "EmailNormalized") {
		columns, err := db.Migrator().ColumnTypes(&EntityComplete{})
		if err != nil {
			return err
		}
		for _, column := range columns {
			if column.Name() == "username_lower" {
				if nullable, known := column.Nullable(); known && !nullable {
					return nil
				}
			}
		}
	}
	for _, field := range []string{"UsernameLower", "EmailNormalized"} {
		if !db.Migrator().HasColumn(&userIdentityColumns{}, field) {
			if err := db.Migrator().AddColumn(&userIdentityColumns{}, field); err != nil {
				return err
			}
		}
	}
	var cursor uint64
	for {
		var batch []userIdentityBackfill
		result := db.Model(&EntityComplete{}).Unscoped().Where("id > ?", cursor).
			Where("username_lower IS NULL OR username_lower = '' OR (email_normalized IS NULL AND email IS NOT NULL AND email <> '') OR email IS NULL").
			Order("id ASC").Limit(500).Find(&batch)
		if result.Error != nil {
			return result.Error
		}
		if len(batch) == 0 {
			return nil
		}
		for _, row := range batch {
			var email any
			if normalized := strings.ToLower(strings.TrimSpace(row.Email)); normalized != "" {
				email = normalized
			}
			if err := db.Model(&EntityComplete{}).Unscoped().Where("id = ?", row.Id).
				UpdateColumns(map[string]any{"username_lower": strings.ToLower(strings.TrimSpace(row.Username)), "email_normalized": email, "email": row.Email}).Error; err != nil {
				return err
			}
			cursor = row.Id
		}
	}
}
