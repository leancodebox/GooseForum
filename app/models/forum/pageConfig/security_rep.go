package pageConfig

import (
	"errors"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"

	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"gorm.io/gorm"
)

func SecuritySettingsForPolicy(defaults SecurityAndRegistration) (SecurityAndRegistration, error) {
	var row struct{ Config string }
	err := dbconnect.Connect().Model(&Entity{}).Where("page_type = ?", SecuritySettings).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return defaults, nil
	}
	if err != nil {
		return SecurityAndRegistration{}, err
	}
	return DecodeSecuritySettings(row.Config, defaults), nil
}

func SaveSecuritySettingsForPolicy(settings SecurityAndRegistration) error {
	var row struct{ Id uint64 }
	err := dbconnect.Connect().Model(&Entity{}).Where("page_type = ?", SecuritySettings).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dbconnect.Connect().Create(&Entity{PageType: SecuritySettings, Config: jsonopt.Encode(settings)}).Error
	}
	if err != nil {
		return err
	}
	return dbconnect.Connect().Model(&Entity{}).Where("id = ?", row.Id).Update("config", jsonopt.Encode(settings)).Error
}
