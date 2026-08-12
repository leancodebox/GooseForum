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

// cachedAccessSnapshot returns the snapshot the handler already resolved for this
// request. It never resolves nor aborts, so payload builders can stay side-effect
// free; when no handler resolved one it falls back to the empty snapshot, which
// reads as "nothing is readable" and keeps the caller fail-closed.
func cachedAccessSnapshot(c *gin.Context) accesscontrol.Snapshot {
	if value, ok := c.Get(accessSnapshotContextKey); ok {
		if snapshot, valid := value.(accesscontrol.Snapshot); valid {
			return snapshot
		}
	}
	return accesscontrol.Snapshot{}
}

func redirectGuestToLogin(c *gin.Context) {
	c.Redirect(http.StatusFound, "/login?redirect="+url.QueryEscape(c.Request.URL.String()))
}
