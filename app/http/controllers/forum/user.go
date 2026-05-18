package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cast"
)

func UserProfile(c *gin.Context) {
	userID := cast.ToUint64(c.Param("userId"))
	user, err := users.Get(userID)
	if err != nil || user.Id == 0 {
		c.String(http.StatusNotFound, "用户不存在")
		return
	}

	props := buildUserProfileProps(c, user)
	payload := PagePayload{
		Component: "user.profile",
		Props:     props,
		Meta:      buildUserMeta(c, props.User),
		Layout:    buildLayout(c, "user"),
		URL:       buildPageURL(c),
		Version:   payloadVersion,
	}
	renderPage(c, "user.gohtml", payload)
}
