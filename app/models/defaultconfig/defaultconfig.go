package defaultconfig

import "github.com/leancodebox/GooseForum/app/models/forum/pageConfig"

var defaultSiteSettingsConfig = pageConfig.SiteSettingsConfig{
	SiteName:        "GooseForum",
	SiteLogo:        "/static/pic/icon.webp",
	SiteDescription: "一个现代化的论坛系统",
	SiteKeywords:    "forum,discussion,community",
	SiteUrl:         "",
	SiteEmail:       "example@example.example",
	FooterInfo: pageConfig.FooterInfo{
		Primary: []pageConfig.PItem{
			{Content: `Power by <a href="https://github.com/leancodebox/GooseForum">GooseForum</a>.`},
			{Content: `Providing reliable tech since 2025`},
		},
		List: []pageConfig.FooterItem{
			{Name: "Github", Url: "https://github.com/leancodebox/GooseForum"},
			{Name: "License", Url: "https://github.com/leancodebox/GooseForum/blob/main/LICENSE"},
			{Name: "LeanCodeBox", Url: "https://github.com/leancodebox"},
		},
	},
}

func GetDefaultSiteSettingsConfig() pageConfig.SiteSettingsConfig {
	return defaultSiteSettingsConfig
}

var defaultEmailSettingsConfig = pageConfig.MailSettingsConfig{
	EnableMail:   false,
	SmtpHost:     "",
	SmtpPort:     587,
	UseSSL:       false,
	SmtpUsername: "",
	SmtpPassword: "",
	FromName:     "GooseForum",
	FromEmail:    "",
}

func GetDefaultEmailSettingsConfig() pageConfig.MailSettingsConfig {
	return defaultEmailSettingsConfig
}

var defaultSecuritySettingsConfig = pageConfig.SecurityAndRegistration{
	EnableSignup:            true,
	EnableEmailVerification: false,
	AllowedDomains:          []string{},
}

func GetDefaultSecuritySettingsConfig() pageConfig.SecurityAndRegistration {
	return defaultSecuritySettingsConfig
}

var defaultPostingSettingsConfig = pageConfig.PostingContent{
	TextControl: struct {
		MinPostLength              int `json:"minPostLength"`
		MaxPostLength              int `json:"maxPostLength"`
		MinTitleLength             int `json:"minTitleLength"`
		MaxTitleLength             int `json:"maxTitleLength"`
		NewUserPostCooldownMinutes int `json:"newUserPostCooldownMinutes"`
	}{
		MinPostLength:              5,
		MaxPostLength:              50000,
		MinTitleLength:             5,
		MaxTitleLength:             100,
		NewUserPostCooldownMinutes: 0,
	},
	UploadControl: struct {
		AllowAttachments             bool     `json:"allowAttachments"`
		AuthorizedExtensions         []string `json:"authorizedExtensions"`
		MaxAttachmentSizeKb          int      `json:"maxAttachmentSizeKb"`
		MaxDailyUploadsPerUser       int      `json:"maxDailyUploadsPerUser"`
		NewUserUploadCooldownMinutes int      `json:"newUserUploadCooldownMinutes"`
	}{
		AllowAttachments:             true,
		AuthorizedExtensions:         []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		MaxAttachmentSizeKb:          5120,
		MaxDailyUploadsPerUser:       10,
		NewUserUploadCooldownMinutes: 1440,
	},
}

func GetDefaultPostingSettingsConfig() pageConfig.PostingContent {
	return defaultPostingSettingsConfig
}
