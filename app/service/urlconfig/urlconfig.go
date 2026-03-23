package urlconfig

import (
	"fmt"
	"path"

	"github.com/leancodebox/GooseForum/app/bundles/setting"
)

func GetDefaultAvatar() string {
	cdnURL := setting.GetCDNURL()
	if cdnURL != "" {
		return cdnURL + `/static/pic/default-avatar.webp`
	}
	return `/static/pic/default-avatar.webp`
}

func FilePath(filename string) string {
	return path.Join("/file/img", filename)
}

// 页面路径定义
const (
	PathHome          = "/"
	PathPost          = "/p/post"
	PathDocs          = "/docs"
	PathLinks         = "/links"
	PathSponsors      = "/sponsors"
	PathAbout         = "/about"
	PathPublish       = "/publish"
	PathSearch        = "/search"
	PathRegister      = "/login"
	PathLogin         = "/login"
	PathRss           = "/rss.xml"
	PathMessages      = "/messages"
	PathSettings      = "/settings"
	PathNotifications = "/notifications"
	PathActivate      = "/activate"
	PathResetPassword = "/reset-password"
	PathAdmin         = "/admin/"
)

func Home() string { return PathHome }

func Post() string { return PathPost }

func Docs() string { return PathDocs }

func Links() string { return PathLinks }

func Sponsors() string { return PathSponsors }

func About() string { return PathAbout }

func Publish() string { return PathPublish }

func Search() string { return PathSearch }

func Register() string { return PathRegister }

func Login() string { return PathLogin }

func Messages() string { return PathMessages }

func Settings() string { return PathSettings }

func Notifications() string { return PathNotifications }

func Activate() string { return PathActivate }

func ResetPassword() string { return PathResetPassword }

func Admin() string { return PathAdmin }

func Rss() string {
	return PathRss
}

func PostDetail(id any) string {
	return fmt.Sprintf("%s/%v", PathPost, id)
}

func User(id any) string {
	return fmt.Sprintf("/u/%v", id)
}

func DocsProject(slug any) string {
	return fmt.Sprintf("%s/%v", PathDocs, slug)
}

func DocsContent(projectSlug, versionSlug, contentSlug any) string {
	return fmt.Sprintf("%s/%v/%v/%v", PathDocs, projectSlug, versionSlug, contentSlug)
}
