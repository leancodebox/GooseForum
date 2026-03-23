package defaultconfig

import "github.com/leancodebox/GooseForum/app/models/forum/pageConfig"

// Footer 管理相关
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

var defaultSecuritySettingsConfig = pageConfig.SecurityAndRegistration{
	EnableSignup:            true,
	EnableEmailVerification: false,
	MustApproveUsers:        false,
	MinPasswordLength:       6,
	InviteOnly:              false,
	Restrictions: struct {
		AllowedDomains        []string `json:"allowedDomains"`
		BlockedDomains        []string `json:"blockedDomains"`
		MaxRegistrationsPerIp int      `json:"maxRegistrationsPerIp"`
	}{
		AllowedDomains:        []string{},
		BlockedDomains:        []string{},
		MaxRegistrationsPerIp: 10,
	},
}

func GetDefaultSecuritySettingsConfig() pageConfig.SecurityAndRegistration {
	return defaultSecuritySettingsConfig
}

var defaultPostingSettingsConfig = pageConfig.PostingContent{
	TextControl: struct {
		MinPostLength       int  `json:"minPostLength"`
		MaxPostLength       int  `json:"maxPostLength"`
		MinTitleLength      int  `json:"minTitleLength"`
		MaxTitleLength      int  `json:"maxTitleLength"`
		AllowUppercasePosts bool `json:"allowUppercasePosts"`
	}{
		MinPostLength:       5,
		MaxPostLength:       50000,
		MinTitleLength:      5,
		MaxTitleLength:      100,
		AllowUppercasePosts: true,
	},
	UploadControl: struct {
		AllowAttachments      bool     `json:"allowAttachments"`
		AuthorizedExtensions  []string `json:"authorizedExtensions"`
		MaxAttachmentSizeKb   int      `json:"maxAttachmentSizeKb"`
		MaxAttachmentsPerPost int      `json:"maxAttachmentsPerPost"`
	}{
		AllowAttachments:      true,
		AuthorizedExtensions:  []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		MaxAttachmentSizeKb:   5120,
		MaxAttachmentsPerPost: 10,
	},
	EditControl: struct {
		EditingGracePeriod      int  `json:"editingGracePeriod"`
		PostEditTimeLimit       int  `json:"postEditTimeLimit"`
		AllowUsersToDeletePosts bool `json:"allowUsersToDeletePosts"`
	}{
		EditingGracePeriod:      300,
		PostEditTimeLimit:       86400,
		AllowUsersToDeletePosts: true,
	},
}

func GetDefaultPostingSettingsConfig() pageConfig.PostingContent {
	return defaultPostingSettingsConfig
}
