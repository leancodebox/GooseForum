package forum

type PageComponent string

const (
	PageComponentHome          PageComponent = "home.index"
	PageComponentTopic         PageComponent = "topic.detail"
	PageComponentUser          PageComponent = "user.profile"
	PageComponentCategory      PageComponent = "category.index"
	PageComponentLinks         PageComponent = "links.index"
	PageComponentSponsors      PageComponent = "sponsors.index"
	PageComponentNotifications PageComponent = "notifications.index"
	PageComponentMessages      PageComponent = "messages.index"
	PageComponentDrafts        PageComponent = "drafts.index"
	PageComponentModeration    PageComponent = "moderation.index"
	PageComponentSettings      PageComponent = "settings.index"
	PageComponentAccessGroups  PageComponent = "access-groups.index"
	PageComponentThemePreview  PageComponent = "theme.preview"
	PageComponentPublish       PageComponent = "publish.index"
	PageComponentSearch        PageComponent = "search.index"
	PageComponentLogin         PageComponent = "auth.login"
	PageComponentResetPassword PageComponent = "auth.resetPassword"
	PageComponentError         PageComponent = "error.index"
	PageComponentAdmin         PageComponent = "admin.shell"
)
