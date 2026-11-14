package forum

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func Home(c *gin.Context) {
	sort, _ := lo.Coalesce(c.Query("sort"), "latest")
	page := lo.Ternary(cast.ToInt(c.Query("page")) <= 0, 1, cast.ToInt(c.Query("page")))

	snapshot, ok := requestAccessSnapshot(c)
	if !ok {
		return
	}
	// Capture before loading the list: a concurrent publication can then only
	// cause an extra refresh hint, never advance the baseline past unseen data.
	var publicMax *uint64
	if id, err := hotdataserve.PublicTopicMaxID(); err == nil {
		publicMax = &id
	}
	audience, cacheable := snapshot.ListCacheAudience()
	topicPage := hotdataserve.GetLatestTopicsSimpleVoPaginatedForAudience(
		page,
		sort,
		audience,
		snapshot.ReadableCategoryIDs(),
		!snapshot.HasGlobalManage(),
		cacheable,
	)
	props := buildHomeProps(component.LoginUserId(c), page, sort, topicPage.Topics, topicPage.HasNext)
	props.PublicMaxTopicID = publicMax
	payload := PagePayload{
		Component: PageComponentHome,
		Props:     props,
		Meta:      buildHomeMeta(c, page, sort, topicPage.HasNext),
		Layout:    buildLayout(c, activeKeyForHome(sort)),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}

	renderPage(c, "home.gohtml", payload)
}
