package forum

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
)

func Sponsors(c *gin.Context) {
	payload := PagePayload{
		Component: "sponsors.index",
		Props:     buildSponsorsPageProps(hotdataserve.SponsorsConfigCache()),
		Meta:      buildSponsorsMeta(c),
		Layout:    buildLayout(c, "sponsors"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "sponsors.gohtml", payload)
}
