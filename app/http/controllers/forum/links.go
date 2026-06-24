package forum

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func Links(c *gin.Context) {
	payload := PagePayload{
		Component: "links.index",
		Props:     buildLinksPageProps(hotdataserve.GetFriendLinksConfigCache()),
		Meta:      buildLinksMeta(c),
		Layout:    buildLayout(c, "links"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "links.gohtml", payload)
}
