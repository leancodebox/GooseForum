package controllers

import "github.com/leancodebox/GooseForum/app/http/controllers/component"

type ApplyShowReq struct {
	Title     string   `json:"comment"`
	Desc      string   `json:"desc"`
	ImageList []string `json:"imageList"`
}

// ApplyShow 申请展示,是侧边栏
func ApplyShow(req component.BetterRequest[ApplyShowReq]) component.Response {
	return component.SuccessResponse("success")
}

type ApplyTopReq struct {
	ArticleId uint64 `json:"articleId"`
}

// ApplyTop 置顶申请，考虑滚动
func ApplyTop(req component.BetterRequest[ApplyTopReq]) component.Response {
	return component.SuccessResponse("success")
}
