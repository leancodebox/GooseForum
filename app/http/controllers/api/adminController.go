package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/defaultconfig"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"
	"github.com/leancodebox/GooseForum/app/models/forum/articleCategoryRs"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/dailyStats"
	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/mailservice"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/searchservice"
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
		return component.FailResponse("获取统计数据失败")
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
		}
	})
	return component.SuccessPage(
		list,
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
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
		return component.FailResponse("目标用户查询失败")
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
		users.Save(&user)
		optlogger.UserOpt(req.UserId, optlogger.EditUser, user.Id, msg)
	}
	return component.SuccessResponse("success")
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
	Id            uint64 `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Type          int8   `json:"type"`
	UserId        uint64 `json:"userId"`
	Username      string `json:"username"`
	UserAvatarUrl string `json:"userAvatarUrl"`
	ArticleStatus int8   `json:"articleStatus"`
	ProcessStatus int8   `json:"processStatus"`
	ViewCount     uint64 `json:"viewCount"`
	ReplyCount    uint64 `json:"replyCount"`
	LikeCount     uint64 `json:"likeCount"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
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
			if user, _ := userMap[t.UserId]; user != nil {
				username = user.Username
				userAvatarUrl = user.GetWebAvatarUrl()
			}
			return ArticlesInfoAdminVo{
				Id:            t.Id,
				Title:         t.Title,
				Description:   t.Description,
				Type:          t.Type,
				UserId:        t.UserId,
				Username:      username,
				UserAvatarUrl: userAvatarUrl,
				ArticleStatus: t.ArticleStatus,
				ProcessStatus: t.ProcessStatus,
				ViewCount:     t.ViewCount,
				ReplyCount:    t.ReplyCount,
				LikeCount:     t.LikeCount,
				CreatedAt:     t.CreatedAt.Format(time.DateTime),
				UpdatedAt:     t.UpdatedAt.Format(time.DateTime),
			}
		}),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type EditArticleReq struct {
	Id            uint64 `json:"id" validate:"required"`
	ProcessStatus int8   `json:"processStatus" validate:"oneof=0 1"` // 0正常 1封禁
}

// EditArticle 文章状态管理
func EditArticle(req component.BetterRequest[EditArticleReq]) component.Response {
	article := articles.Get(req.Params.Id)
	if article.Id == 0 {
		return component.FailResponse("文章不存在")
	}

	// 更新文章状态
	article.ProcessStatus = req.Params.ProcessStatus
	err := articles.Save(&article)
	if err != nil {
		return component.FailResponse("操作失败")
	}

	// 记录操作日志
	status := "解除封禁"
	if req.Params.ProcessStatus == 1 {
		status = "封禁"
	}
	optlogger.UserOpt(req.UserId, optlogger.EditArticle, article.Id,
		fmt.Sprintf("文章%s操作:[%s]", status, article.Title))
	searchservice.BuildSingleArticleSearchDocument(&article)
	return component.SuccessResponse("操作成功")
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
	role.SaveOrCreateById(&roleEntity)

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

	return component.SuccessResponse(true)
}

type RoleSaveDel struct {
	Id uint `json:"id"`
}

func RoleDel(req component.BetterRequest[RoleSaveDel]) component.Response {
	roleEntity := role.Get(req.Params.Id)
	if roleEntity.Id == 0 {
		return component.FailResponse("角色不存在")
	}
	rsList := rolePermissionRs.GetRsByRoleId(roleEntity.Id)
	// 删除
	lo.ForEach(rsList, func(item *rolePermissionRs.Entity, _ int) {
		rolePermissionRs.DeleteEntity(item)
	})
	role.DeleteEntity(&roleEntity)
	return component.SuccessResponse(true)
}

type OptRecordPageReq struct {
}

func OptRecordPage(req component.BetterRequest[OptRecordPageReq]) component.Response {
	pageData := optRecord.Page(optRecord.PageQuery{})
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
		return component.FailResponse("分类名称不能为空")
	}

	entity := articleCategory.Get(req.Params.Id)
	if req.Params.Id != 0 && entity.Id == 0 {
		return component.FailResponse("数据不存在")
	}
	entity.Category = req.Params.Category
	entity.Desc = req.Params.Desc
	entity.Icon = req.Params.Icon
	entity.Color = req.Params.Color
	entity.Slug = req.Params.Slug

	articleCategory.SaveOrCreateById(&entity)
	return component.SuccessResponse(true)
}

// DeleteCategory 删除分类
func DeleteCategory(req component.BetterRequest[struct {
	Id uint64 `json:"id"`
}]) component.Response {
	entity := articleCategory.Get(req.Params.Id)
	if entity.Id == 0 {
		return component.FailResponse("分类不存在")
	}
	if articleCategory.Count() == 1 {
		return component.FailResponse("至少保留1个分类")
	}
	if articleCategoryRs.GetOneByCategoryId(entity.Id).Id > 0 {
		return component.FailResponse("当前分类存在有效文章")
	}
	articleCategory.DeleteEntity(&entity)
	return component.SuccessResponse(true)
}

func GetFriendLinks(req component.BetterRequest[component.Null]) component.Response {
	lItem := pageConfig.LinkItem{
		Name:    "GooseForum",
		Desc:    "简单的社区构建软件 / Easy forum software for building friendly communities.",
		Url:     "https://gooseforum.online",
		LogoUrl: "/static/pic/default-avatar.webp",
	}
	res := pageConfig.GetConfigByPageType(pageConfig.FriendShipLinks, []pageConfig.FriendLinksGroup{
		{
			Name:  "community",
			Emoji: "👥",
			Color: "#3b82f6",
			Links: []pageConfig.LinkItem{lItem},
		},
		{
			Name:  "blog",
			Emoji: "✍️",
			Color: "#22c55e",
			Links: []pageConfig.LinkItem{lItem},
		},
		{
			Name:  "tool",
			Emoji: "🛠️",
			Color: "#a855f7",
			Links: []pageConfig.LinkItem{lItem},
		},
	})
	return component.SuccessResponse(res)
}

type SaveFriendLinksReq struct {
	LinksInfo []pageConfig.FriendLinksGroup `json:"linksInfo"`
}

// SaveFriendLinks 保存友情链接
func SaveFriendLinks(req component.BetterRequest[SaveFriendLinksReq]) component.Response {
	return savePageConfig(pageConfig.FriendShipLinks, req.Params.LinksInfo, hotdataserve.ClearFriendLinksConfigCache)
}

// GetSponsors 获取赞助商配置
func GetSponsors(req component.BetterRequest[component.Null]) component.Response {
	defaultSponsors := pageConfig.SponsorsConfig{
		Sponsors: pageConfig.Sponsors{
			Level0: []pageConfig.SponsorItem{},
			Level1: []pageConfig.SponsorItem{},
			Level2: []pageConfig.SponsorItem{},
			Level3: []pageConfig.SponsorItem{},
		},
		Users: []pageConfig.UserSponsor{},
	}
	res := pageConfig.GetConfigByPageType(pageConfig.SponsorsPage, defaultSponsors)
	return component.SuccessResponse(res)
}

type SaveSponsorsReq struct {
	SponsorsInfo pageConfig.SponsorsConfig `json:"sponsorsInfo"`
}

// SaveSponsors 保存赞助商配置
func SaveSponsors(req component.BetterRequest[SaveSponsorsReq]) component.Response {
	return savePageConfig(pageConfig.SponsorsPage, req.Params.SponsorsInfo, hotdataserve.ClearSponsorsConfigCache)
}

// GetSiteSettings 获取站点设置
func GetSiteSettings(req component.BetterRequest[component.Null]) component.Response {
	defaultSettings := defaultconfig.GetDefaultSiteSettingsConfig()
	res := pageConfig.GetConfigByPageType(pageConfig.SiteSettings, defaultSettings)
	return component.SuccessResponse(res)
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
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestMailConnection 测试邮件连接
func TestMailConnection(req component.BetterRequest[TestMailConnectionReq]) component.Response {
	if req.Params.TestEmail == "" {
		return component.FailResponse("请输入测试邮箱地址")
	}

	err := mailservice.SendTestEmailWithConfig(req.Params.Settings, req.Params.TestEmail)
	if err != nil {
		return component.SuccessResponse(TestMailConnectionResp{
			Success: false,
			Message: fmt.Sprintf("邮件测试失败: %v", err),
		})
	}

	return component.SuccessResponse(TestMailConnectionResp{
		Success: true,
		Message: "邮件配置测试成功！测试邮件已发送到 " + req.Params.TestEmail,
	})
}

// GetAnnouncement 获取公告设置
func GetAnnouncement(req component.BetterRequest[component.Null]) component.Response {
	config := pageConfig.GetConfigByPageType(pageConfig.Announcement, pageConfig.AnnouncementConfig{})
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
