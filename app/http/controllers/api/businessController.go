package api

import (
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/applySheet"
)

type ApplyAddLinkReq struct {
	SiteName string `json:"siteName" validate:"required"`
	SiteUrl  string `json:"siteUrl" validate:"required"`
	SiteLogo string `json:"siteLogo" validate:"required"`
	SiteDesc string `json:"siteDesc" validate:"required"`
	Email    string `json:"email"`
	Contact  string `json:"contact"`
}

func ApplyAddLink(req component.BetterRequest[ApplyAddLinkReq]) component.Response {
	if applySheet.CantWriteNew(applySheet.ApplyAddLink, 33) {
		return component.FailResponse("今日网站已经收到很多申请，请明日再来提交")
	}
	entity := applySheet.Entity{
		UserId: req.UserId,
		ApplyUserInfo: jsonopt.Encode(map[string]any{
			"ip": "127.0.0.1",
		}),
		Type:    applySheet.ApplyAddLink,
		Title:   "友情链接申请",
		Content: jsonopt.Encode(req.Params),
	}
	applySheet.SaveOrCreateById(&entity)

	return component.SuccessResponse("")
}
