package oidcProviderStore

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

func tokenBuilder() *gorm.DB { return dbconnect.Connect().Model(&TokenEntity{}) }

func codeBuilder() *gorm.DB { return dbconnect.Connect().Model(&AuthorizationCodeEntity{}) }
