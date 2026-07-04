package defaultconfig

import "testing"

func TestPageConfigDefaultsLoad(t *testing.T) {
	defaults, err := loadPageConfigDefaults()
	if err != nil {
		t.Fatalf("load page config defaults: %v", err)
	}
	if defaults.Site.SiteName != "GooseForum" {
		t.Fatalf("site name = %q, want GooseForum", defaults.Site.SiteName)
	}
	if len(defaults.FriendLinks) == 0 {
		t.Fatal("friend links defaults should not be empty")
	}
	if len(defaults.Sponsors.Rules) == 0 {
		t.Fatal("sponsor rules defaults should not be empty")
	}
	if defaults.Posting.UploadControl.MaxAttachmentSizeKb == 0 {
		t.Fatal("posting max attachment size should not be zero")
	}
}

func TestPageConfigDefaultGettersReturnCopies(t *testing.T) {
	links := GetDefaultFriendLinksConfig()
	links[0].Links[0].Name = "changed"
	if got := GetDefaultFriendLinksConfig()[0].Links[0].Name; got == "changed" {
		t.Fatal("friend links getter returned shared mutable data")
	}

	site := GetDefaultSiteSettingsConfig()
	site.SiteName = "changed"
	if got := GetDefaultSiteSettingsConfig().SiteName; got == "changed" {
		t.Fatal("site settings getter returned shared mutable data")
	}

	chrome := GetDefaultSiteChromeConfig()
	chrome.FooterInfo.List[0].Name = "changed"
	if got := GetDefaultSiteChromeConfig().FooterInfo.List[0].Name; got == "changed" {
		t.Fatal("site chrome getter returned shared mutable footer data")
	}

	sponsors := GetDefaultSponsorsConfig()
	sponsors.Rules[0].Content = "changed"
	if got := GetDefaultSponsorsConfig().Rules[0].Content; got == "changed" {
		t.Fatal("sponsors getter returned shared mutable rules")
	}
}
