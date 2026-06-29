package badgeservice

import "github.com/leancodebox/GooseForum/app/models/forum/badges"

const (
	CodeFirstPost      = "first_post"
	CodeFirstComment   = "first_comment"
	CodeFirstLikeGiven = "first_like_given"
	CodeFirstFollower  = "first_follower"
	CodeWriter10       = "writer_10"
	CodeCommenter50    = "commenter_50"
	CodeLiked10        = "liked_10"
	CodePopular100     = "popular_100"
	CodeSocial10       = "social_10"
	CodeEarlyMember    = "early_member"
	CodeContributor    = "contributor"
	CodeModerator      = "moderator"
	CodeSponsor        = "sponsor"
	CodeKing           = "king"
	CodeRobot          = "robot"
)

const (
	LevelBronze  = "bronze"
	LevelSilver  = "silver"
	LevelGold    = "gold"
	LevelSpecial = "special"
)

type Definition struct {
	Code        string
	Type        string
	GrantMode   string
	Name        string
	Description string
	IconType    string
	IconKey     string
	IconURL     string
	Color       string
	Level       string
	IsEnabled   bool
	IsWearable  bool
	SortOrder   int
}

func systemDefinitions() []Definition {
	return []Definition{
		{Code: CodeFirstPost, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "初次发帖", Description: "发布了第一篇主题", IconType: badges.IconTypeAsset, IconURL: "/static/badges/first-post.svg", Color: "blue", Level: LevelBronze, IsEnabled: true, IsWearable: false, SortOrder: 10},
		{Code: CodeFirstComment, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "初次评论", Description: "留下了第一条评论或回复", IconType: badges.IconTypeAsset, IconURL: "/static/badges/first-comment.svg", Color: "teal", Level: LevelBronze, IsEnabled: true, IsWearable: false, SortOrder: 20},
		{Code: CodeFirstLikeGiven, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "友善点赞", Description: "第一次为他人的内容点赞", IconType: badges.IconTypeAsset, IconURL: "/static/badges/first-like-given.svg", Color: "rose", Level: LevelBronze, IsEnabled: true, IsWearable: false, SortOrder: 30},
		{Code: CodeFirstFollower, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "被看见了", Description: "获得了第一位粉丝", IconType: badges.IconTypeAsset, IconURL: "/static/badges/first-follower.svg", Color: "violet", Level: LevelBronze, IsEnabled: true, IsWearable: false, SortOrder: 40},
		{Code: CodeWriter10, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "持续创作", Description: "累计发布 10 篇主题", IconType: badges.IconTypeAsset, IconURL: "/static/badges/writer-10.svg", Color: "sky", Level: LevelSilver, IsEnabled: true, IsWearable: true, SortOrder: 50},
		{Code: CodeCommenter50, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "热心讨论", Description: "累计发布 50 条评论或回复", IconType: badges.IconTypeAsset, IconURL: "/static/badges/commenter-50.svg", Color: "emerald", Level: LevelSilver, IsEnabled: true, IsWearable: true, SortOrder: 60},
		{Code: CodeLiked10, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "受到认可", Description: "累计获得 10 个赞", IconType: badges.IconTypeAsset, IconURL: "/static/badges/liked-10.svg", Color: "amber", Level: LevelSilver, IsEnabled: true, IsWearable: true, SortOrder: 70},
		{Code: CodePopular100, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "社区之光", Description: "累计获得 100 个赞", IconType: badges.IconTypeAsset, IconURL: "/static/badges/popular-100.svg", Color: "orange", Level: LevelGold, IsEnabled: true, IsWearable: true, SortOrder: 80},
		{Code: CodeSocial10, Type: badges.TypeSystem, GrantMode: badges.GrantModeAuto, Name: "小有名气", Description: "累计获得 10 位粉丝", IconType: badges.IconTypeAsset, IconURL: "/static/badges/social-10.svg", Color: "purple", Level: LevelSilver, IsEnabled: true, IsWearable: true, SortOrder: 90},
		{Code: CodeEarlyMember, Type: badges.TypeSystem, GrantMode: badges.GrantModeManual, Name: "早期成员", Description: "社区早期加入者", IconType: badges.IconTypeAsset, IconURL: "/static/badges/early-member.svg", Color: "cyan", Level: LevelSpecial, IsEnabled: true, IsWearable: true, SortOrder: 100},
		{Code: CodeContributor, Type: badges.TypeSystem, GrantMode: badges.GrantModeManual, Name: "贡献者", Description: "为社区建设做出贡献", IconType: badges.IconTypeAsset, IconURL: "/static/badges/contributor.svg", Color: "fuchsia", Level: LevelSpecial, IsEnabled: true, IsWearable: true, SortOrder: 110},
		{Code: CodeModerator, Type: badges.TypeSystem, GrantMode: badges.GrantModeManual, Name: "社区维护者", Description: "协助维护社区秩序", IconType: badges.IconTypeAsset, IconURL: "/static/badges/moderator.svg", Color: "emerald", Level: LevelSpecial, IsEnabled: true, IsWearable: true, SortOrder: 120},
		{Code: CodeSponsor, Type: badges.TypeSystem, GrantMode: badges.GrantModeManual, Name: "赞助者", Description: "支持社区持续运行", IconType: badges.IconTypeAsset, IconURL: "/static/badges/sponsor.svg", Color: "yellow", Level: LevelSpecial, IsEnabled: true, IsWearable: true, SortOrder: 130},
		{Code: CodeKing, Type: badges.TypeSystem, GrantMode: badges.GrantModeManual, Name: "King", Description: "社区之王", IconType: badges.IconTypeAsset, IconURL: "/static/badges/king.svg", Color: "amber", Level: LevelSpecial, IsEnabled: true, IsWearable: true, SortOrder: 140},
		{Code: CodeRobot, Type: badges.TypeSystem, GrantMode: badges.GrantModeManual, Name: "机器人", Description: "你就是机器人！", IconType: badges.IconTypeAsset, IconURL: "/static/badges/robot.svg", Color: "slate", Level: LevelSpecial, IsEnabled: true, IsWearable: true, SortOrder: 150},
	}
}
