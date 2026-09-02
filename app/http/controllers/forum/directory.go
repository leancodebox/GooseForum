package forum

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/i18n"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/spf13/cast"
)

const memberDirectoryPageSize = 24

func Categories(c *gin.Context) {
	snapshot, ok := requestAccessSnapshot(c)
	if !ok {
		return
	}
	visible := make([]*category.Entity, 0)
	for _, item := range hotdataserve.GetCategory() {
		if item == nil || !snapshot.CanReadCategory(item.Id) {
			continue
		}
		visible = append(visible, item)
	}
	payload := PagePayload{
		Component: PageComponentCategories,
		Props:     buildCategoriesPageProps(visible),
		Meta:      buildDirectoryMeta(c, "meta.categories", "meta.categoriesDesc", "/categories", "", ""),
		Layout:    buildLayout(c, "categories"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "categories.gohtml", payload)
}

func Members(c *gin.Context) {
	beforeID := cast.ToUint64(c.Query("before"))
	afterID := cast.ToUint64(c.Query("after"))
	if beforeID > 0 {
		afterID = 0
	}
	result := users.PublicPage(users.PublicPageQuery{PageSize: memberDirectoryPageSize, BeforeID: beforeID, AfterID: afterID})
	userIDs := make([]uint64, 0, len(result.Data))
	for _, user := range result.Data {
		userIDs = append(userIDs, user.Id)
	}
	stats := map[uint64]*userStatistics.Entity{}
	if len(userIDs) > 0 {
		for _, item := range userStatistics.GetByUserIds(userIDs) {
			if item != nil {
				stats[item.UserId] = item
			}
		}
	}
	props := buildMembersPageProps(result, stats)
	payload := PagePayload{
		Component: PageComponentMembers,
		Props:     props,
		Meta:      buildDirectoryMeta(c, "meta.members", "meta.membersDesc", "/members", props.PreviousURL, props.Pagination.NextURL),
		Layout:    buildLayout(c, "members"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "members.gohtml", payload)
}

func buildCategoriesPageProps(items []*category.Entity) CategoriesPageProps {
	result := make([]CategoryDirectoryPayload, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, CategoryDirectoryPayload{
			ID: item.Id, Name: item.Name, Description: item.Desc, Icon: item.Icon,
			Color: item.Color, URL: categoryURL(item), TopicCount: int64(item.TopicCount),
		})
	}
	return CategoriesPageProps{Categories: result, Total: len(result)}
}

func buildMembersPageProps(result users.PublicPageResult, stats map[uint64]*userStatistics.Entity) MembersPageProps {
	members := make([]MemberDirectoryPayload, 0, len(result.Data))
	for i := range result.Data {
		user := &result.Data[i]
		member := MemberDirectoryPayload{
			ID: user.Id, Username: user.Username, Nickname: user.Nickname,
			AvatarURL: user.GetWebAvatarUrl(), Bio: user.Bio, Prestige: user.Prestige,
			JoinedAt: user.CreatedAt.Format("2006-01"), URL: "/u/" + strconv.FormatUint(user.Id, 10),
		}
		if member.Nickname == "" {
			member.Nickname = member.Username
		}
		if stat := stats[user.Id]; stat != nil {
			member.TopicCount = stat.TopicCount
			member.ReplyCount = stat.ReplyCount
			if !stat.LastActiveTime.IsZero() {
				member.LastActiveAt = stat.LastActiveTime.Format(time.DateTime)
			}
		}
		members = append(members, member)
	}
	previousURL := ""
	nextURL := ""
	if len(members) > 0 {
		if result.HasPrevious {
			previousURL = membersCursorURL("after", members[0].ID)
		}
		if result.HasNext {
			nextURL = membersCursorURL("before", members[len(members)-1].ID)
		}
	}
	return MembersPageProps{
		Members:     members,
		PreviousURL: previousURL,
		Pagination:  PaginationPayload{HasNext: result.HasNext, NextURL: nextURL},
	}
}

func membersCursorURL(direction string, id uint64) string {
	if id == 0 || (direction != "before" && direction != "after") {
		return ""
	}
	return "/members?" + direction + "=" + strconv.FormatUint(id, 10)
}

func buildDirectoryMeta(c *gin.Context, titleKey, descriptionKey, path, previousURL, nextURL string) PageMeta {
	lang := requestLang(c)
	meta := PageMeta{
		Title: pageTitle(i18n.T(lang, titleKey)), Description: i18n.T(lang, descriptionKey, "site", siteTitle()),
		Canonical: component.GetBaseUri(c) + path,
	}
	meta.PrevURL = previousURL
	meta.NextURL = nextURL
	return meta
}
