package defaultconfig

import "github.com/leancodebox/GooseForum/app/models/forum/pageConfig"

// Footer管理相关
var defaultFooter = pageConfig.FooterConfig{
	HtmlList: []pageConfig.HtmlItem{
		{Item: `Power by <a href="https://github.com/leancodebox/GooseForum">GooseForum</a>.`},
		{Item: `Providing reliable tech since 2025`},
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
