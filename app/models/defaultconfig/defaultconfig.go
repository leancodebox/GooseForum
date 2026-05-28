package defaultconfig

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

//go:embed pageconfig/*.json
var defaultConfigFS embed.FS

type pageConfigDefaults struct {
	Announcement pageConfig.AnnouncementConfig
	Email        pageConfig.MailSettingsConfig
	FriendLinks  []pageConfig.FriendLinksGroup
	Posting      pageConfig.PostingContent
	Security     pageConfig.SecurityAndRegistration
	Site         pageConfig.SiteSettingsConfig
	Sponsors     pageConfig.SponsorsConfig
}

var (
	loadPageConfigDefaultsOnce sync.Once
	pageConfigDefaultsValue    pageConfigDefaults
	pageConfigDefaultsErr      error
)

func loadPageConfigDefaults() (pageConfigDefaults, error) {
	loadPageConfigDefaultsOnce.Do(func() {
		pageConfigDefaultsErr = loadJSON("announcement.json", &pageConfigDefaultsValue.Announcement)
		if pageConfigDefaultsErr != nil {
			return
		}
		pageConfigDefaultsErr = loadJSON("email.json", &pageConfigDefaultsValue.Email)
		if pageConfigDefaultsErr != nil {
			return
		}
		pageConfigDefaultsErr = loadJSON("friend_links.json", &pageConfigDefaultsValue.FriendLinks)
		if pageConfigDefaultsErr != nil {
			return
		}
		pageConfigDefaultsErr = loadJSON("posting.json", &pageConfigDefaultsValue.Posting)
		if pageConfigDefaultsErr != nil {
			return
		}
		pageConfigDefaultsErr = loadJSON("security.json", &pageConfigDefaultsValue.Security)
		if pageConfigDefaultsErr != nil {
			return
		}
		pageConfigDefaultsErr = loadJSON("site.json", &pageConfigDefaultsValue.Site)
		if pageConfigDefaultsErr != nil {
			return
		}
		pageConfigDefaultsErr = loadJSON("sponsors.json", &pageConfigDefaultsValue.Sponsors)
	})
	return pageConfigDefaultsValue, pageConfigDefaultsErr
}

func mustPageConfigDefaults() pageConfigDefaults {
	defaults, err := loadPageConfigDefaults()
	if err != nil {
		panic(err)
	}
	return defaults
}

func loadJSON(name string, out any) error {
	data, err := defaultConfigFS.ReadFile("pageconfig/" + name)
	if err != nil {
		return fmt.Errorf("read default page config %s: %w", name, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode default page config %s: %w", name, err)
	}
	return nil
}

func GetDefaultAnnouncementConfig() pageConfig.AnnouncementConfig {
	return mustPageConfigDefaults().Announcement
}

func GetDefaultEmailSettingsConfig() pageConfig.MailSettingsConfig {
	return mustPageConfigDefaults().Email
}

func GetDefaultFriendLinksConfig() []pageConfig.FriendLinksGroup {
	return cloneFriendLinks(mustPageConfigDefaults().FriendLinks)
}

func GetDefaultPostingSettingsConfig() pageConfig.PostingContent {
	config := mustPageConfigDefaults().Posting
	config.UploadControl.AuthorizedExtensions = append([]string(nil), config.UploadControl.AuthorizedExtensions...)
	return config
}

func GetDefaultSecuritySettingsConfig() pageConfig.SecurityAndRegistration {
	config := mustPageConfigDefaults().Security
	config.AllowedDomains = append([]string(nil), config.AllowedDomains...)
	return config
}

func GetDefaultSiteSettingsConfig() pageConfig.SiteSettingsConfig {
	config := mustPageConfigDefaults().Site
	config.FooterInfo.Primary = append([]pageConfig.PItem(nil), config.FooterInfo.Primary...)
	config.FooterInfo.List = append([]pageConfig.FooterItem(nil), config.FooterInfo.List...)
	return config
}

func GetDefaultSponsorsConfig() pageConfig.SponsorsConfig {
	return cloneSponsorsConfig(mustPageConfigDefaults().Sponsors)
}

func cloneFriendLinks(groups []pageConfig.FriendLinksGroup) []pageConfig.FriendLinksGroup {
	if groups == nil {
		return nil
	}
	cloned := make([]pageConfig.FriendLinksGroup, len(groups))
	for i, group := range groups {
		cloned[i] = group
		cloned[i].Links = append([]pageConfig.LinkItem(nil), group.Links...)
	}
	return cloned
}

func cloneSponsorsConfig(config pageConfig.SponsorsConfig) pageConfig.SponsorsConfig {
	config.Sponsors.Level0 = append([]pageConfig.SponsorItem(nil), config.Sponsors.Level0...)
	config.Sponsors.Level1 = append([]pageConfig.SponsorItem(nil), config.Sponsors.Level1...)
	config.Sponsors.Level2 = append([]pageConfig.SponsorItem(nil), config.Sponsors.Level2...)
	config.Sponsors.Level3 = append([]pageConfig.SponsorItem(nil), config.Sponsors.Level3...)
	config.Rules = append([]pageConfig.SponsorsRule(nil), config.Rules...)
	return config
}
