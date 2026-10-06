package forum

import (
	"testing"

	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
)

func TestSponsorTierLabelsFollowLocale(t *testing.T) {
	item := []pageConfig.SponsorItem{{Name: "Example", Message: "Original message"}}
	config := pageConfig.SponsorsConfig{
		Sponsors: pageConfig.Sponsors{Level0: item, Level1: item, Level2: item, Level3: item},
		Content:  pageConfig.SponsorsPageIntro{Title: "Original title"},
	}
	for lang, labels := range map[string][]string{
		"zh": {"钻石合作伙伴", "金牌赞助商", "银牌赞助商", "支持者"},
		"en": {"Diamond Partners", "Gold Sponsors", "Silver Sponsors", "Supporters"},
		"ja": {"ダイヤモンドパートナー", "ゴールドスポンサー", "シルバースポンサー", "サポーター"},
		"it": {"Partner diamante", "Sponsor oro", "Sponsor argento", "Sostenitori"},
	} {
		t.Run(lang, func(t *testing.T) {
			props := buildSponsorsPageProps(config, lang)
			if len(props.Sections) != 4 || props.TotalCount != 4 || props.Content.Title != config.Content.Title {
				t.Fatalf("unexpected sponsor payload: %+v", props)
			}
			for index, label := range labels {
				section := props.Sections[index]
				if section.Label != label || section.Sponsors[0].Message != item[0].Message {
					t.Fatalf("unexpected section: %+v", section)
				}
			}
		})
	}
}
