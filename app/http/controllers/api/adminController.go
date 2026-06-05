package api

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/buildinfo"
	"github.com/leancodebox/GooseForum/app/bundles/randopt"
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/badges"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/badgeservice"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
	"github.com/samber/lo"
)

type TrafficOverviewReq struct {
	StartDate string `json:"startDate"` // YYYY-MM-DD
	EndDate   string `json:"endDate"`   // YYYY-MM-DD
}

type DailyTraffic struct {
	Date         string `json:"date"`
	RegCount     int64  `json:"regCount"`
	ArticleCount int64  `json:"articleCount"`
	ReplyCount   int64  `json:"replyCount"`
}

func GetTrafficOverview(req component.BetterRequest[TrafficOverviewReq]) component.Response {
	startDate := req.Params.StartDate
	endDate := req.Params.EndDate

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	keys := []dailyStats.StatType{
		dailyStats.StatTypeRegCount,
		dailyStats.StatTypeArticleCount,
		dailyStats.StatTypeReplyCount,
	}

	stats, err := dailyStats.GetStatsInRange(keys, startDate, endDate)
	if err != nil {
		return component.FailResponseCode(component.MessageAdminStatsFetchFailed, nil)
	}

	// 按日期分组
	dailyMap := make(map[string]*DailyTraffic)

	// 初始化日期范围内的每一天
	curr, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	for !curr.After(end) {
		d := curr.Format("2006-01-02")
		dailyMap[d] = &DailyTraffic{Date: d}
		curr = curr.AddDate(0, 0, 1)
	}

	for _, s := range stats {
		dateStr := s.StatDate.Format("2006-01-02")
		if item, ok := dailyMap[dateStr]; ok {
			switch dailyStats.StatType(s.StatKey) {
			case dailyStats.StatTypeRegCount:
				item.RegCount = s.StatValue
			case dailyStats.StatTypeArticleCount:
				item.ArticleCount = s.StatValue
			case dailyStats.StatTypeReplyCount:
				item.ReplyCount = s.StatValue
			}
		}
	}

	// 转换为数组并排序
	var result []*DailyTraffic
	curr, _ = time.Parse("2006-01-02", startDate)
	for !curr.After(end) {
		d := curr.Format("2006-01-02")
		result = append(result, dailyMap[d])
		curr = curr.AddDate(0, 0, 1)
	}

	return component.SuccessResponse(result)
}

type UserListReq struct {
	Username string `json:"username"`
	UserId   uint64 `json:"userId"`
	Email    string `json:"email"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type UserItem struct {
	UserId         uint64                              `json:"userId"`
	Username       string                              `json:"username"`
	AvatarUrl      string                              `json:"avatarUrl"`
	Email          string                              `json:"email"`
	Status         int8                                `json:"status"`
	Validate       int8                                `json:"validate"`
	Prestige       int64                               `json:"prestige"`
	RoleList       []datastruct.Option[string, uint64] `json:"roleList"`
	RoleId         uint64                              `json:"roleId,omitempty"`
	CreateTime     string                              `json:"createTime"`
	LastActiveTime string                              `json:"lastActiveTime"`
	Badges         []badgeservice.UserBadge            `json:"badges"`
}

func UserList(req component.BetterRequest[UserListReq]) component.Response {
	var pageData = users.Page(users.PageQuery{
		Page:     req.Params.Page,
		PageSize: req.Params.PageSize,
		Username: req.Params.Username,
		UserId:   req.Params.UserId,
		Email:    req.Params.Email,
	})

	userIds := lo.Map(pageData.Data, func(item users.EntityComplete, _ int) uint64 {
		return item.Id
	})
	usList := userStatistics.GetByUserIds(userIds)
	usMap := lo.KeyBy(usList, func(v *userStatistics.Entity) uint64 {
		return v.UserId
	})
	roleEntityList := role.AllEffective()
	roleMap := lo.KeyBy(roleEntityList, func(v *role.Entity) uint64 {
		return v.Id
	})
	list := lo.Map(pageData.Data, func(t users.EntityComplete, _ int) UserItem {
		var roleList []datastruct.Option[string, uint64]
		if roleEntity, ok := roleMap[t.RoleId]; ok {
			roleList = append(roleList, datastruct.Option[string, uint64]{
				Name:  roleEntity.RoleName,
				Value: roleEntity.Id,
			})
		}
		LastActiveTime := t.CreatedAt.Format(time.DateTime)
		if usItem, ok := usMap[t.Id]; ok {
			LastActiveTime = usItem.LastActiveTime.Format(time.DateTime)
		}
		return UserItem{
			UserId:         t.Id,
			AvatarUrl:      t.GetWebAvatarUrl(),
			Username:       t.Username,
			Email:          t.Email,
			Status:         t.IsFrozen,
			Validate:       t.IsActivated,
			Prestige:       t.Prestige,
			RoleList:       roleList,
			RoleId:         t.RoleId,
			CreateTime:     t.CreatedAt.Format(time.DateTime),
			LastActiveTime: LastActiveTime,
			Badges:         badgeservice.GetUserBadges(t.Id),
		}
	})
	return component.SuccessPage(
		list,
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type BadgeSaveReq struct {
	Code        string `json:"code"`
	Type        string `json:"type"`
	GrantMode   string `json:"grantMode"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconType    string `json:"iconType"`
	IconKey     string `json:"iconKey"`
	IconURL     string `json:"iconUrl"`
	Color       string `json:"color"`
	Level       string `json:"level"`
	IsEnabled   bool   `json:"isEnabled"`
	SortOrder   int    `json:"sortOrder"`
}

func BadgeList(req component.BetterRequest[component.Null]) component.Response {
	return component.SuccessResponse(badgeservice.AllForAdmin())
}

func generateCustomBadgeCode() string {
	for i := 0; i < 16; i++ {
		code := "custom_" + strings.ToLower(randopt.RandomString(10))
		if badges.GetByCode(code).Id == 0 {
			return code
		}
	}
	return fmt.Sprintf("custom_%x", time.Now().UnixNano())
}

func SaveBadge(req component.BetterRequest[BadgeSaveReq]) component.Response {
	params := req.Params
	params.Code = strings.TrimSpace(params.Code)
	params.Name = strings.TrimSpace(params.Name)
	if params.Name == "" {
		return component.FailResponseCode(component.MessageAdminBadgeNameRequired, nil)
	}
	if params.Type == "" {
		params.Type = badges.TypeCustom
	}
	if params.Type != badges.TypeSystem && params.Type != badges.TypeCustom {
		return component.FailResponseCode(component.MessageAdminBadgeTypeInvalid, nil)
	}
	if params.Code == "" {
		if params.Type == badges.TypeSystem {
			return component.FailResponseCode(component.MessageAdminBadgeCodeRequired, nil)
		}
		params.Code = generateCustomBadgeCode()
	}
	if params.GrantMode == "" {
		params.GrantMode = badges.GrantModeManual
	}
	if params.GrantMode != badges.GrantModeAuto && params.GrantMode != badges.GrantModeManual {
		return component.FailResponseCode(component.MessageAdminBadgeGrantModeInvalid, nil)
	}
	if params.IconType == "" {
		params.IconType = badges.IconTypeAsset
	}
	if params.Type == badges.TypeSystem {
		systemBadge := badgeservice.ResolveOne(params.Code)
		if systemBadge.Code == "" || systemBadge.Type != badges.TypeSystem {
			return component.FailResponseCode(component.MessageAdminBadgeSystemNotFound, nil)
		}
		params.GrantMode = systemBadge.GrantMode
	}

	entity := badges.GetByCode(params.Code)
	if entity.Id == 0 {
		entity.Code = params.Code
	}
	entity.Type = params.Type
	entity.GrantMode = params.GrantMode
	entity.Name = params.Name
	entity.Description = params.Description
	entity.IconType = params.IconType
	entity.IconKey = params.IconKey
	entity.IconURL = params.IconURL
	entity.Color = params.Color
	entity.Level = params.Level
	entity.IsEnabled = params.IsEnabled
	entity.SortOrder = params.SortOrder
	if err := badges.Save(&entity); err != nil {
		return component.FailResponseCode(component.MessageAdminBadgeSaveFailed, nil)
	}
	badgeservice.InvalidateDefinitions()
	userservice.ClearUserPublicProfileCache()
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}

type BadgeDeleteReq struct {
	Code string `json:"code"`
}

func DeleteBadge(req component.BetterRequest[BadgeDeleteReq]) component.Response {
	code := strings.TrimSpace(req.Params.Code)
	if code == "" {
		return component.FailResponseCode(component.MessageAdminBadgeCodeRequired, nil)
	}
	badge := badgeservice.ResolveOne(code)
	if badge.Type == badges.TypeSystem {
		return component.FailResponseCode(component.MessageAdminBadgeSystemDeleteBlock, nil)
	}
	if err := badges.DeleteByCode(code); err != nil {
		return component.FailResponseCode(component.MessageAdminBadgeDeleteFailed, nil)
	}
	badgeservice.InvalidateDefinitions()
	userservice.ClearUserPublicProfileCache()
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}

type UserBadgeOptionsReq struct {
	UserId uint64 `json:"userId"`
}

type UserBadgeOptionsResp struct {
	Options []badgeservice.Badge     `json:"options"`
	Active  []badgeservice.UserBadge `json:"active"`
}

func UserBadgeOptions(req component.BetterRequest[UserBadgeOptionsReq]) component.Response {
	return component.SuccessResponse(UserBadgeOptionsResp{
		Options: badgeservice.ManualGrantBadgesForAdmin(),
		Active:  badgeservice.GetUserBadges(req.Params.UserId),
	})
}

type SaveUserBadgesReq struct {
	UserId     uint64   `json:"userId"`
	BadgeCodes []string `json:"badgeCodes"`
}

func SaveUserBadges(req component.BetterRequest[SaveUserBadgesReq]) component.Response {
	userID := req.Params.UserId
	if userID == 0 {
		return component.FailResponseCode(component.MessageUserNotFound, nil)
	}
	allowed := lo.KeyBy(badgeservice.ManualGrantBadgesForAdmin(), func(item badgeservice.Badge) string { return item.Code })
	nextCodes := lo.Uniq(req.Params.BadgeCodes)
	nextSet := map[string]bool{}
	for _, code := range nextCodes {
		if _, ok := allowed[code]; !ok {
			continue
		}
		nextSet[code] = true
		_, _ = badgeservice.Grant(userID, code, userBadges.SourceManual, "管理员手动授予", req.UserId)
	}
	for _, active := range badgeservice.GetUserBadges(userID) {
		if active.Source != userBadges.SourceManual {
			continue
		}
		if !nextSet[active.Code] {
			_ = userBadges.Revoke(userID, active.Code)
		}
	}
	userservice.InvalidateUserPublicProfileCache(userID)
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}

type EditUserReq struct {
	UserId   uint64 `json:"userId"`
	Status   int8   `json:"status"`
	Validate int8   `json:"validate"`
	RoleId   uint64 `json:"roleId"`
}

func EditUser(req component.BetterRequest[EditUserReq]) component.Response {
	params := req.Params
	user, err := users.Get(params.UserId)
	if err != nil || user.Id == 0 {
		return component.FailResponseCode(component.MessageAdminTargetUserFetchFailed, nil)
	}
	opt := false
	msg := "用户编辑"
	if user.IsFrozen != params.Status {
		msg = msg + fmt.Sprintf("[用户状态调整:%v->%v]", user.IsFrozen, params.Status)
		user.IsFrozen = params.Status
		opt = true
	}
	if user.IsActivated != params.Validate {
		msg = msg + fmt.Sprintf("[用户验证状态:%v->%v]", user.IsActivated, params.Validate)
		user.IsActivated = params.Validate
		opt = true
	}
	if user.RoleId != params.RoleId {
		msg = msg + fmt.Sprintf("[用户角色调整:%v->%v]", user.RoleId, params.RoleId)
		user.RoleId = params.RoleId
		opt = true
	}
	if opt {
		if err := userservice.SaveUser(&user); err != nil {
			return component.FailResponseCode(component.MessageUserUpdateFailed, nil)
		}
		optlogger.UserOpt(req.UserId, optlogger.EditUser, user.Id, msg)
	}
	return component.SuccessResponseCode("success", component.MessageOperationSuccess, nil)
}

type ArticlesListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Search   string `form:"search"`
	UserId   uint64 `form:"userId"`
}

type ArticlesInfoVo struct {
	Id            uint64 `json:"id"`
	Title         string `json:"title"`
	Type          int8   `json:"type"`
	UserId        uint64 `json:"userId"`
	Username      string `json:"username"`
	ArticleStatus int8   `json:"articleStatus"`
	ProcessStatus int8   `json:"processStatus"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type ArticlesInfoAdminVo struct {
	Id            uint64   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Type          int8     `json:"type"`
	CategoryId    []uint64 `json:"categoryId"`
	UserId        uint64   `json:"userId"`
	Username      string   `json:"username"`
	UserAvatarUrl string   `json:"userAvatarUrl"`
	ArticleStatus int8     `json:"articleStatus"`
	ProcessStatus int8     `json:"processStatus"`
	ViewCount     uint64   `json:"viewCount"`
	ReplyCount    uint64   `json:"replyCount"`
	LikeCount     uint64   `json:"likeCount"`
	PinWeight     int      `json:"pinWeight"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

type ArticleSourceReq struct {
	Id uint64 `json:"id" validate:"required"`
}

type ArticleSourceVo struct {
	Id            uint64   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Content       string   `json:"content"`
	Type          int8     `json:"type"`
	CategoryId    []uint64 `json:"categoryId"`
	UserId        uint64   `json:"userId"`
	ArticleStatus int8     `json:"articleStatus"`
	ProcessStatus int8     `json:"processStatus"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

func ArticlesList(req component.BetterRequest[ArticlesListReq]) component.Response {
	param := req.Params
	pageData := articles.Page[articles.SmallEntity](articles.PageQuery{Page: max(param.Page, 1), PageSize: param.PageSize, UserId: param.UserId})
	userIds := lo.Map(pageData.Data, func(t articles.SmallEntity, _ int) uint64 {
		return t.UserId
	})
	userMap := users.GetMapByIds(userIds)
	return component.SuccessPage(
		lo.Map(pageData.Data, func(t articles.SmallEntity, _ int) ArticlesInfoAdminVo {
			username := ""
			userAvatarUrl := ""
			if user := userMap[t.UserId]; user != nil {
				username = user.Username
				userAvatarUrl = user.GetWebAvatarUrl()
			}
			return ArticlesInfoAdminVo{
				Id:            t.Id,
				Title:         t.Title,
				Description:   t.Description,
				Type:          t.Type,
				CategoryId:    t.CategoryId,
				UserId:        t.UserId,
				Username:      username,
				UserAvatarUrl: userAvatarUrl,
				ArticleStatus: t.ArticleStatus,
				ProcessStatus: t.ProcessStatus,
				ViewCount:     t.ViewCount,
				ReplyCount:    t.ReplyCount,
				LikeCount:     t.LikeCount,
				PinWeight:     t.PinWeight,
				CreatedAt:     t.CreatedAt.Format(time.DateTime),
				UpdatedAt:     t.UpdatedAt.Format(time.DateTime),
			}
		}),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

func ArticleSource(req component.BetterRequest[ArticleSourceReq]) component.Response {
	article := articles.Get(req.Params.Id)
	if article.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	return component.SuccessResponse(ArticleSourceVo{
		Id:            article.Id,
		Title:         article.Title,
		Description:   article.Description,
		Content:       article.Content,
		Type:          article.Type,
		CategoryId:    article.CategoryId,
		UserId:        article.UserId,
		ArticleStatus: article.ArticleStatus,
		ProcessStatus: article.ProcessStatus,
		CreatedAt:     article.CreatedAt.Format(time.DateTime),
		UpdatedAt:     article.UpdatedAt.Format(time.DateTime),
	})
}

type EditArticleReq struct {
	Id            uint64 `json:"id" validate:"required"`
	ProcessStatus int8   `json:"processStatus" validate:"oneof=0 1"` // 0正常 1封禁
}

type EditArticlePinReq struct {
	Id        uint64 `json:"id" validate:"required"`
	PinWeight int    `json:"pinWeight" validate:"min=0,max=1000000"`
}

type EditArticleCategoriesReq struct {
	Id         uint64   `json:"id" validate:"required"`
	CategoryId []uint64 `json:"categoryId" validate:"min=1,max=3"`
}

// EditArticle 文章状态管理
func EditArticle(req component.BetterRequest[EditArticleReq]) component.Response {
	article := articles.Get(req.Params.Id)
	if article.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	// 更新文章状态
	article.ProcessStatus = req.Params.ProcessStatus
	err := articles.Save(&article)
	if err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}

	// 记录操作日志
	status := "解除封禁"
	if req.Params.ProcessStatus == 1 {
		status = "封禁"
	}
	optlogger.UserOpt(req.UserId, optlogger.EditArticle, article.Id,
		fmt.Sprintf("文章%s操作:[%s]", status, article.Title))
	if _, err := searchservice.BuildSingleArticleSearchDocument(&article); err != nil {
		slog.Error("failed to rebuild article search document", "articleId", article.Id, "err", err)
	}
	return component.SuccessResponseCode("操作成功", component.MessageOperationSuccess, nil)
}

func EditArticlePin(req component.BetterRequest[EditArticlePinReq]) component.Response {
	article := articles.Get(req.Params.Id)
	if article.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}
	if article.PinWeight == req.Params.PinWeight {
		return component.SuccessResponseCode("操作成功", component.MessageOperationSuccess, nil)
	}
	oldPinWeight := article.PinWeight
	if err := articles.UpdatePinWeight(article.Id, req.Params.PinWeight); err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}
	hotdataserve.ClearArticleListCache()
	optlogger.UserOpt(req.UserId, optlogger.EditArticle, article.Id,
		fmt.Sprintf("文章置顶权重调整:[%s] %d -> %d", article.Title, oldPinWeight, req.Params.PinWeight))
	return component.SuccessResponseCode("操作成功", component.MessageOperationSuccess, nil)
}

// EditArticleCategories 文章分类管理
func EditArticleCategories(req component.BetterRequest[EditArticleCategoriesReq]) component.Response {
	categoryIds := lo.Uniq(req.Params.CategoryId)
	if len(categoryIds) == 0 {
		return component.FailResponseCode(component.MessageAdminCategorySelectRequired, nil)
	}
	if len(categoryIds) > 3 {
		return component.FailResponseCode(component.MessageAdminCategorySelectTooMany, nil)
	}
	for _, categoryId := range categoryIds {
		if categoryId == 0 || articleCategory.Get(categoryId).Id == 0 {
			return component.FailResponseCode(component.MessageAdminCategoryNotFound, nil)
		}
	}

	article := articles.Get(req.Params.Id)
	if article.Id == 0 {
		return component.FailResponseCode(component.MessageArticleNotFound, nil)
	}

	oldCategoryIds := append([]uint64(nil), article.CategoryId...)
	article.CategoryId = categoryIds
	if err := articles.Save(&article); err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}

	syncArticleCategoryRelations(article.Id, categoryIds)
	optlogger.UserOpt(req.UserId, optlogger.EditArticle, article.Id,
		fmt.Sprintf("文章分类调整:[%s] %v -> %v", article.Title, oldCategoryIds, categoryIds))
	if _, err := searchservice.BuildSingleArticleSearchDocument(&article); err != nil {
		slog.Error("failed to rebuild article search document", "articleId", article.Id, "err", err)
	}
	return component.SuccessResponseCode("操作成功", component.MessageOperationSuccess, nil)
}

func syncArticleCategoryRelations(articleId uint64, categoryIds []uint64) {
	categoryIDMap := lo.SliceToMap(categoryIds, func(id uint64) (uint64, bool) {
		return id, true
	})
	for _, item := range articleCategoryRs.GetByArticleId(articleId) {
		if _, ok := categoryIDMap[item.ArticleCategoryId]; ok {
			item.Effective = 1
			articleCategoryRs.SaveOrCreateById(item)
			delete(categoryIDMap, item.ArticleCategoryId)
		} else {
			item.Effective = 0
			articleCategoryRs.SaveOrCreateById(item)
		}
	}
	for id := range categoryIDMap {
		rs := &articleCategoryRs.Entity{
			ArticleId:         articleId,
			ArticleCategoryId: id,
			Effective:         1,
		}
		articleCategoryRs.SaveOrCreateById(rs)
	}
}

type PermissionListReq struct {
}

func GetPermissionList(req component.BetterRequest[PermissionListReq]) component.Response {
	res := permission.BuildOptions()
	return component.SuccessResponse(res)
}

type GetAllRoleItemReq struct {
}

func GetAllRoleItem(req component.BetterRequest[GetAllRoleItemReq]) component.Response {
	res := lo.Map(role.AllEffective(), func(t *role.Entity, _ int) datastruct.Option[string, uint64] {
		return datastruct.Option[string, uint64]{Name: t.RoleName, Label: t.RoleName, Value: t.Id}
	})
	return component.SuccessResponse(res)
}

type RoleListReq struct {
}

type RoleItem struct {
	RoleId      uint64           `json:"roleId"`
	RoleName    string           `json:"roleName"`
	Effective   int              `json:"effective"`
	Permissions []PermissionItem `json:"permissions"`
	CreateTime  string           `json:"createTime"`
}

type PermissionItem struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func RoleList(req component.BetterRequest[RoleListReq]) component.Response {
	pageData := role.Page(role.PageQuery{})
	roleIds := lo.Map(pageData.Data, func(t role.Entity, _ int) uint64 {
		return t.Id
	})
	rpGroup := make(map[uint64][]uint64)
	if len(roleIds) > 0 {
		rpGroup = rolePermissionRs.GetRsGroupByRoleIds(roleIds)
	}
	list := lo.Map(pageData.Data, func(t role.Entity, _ int) RoleItem {
		pList, ok := rpGroup[t.Id]
		permissionItemList := make([]PermissionItem, 0)
		if ok {
			permissionItemList = lo.Map(pList, func(t uint64, _ int) PermissionItem {
				p := permission.Enum(t)
				return PermissionItem{Id: p.Id(), Name: p.Name()}
			})
		}
		return RoleItem{
			RoleId:      t.Id,
			RoleName:    t.RoleName,
			Effective:   t.Effective,
			Permissions: permissionItemList,
			CreateTime:  t.CreatedAt.Format(time.DateTime),
		}
	})

	return component.SuccessPage(
		list,
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type RoleSaveReq struct {
	Id          uint     `json:"id"`
	RoleName    string   `json:"roleName" validate:"required"`
	Permissions []uint64 `json:"permissions" validate:"required,min=1,max=100"`
}

func RoleSave(req component.BetterRequest[RoleSaveReq]) component.Response {
	var roleEntity role.Entity
	if req.Params.Id > 0 {
		roleEntity = role.Get(req.Params.Id)
	} else {
		roleEntity = role.Entity{
			Effective: 1,
		}
	}
	roleEntity.RoleName = req.Params.RoleName
	if err := role.SaveOrCreateById(&roleEntity); err != nil {
		return component.FailResponseCode(component.MessageOperationFailed, nil)
	}

	rsList := rolePermissionRs.GetRsByRoleId(roleEntity.Id)
	canUpdateMap := lo.SliceToMap(req.Params.Permissions, func(id uint64) (uint64, bool) {
		return id, true
	})

	// 更新数据
	for _, item := range rsList {
		item.Effective = 0
		if _, ok := canUpdateMap[item.PermissionId]; ok {
			item.Effective = 1
			// 如果已经存在，从 map 中删除，避免重复插入
			delete(canUpdateMap, item.PermissionId)
		}
		rolePermissionRs.SaveOrCreateById(item)
	}
	// 插入新的条目
	for id := range canUpdateMap {
		rsItem := rolePermissionRs.Entity{
			RoleId:       roleEntity.Id,
			PermissionId: id,
			Effective:    1,
		}
		rolePermissionRs.SaveOrCreateById(&rsItem)
	}
	permission.InvalidateRole(roleEntity.Id)

	return component.SuccessResponse(true)
}

type RoleSaveDel struct {
	Id uint `json:"id"`
}

func RoleDel(req component.BetterRequest[RoleSaveDel]) component.Response {
	roleEntity := role.Get(req.Params.Id)
	if roleEntity.Id == 0 {
		return component.FailResponseCode(component.MessageAdminRoleNotFound, nil)
	}
	rsList := rolePermissionRs.GetRsByRoleId(roleEntity.Id)
	// 删除
	lo.ForEach(rsList, func(item *rolePermissionRs.Entity, _ int) {
		rolePermissionRs.DeleteEntity(item)
	})
	role.DeleteEntity(&roleEntity)
	permission.InvalidateRole(roleEntity.Id)
	return component.SuccessResponse(true)
}

type OptRecordPageReq struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	OptUserId  uint64 `json:"optUserId"`
	OptType    int    `json:"optType"`
	TargetType int    `json:"targetType"`
	TargetId   int    `json:"targetId"`
}

func OptRecordPage(req component.BetterRequest[OptRecordPageReq]) component.Response {
	pageData := optRecord.Page(optRecord.PageQuery{
		Page:       req.Params.Page,
		PageSize:   component.BoundPageSizeWithRange(req.Params.PageSize, 10, 50),
		OptUserId:  req.Params.OptUserId,
		OptType:    req.Params.OptType,
		TargetType: req.Params.TargetType,
		TargetId:   req.Params.TargetId,
	})
	return component.SuccessPage(
		lo.Map(pageData.Data, func(item optRecord.Entity, _ int) optRecord.Entity {
			return item
		}),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type CategoryListReq struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type CategoryItem struct {
	Id       uint64 `json:"id"`
	Category string `json:"category"`
	Desc     string `json:"desc"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Slug     string `json:"slug"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

// GetCategoryList 获取分类列表
func GetCategoryList(req component.BetterRequest[CategoryListReq]) component.Response {
	categories := articleCategory.All()
	return component.SuccessResponse(lo.Map(categories, func(t *articleCategory.Entity, _ int) CategoryItem {
		return CategoryItem{
			Id:       t.Id,
			Category: t.Category,
			Desc:     t.Desc,
			Icon:     t.Icon,
			Color:    t.Color,
			Slug:     t.Slug,
			//Sort:     t.Sort,
			//Status:   t.Status,
		}
	}))
}

type CategorySaveReq struct {
	Id       uint64 `json:"id"`
	Category string `json:"category" validate:"required"`
	Desc     string `json:"desc"`
	Icon     string `json:"icon"`
	Color    string `json:"color"`
	Slug     string `json:"slug"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

// SaveCategory 保存分类
func SaveCategory(req component.BetterRequest[CategorySaveReq]) component.Response {
	if len(strings.TrimSpace(req.Params.Category)) == 0 {
		return component.FailResponseCode(component.MessageAdminCategoryRequired, nil)
	}

	entity := articleCategory.Get(req.Params.Id)
	if req.Params.Id != 0 && entity.Id == 0 {
		return component.FailResponseCode(component.MessageAdminCategoryDataNotFound, nil)
	}
	entity.Category = req.Params.Category
	entity.Desc = req.Params.Desc
	entity.Icon = req.Params.Icon
	entity.Color = req.Params.Color
	entity.Slug = req.Params.Slug

	articleCategory.SaveOrCreateById(&entity)
	hotdataserve.ClearArticleCategoryCache()
	return component.SuccessResponse(true)
}

// DeleteCategory 删除分类
func DeleteCategory(req component.BetterRequest[struct {
	Id uint64 `json:"id"`
}]) component.Response {
	entity := articleCategory.Get(req.Params.Id)
	if entity.Id == 0 {
		return component.FailResponseCode(component.MessageAdminCategoryNotFound, nil)
	}
	if articleCategory.Count() == 1 {
		return component.FailResponseCode(component.MessageAdminCategoryKeepOne, nil)
	}
	if articleCategoryRs.GetOneByCategoryId(entity.Id).Id > 0 {
		return component.FailResponseCode(component.MessageAdminCategoryHasArticles, nil)
	}
	articleCategory.DeleteEntity(&entity)
	hotdataserve.ClearArticleCategoryCache()
	return component.SuccessResponse(true)
}

func GetFriendLinks(req component.BetterRequest[component.Null]) component.Response {
	res := pageConfig.GetConfigByPageType(pageConfig.FriendShipLinks, defaultconfig.GetDefaultFriendLinksConfig())
	normalizeFriendLinks(res)
	return component.SuccessResponse(res)
}

type SaveFriendLinksReq struct {
	LinksInfo []pageConfig.FriendLinksGroup `json:"linksInfo"`
}

// SaveFriendLinks 保存友情链接
func SaveFriendLinks(req component.BetterRequest[SaveFriendLinksReq]) component.Response {
	normalizeFriendLinks(req.Params.LinksInfo)
	return savePageConfig(pageConfig.FriendShipLinks, req.Params.LinksInfo, hotdataserve.ClearFriendLinksConfigCache)
}

func normalizeFriendLinks(groups []pageConfig.FriendLinksGroup) {
	for i := range groups {
		if groups[i].Links == nil {
			groups[i].Links = []pageConfig.LinkItem{}
		}
	}
}

// GetSponsors 获取赞助商配置
func GetSponsors(req component.BetterRequest[component.Null]) component.Response {
	res := pageConfig.GetConfigByPageType(pageConfig.SponsorsPage, defaultconfig.GetDefaultSponsorsConfig())
	fillSponsorsConfigDefaults(&res)
	return component.SuccessResponse(res)
}

type SaveSponsorsReq struct {
	SponsorsInfo pageConfig.SponsorsConfig `json:"sponsorsInfo"`
}

// SaveSponsors 保存赞助商配置
func SaveSponsors(req component.BetterRequest[SaveSponsorsReq]) component.Response {
	config := req.Params.SponsorsInfo
	fillSponsorsConfigDefaults(&config)
	return savePageConfig(pageConfig.SponsorsPage, config, hotdataserve.ClearSponsorsConfigCache)
}

func fillSponsorsConfigDefaults(config *pageConfig.SponsorsConfig) {
	defaultConfig := defaultconfig.GetDefaultSponsorsConfig()
	if config.Sponsors.Level0 == nil {
		config.Sponsors.Level0 = []pageConfig.SponsorItem{}
	}
	if config.Sponsors.Level1 == nil {
		config.Sponsors.Level1 = []pageConfig.SponsorItem{}
	}
	if config.Sponsors.Level2 == nil {
		config.Sponsors.Level2 = []pageConfig.SponsorItem{}
	}
	if config.Sponsors.Level3 == nil {
		config.Sponsors.Level3 = []pageConfig.SponsorItem{}
	}
	if config.Content.Title == "" {
		config.Content.Title = defaultConfig.Content.Title
	}
	if config.Content.Description == "" {
		config.Content.Description = defaultConfig.Content.Description
	}
	if config.Contact.Title == "" {
		config.Contact.Title = defaultConfig.Contact.Title
	}
	if config.Contact.Description == "" {
		config.Contact.Description = defaultConfig.Contact.Description
	}
	if config.Contact.ButtonText == "" {
		config.Contact.ButtonText = defaultConfig.Contact.ButtonText
	}
	if config.Contact.ButtonLink == "" {
		config.Contact.ButtonLink = defaultConfig.Contact.ButtonLink
	}
	if config.Rules == nil {
		config.Rules = defaultConfig.Rules
	}
}

// GetSiteSettings 获取站点设置
func GetSiteSettings(req component.BetterRequest[component.Null]) component.Response {
	defaultSettings := defaultconfig.GetDefaultSiteSettingsConfig()
	res := pageConfig.GetConfigByPageType(pageConfig.SiteSettings, defaultSettings)
	return component.SuccessResponse(res)
}

func ServerVersion(req component.BetterRequest[component.Null]) component.Response {
	return component.SuccessResponse(buildinfo.Get())
}

type SaveSiteSettingsReq struct {
	Settings pageConfig.SiteSettingsConfig `json:"settings"`
}

// SaveSiteSettings 保存站点设置
func SaveSiteSettings(req component.BetterRequest[SaveSiteSettingsReq]) component.Response {
	return savePageConfig(pageConfig.SiteSettings, req.Params.Settings, hotdataserve.ClearSiteSettingsConfigCache)
}

// GetMailSettings 获取邮件设置
func GetMailSettings(req component.BetterRequest[component.Null]) component.Response {
	// 获取当前站点设置
	defaultSettings := defaultconfig.GetDefaultEmailSettingsConfig()
	emailSettings := pageConfig.GetConfigByPageType(pageConfig.EmailSettings, defaultSettings)
	return component.SuccessResponse(emailSettings)
}

type SaveMailSettingsReq struct {
	Settings pageConfig.MailSettingsConfig `json:"settings" validate:"required"`
}

// SaveMailSettings 保存邮件设置
func SaveMailSettings(req component.BetterRequest[SaveMailSettingsReq]) component.Response {
	return savePageConfig(pageConfig.EmailSettings, req.Params.Settings, hotdataserve.ClearMailSettingsConfigCache)
}

type TestMailConnectionReq struct {
	Settings  pageConfig.MailSettingsConfig `json:"settings" validate:"required"`
	TestEmail string                        `json:"testEmail" validate:"required,email"`
}

type TestMailConnectionResp struct {
	Success     bool                    `json:"success"`
	MessageCode component.MessageCode   `json:"messageCode"`
	Params      component.MessageParams `json:"params,omitempty"`
}

// TestMailConnection 测试邮件连接
func TestMailConnection(req component.BetterRequest[TestMailConnectionReq]) component.Response {
	if req.Params.TestEmail == "" {
		return component.FailResponseCode(component.MessageAdminTestEmailRequired, nil)
	}

	err := mailservice.SendTestEmailWithConfig(req.Params.Settings, req.Params.TestEmail)
	if err != nil {
		errText := err.Error()
		return component.SuccessResponse(TestMailConnectionResp{
			Success:     false,
			MessageCode: component.MessageAdminTestEmailFailed,
			Params:      component.MessageParams{"error": errText},
		})
	}

	return component.SuccessResponse(TestMailConnectionResp{
		Success:     true,
		MessageCode: component.MessageAdminTestEmailSuccess,
		Params:      component.MessageParams{"email": req.Params.TestEmail},
	})
}

// GetAnnouncement 获取公告设置
func GetAnnouncement(req component.BetterRequest[component.Null]) component.Response {
	config := pageConfig.GetConfigByPageType(pageConfig.Announcement, defaultconfig.GetDefaultAnnouncementConfig())
	return component.SuccessResponse(config)
}

type SaveAnnouncementReq struct {
	Settings pageConfig.AnnouncementConfig `json:"settings" validate:"required"`
}

// SaveAnnouncement 保存公告设置
func SaveAnnouncement(req component.BetterRequest[SaveAnnouncementReq]) component.Response {
	return savePageConfig(pageConfig.Announcement, req.Params.Settings, hotdataserve.ClearAnnouncementConfigCache)
}

// GetSecuritySettings 获取安全与注册设置
func GetSecuritySettings(req component.BetterRequest[component.Null]) component.Response {
	defaultSettings := defaultconfig.GetDefaultSecuritySettingsConfig()
	res := pageConfig.GetConfigByPageType(pageConfig.SecuritySettings, defaultSettings)
	return component.SuccessResponse(res)
}

type SaveSecuritySettingsReq struct {
	Settings pageConfig.SecurityAndRegistration `json:"settings" validate:"required"`
}

// SaveSecuritySettings 保存安全与注册设置
func SaveSecuritySettings(req component.BetterRequest[SaveSecuritySettingsReq]) component.Response {
	return savePageConfig(pageConfig.SecuritySettings, req.Params.Settings, hotdataserve.ClearSecuritySettingsConfigCache)
}

// GetPostingSettings 获取发布内容设置
func GetPostingSettings(req component.BetterRequest[component.Null]) component.Response {
	defaultSettings := defaultconfig.GetDefaultPostingSettingsConfig()
	res := pageConfig.GetConfigByPageType(pageConfig.PostingSettings, defaultSettings)
	return component.SuccessResponse(res)
}

type SavePostingSettingsReq struct {
	Settings pageConfig.PostingContent `json:"settings" validate:"required"`
}

// SavePostingSettings 保存发布内容设置
func SavePostingSettings(req component.BetterRequest[SavePostingSettingsReq]) component.Response {
	return savePageConfig(pageConfig.PostingSettings, req.Params.Settings, hotdataserve.ClearPostingSettingsConfigCache)
}
