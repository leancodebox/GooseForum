package routes

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	oidchttp "github.com/leancodebox/GooseForum/app/http/controllers/oidcprovider"
	"github.com/leancodebox/GooseForum/app/http/middleware"
	"github.com/leancodebox/GooseForum/app/service/oidcproviderservice"
)

// RegisterOIDCProvider registers protocol endpoints after the application has
// constructed its provider runtime. These responses intentionally bypass the
// GooseForum API envelope.
func RegisterOIDCProvider(engine *gin.Engine, handler *oidchttp.Handler) {
	group := engine.Group("/oauth2")
	group.GET("/.well-known/openid-configuration", handler.Discovery)
	group.GET("/authorize", middleware.JWTAuth, handler.Authorize)
	group.POST("/authorize", middleware.JWTAuth, handler.Authorize)
	group.GET("/authorize/resume", middleware.JWTAuth, handler.ResumeAuthorization)
	group.GET("/consent/details", middleware.JWTAuth, handler.ConsentDetails)
	group.POST("/consent", middleware.JWTAuth, handler.ConsentDecision)
	group.POST("/token", handler.Token)
	group.GET("/userinfo", handler.UserInfo)
	group.POST("/userinfo", handler.UserInfo)
	group.GET("/jwks.json", handler.JWKS)
	group.POST("/revoke", handler.Revoke)
}

func registerDefaultOIDCProvider(engine *gin.Engine) {
	handler, err := oidchttp.New(oidcproviderservice.DefaultProvider, func(c *gin.Context) core.Authentication {
		userID := c.GetUint64("userId")
		authTime := c.GetInt64("authTime")
		if userID == 0 {
			return core.Authentication{}
		}
		auth := core.Authentication{UserID: fmt.Sprint(userID), AuthID: c.GetString("authId")}
		if authTime > 0 {
			auth.AuthTime = time.Unix(authTime, 0)
		}
		return auth
	})
	if err != nil {
		panic(err)
	}
	handler.WithInteractions(oidcproviderservice.DefaultInteractions{}, time.Now, nil)
	RegisterOIDCProvider(engine, handler)
}
