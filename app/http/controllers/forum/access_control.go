package forum

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

const accessSnapshotContextKey = "forumAccessSnapshot"

func requestAccessSnapshot(c *gin.Context) (accesscontrol.Snapshot, bool) {
	if value, ok := c.Get(accessSnapshotContextKey); ok {
		if snapshot, valid := value.(accesscontrol.Snapshot); valid {
			return snapshot, true
		}
	}
	snapshot, err := accesscontrol.Resolve(component.LoginUserId(c))
	if err != nil {
		slog.Error("resolve request access snapshot failed", "path", c.Request.URL.Path, "err", err)
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return accesscontrol.Snapshot{}, false
	}
	c.Set(accessSnapshotContextKey, snapshot)
	return snapshot, true
}

func redirectGuestToLogin(c *gin.Context) {
	c.Redirect(http.StatusFound, "/login?redirect="+url.QueryEscape(c.Request.URL.String()))
}
