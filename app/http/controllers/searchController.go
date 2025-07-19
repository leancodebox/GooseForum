package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
)

// SearchArticlesRequest 搜索文章请求结构
type SearchArticlesRequest struct {
	Query      string   `json:"query" form:"query" validate:"required,min=1,max=100"` // 搜索关键词
	Categories []uint64 `json:"categories" form:"categories"`                         // 分类ID列表
	Page       int      `json:"page" form:"page"`                                     // 页码，从1开始
	PageSize   int      `json:"pageSize" form:"pageSize"`                             // 每页数量
}

// SearchArticles 搜索文章接口
// 通过名字和类别查询文章，先查Meilisearch索引，再查文章表获取具体数据
func SearchArticles(req component.BetterRequest[SearchArticlesRequest]) component.Response {
	// 参数验证
	if req.Params.Query == "" {
		return component.FailResponse("搜索关键词不能为空")
	}

	// 设置默认值
	if req.Params.Page <= 0 {
		req.Params.Page = 1
	}
	if req.Params.PageSize <= 0 {
		req.Params.PageSize = 20
	}
	if req.Params.PageSize > 100 {
		req.Params.PageSize = 100
	}

	// 计算偏移量
	offset := (req.Params.Page - 1) * req.Params.PageSize

	// 构建搜索请求
	searchReq := searchservice.SearchRequest{
		Query:      req.Params.Query,
		Categories: req.Params.Categories,
		Limit:      req.Params.PageSize,
		Offset:     offset,
	}

	// 执行搜索
	result, err := searchservice.SearchArticles(searchReq)
	if err != nil {
		return component.FailResponse("搜索失败: " + err.Error())
	}

	// 构建响应数据
	responseData := map[string]interface{}{
		"results":  result.Results,
		"total":    result.Total,
		"page":     req.Params.Page,
		"pageSize": req.Params.PageSize,
		"query":    req.Params.Query,
	}

	if len(req.Params.Categories) > 0 {
		responseData["categories"] = req.Params.Categories
	}

	return component.SuccessResponse(responseData)
}