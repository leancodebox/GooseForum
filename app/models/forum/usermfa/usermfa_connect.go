package usermfa

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

func builder() *gorm.DB { return dbconnect.Connect().Model(&Factor{}) }

func recoveryBuilder() *gorm.DB { return dbconnect.Connect().Model(&RecoveryCode{}) }
