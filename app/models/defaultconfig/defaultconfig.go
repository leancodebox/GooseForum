package defaultconfig

import "github.com/leancodebox/GooseForum/app/models/forum/pageConfig"

// Footer管理相关
var defaultFooter = pageConfig.FooterConfig{
	Primary: []pageConfig.PItem{
		{Content: `Power by <a href="https://github.com/leancodebox/GooseForum">GooseForum</a>.`},
		{Content: `Providing reliable tech since 2025`},
	},
	List: []pageConfig.FooterGroup{
		{
			Name: "Services",
			Children: []pageConfig.FooterItem{
				{Name: "Github", Url: "https://github.com/leancodebox/GooseForum"},
				{Name: "License", Url: "https://github.com/leancodebox/GooseForum/blob/main/LICENSE"},
				{Name: "LeanCodeBox", Url: "https://github.com/leancodebox"},
			},
		},
		{
			Name: "Legal",
			Children: []pageConfig.FooterItem{
				{Name: "用户协议", Url: "/terms-of-service"},
				{Name: "隐私政策", Url: "/privacy-policy"},
			},
		},
		{
			Name: "Team",
			Children: []pageConfig.FooterItem{
				{Name: "About", Url: "/about"},
			},
		},
	},
}

func GetDefaultFooter() pageConfig.FooterConfig {
	return defaultFooter
}

var defaultSiteSettingsConfig = pageConfig.SiteSettingsConfig{
	SiteName:        "GooseForum",
	SiteLogo:        "/static/pic/icon.webp",
	SiteDescription: "一个现代化的论坛系统",
	SiteKeywords:    "forum,discussion,community",
	SiteUrl:         "",
	SiteEmail:       "example@example.example",
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
