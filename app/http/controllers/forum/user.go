package forum

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/spf13/cast"
)

const (
	userProfileSectionSummary  = "summary"
	userProfileSectionActivity = "activity"
	userProfileSectionBadges   = "badges"

	userProfileActivityTimeline  = "timeline"
	userProfileActivityTopics    = "topics"
	userProfileActivityLikes     = "likes"
	userProfileActivityFollowing = "following"
	userProfileActivityFollowers = "followers"
)

func UserProfile(c *gin.Context) {
	userID := cast.ToUint64(c.Param("userId"))
	user, err := users.Get(userID)
	if err != nil || user.Id == 0 {
		RenderNotFoundPage(c, component.MessagePageNotFound)
		return
	}

	props := buildUserProfileProps(c, user, resolveUserProfileSection(c.Param("section")), resolveUserProfileActivitySection(c.Param("subsection")))
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

func resolveUserProfileSection(raw string) string {
	switch raw {
	case userProfileSectionActivity, userProfileSectionBadges:
		return raw
	default:
		return userProfileSectionSummary
	}
}

func resolveUserProfileActivitySection(raw string) string {
	switch raw {
	case userProfileActivityTopics, userProfileActivityLikes, userProfileActivityFollowing, userProfileActivityFollowers:
		return raw
	default:
		return userProfileActivityTimeline
	}
}
