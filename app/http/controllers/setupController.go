package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/setupservice"
)

type SetupStatusReq struct{}

// GetSetupStatus 获取安装状态
func GetSetupStatus(req component.BetterRequest[SetupStatusReq]) component.Response {
	isInit := setupservice.IsInitialized()
	return component.SuccessResponse(component.DataMap{
		"isInit": isInit,
	})
}

type InitialSetupReq struct {
	SiteName string `json:"siteName" validate:"required"`
	SiteDesc string `json:"siteDesc"`
	// 数据库配置
	DBHost     string `json:"dbHost" validate:"required"`
	DBPort     string `json:"dbPort" validate:"required"`
	DBName     string `json:"dbName" validate:"required"`
	DBUser     string `json:"dbUser" validate:"required"`
	DBPassword string `json:"dbPassword" validate:"required"`
	// 管理员账号配置
	AdminUsername string `json:"adminUsername" validate:"required"`
	AdminPassword string `json:"adminPassword" validate:"required"`
	AdminEmail    string `json:"adminEmail" validate:"required,email"`
}

// InitialSetup 执行初始化设置
func InitialSetup(req component.BetterRequest[InitialSetupReq]) component.Response {
	if setupservice.IsInitialized() {
		return component.FailResponse("网站已经初始化")
	}

	// 1. 保存配置
	config := setupservice.SiteConfig{
		SiteName: req.Params.SiteName,
		SiteDesc: req.Params.SiteDesc,
		Database: setupservice.DatabaseConfig{
			Host:     req.Params.DBHost,
			Port:     req.Params.DBPort,
			DBName:   req.Params.DBName,
			Username: req.Params.DBUser,
			Password: req.Params.DBPassword,
		},
	}

	if err := setupservice.SaveConfig(config); err != nil {
		return component.FailResponse("保存配置失败: " + err.Error())
	}

	// 2. 初始化数据库
	if err := setupservice.InitDatabase(config.Database); err != nil {
		return component.FailResponse("初始化数据库失败: " + err.Error())
	}

	// 3. 创建管理员账号
	if err := setupservice.CreateAdminUser(
		req.Params.AdminUsername,
		req.Params.AdminPassword,
		req.Params.AdminEmail,
	); err != nil {
		return component.FailResponse("创建管理员账号失败: " + err.Error())
	}

	return component.SuccessResponse("初始化成功")
}
