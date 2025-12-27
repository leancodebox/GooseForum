package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func HomeV3(c *gin.Context) {
	pageMeta := viewrender.NewPageMetaBuilder().
		SetTitle("GooseForum V3 - Discourse Style").
		SetDescription("A modern, table-based forum layout").
		SetCanonicalURL(component.BuildCanonicalHref(c)).
		Build()

	viewrender.SafeRender(c, "index_v3.gohtml", HomeData{
		ArticleCategoryList: hotdataserve.ArticleCategoryLabel(),
		LatestArticles:      hotdataserve.GetLatestArticleSimpleDto(),
		Stats:               hotdataserve.GetSiteStatisticsData(),
		RecommendedArticles: hotdataserve.GetRecommendedArticles(),
		Announcement:        hotdataserve.GetAnnouncementConfigCache(),
		GooseForumInfo:      GetGooseForumInfo(),
	}, pageMeta)
}
