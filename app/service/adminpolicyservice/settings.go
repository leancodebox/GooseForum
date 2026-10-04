package adminpolicyservice

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

var ErrNoUsableAdministrator = errors.New("email verification would leave no usable administrator")

func SaveSecuritySettings(settings pageConfig.SecurityAndRegistration) error {
	var err error
	if !settings.EnableEmailVerification {
		err = pageConfig.SaveSecuritySettingsForPolicy(settings)
	} else {
		err = func() error {
			snapshot, err := ReadSnapshot()
			if err != nil {
				return err
			}
			remaining, err := users.HasNormalAdministrator(snapshot.AdminRoleIDs, 0, 0, time.Now(), true)
			if err != nil {
				return err
			}
			if !remaining {
				return ErrNoUsableAdministrator
			}
			return pageConfig.SaveSecuritySettingsForPolicy(settings)
		}()
	}
	if err == nil {
		hotdataserve.ClearSecuritySettingsConfigCache()
	}
	return err
}
