package urlconfig

import (
	"fmt"
	"path"
)

func GetDefaultAvatar() string {
	return `/static/pic/default-avatar.webp`
}

func FilePath(filename string) string {
	return path.Join("/file/img", filename)
}

// 页面路径定义
const (
	PathHome     = "/"
	PathPost     = "/post"
	PathDocs     = "/docs"
	PathLinks    = "/links"
	PathSponsors = "/sponsors"
	PathAbout    = "/about"
	PathPublish  = "/publish"
	PathSearch   = "/search"
	PathRegister = "/register"
	PathLogin    = "/login"
	PathRss      = "/rss.xml"
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

func Rss() string {
	return PathRss
}

func PostDetail(id any) string {
	return fmt.Sprintf("%s/%v", PathPost, id)
}

func User(id any) string {
	return fmt.Sprintf("/user/%v", id)
}

func DocsProject(slug any) string {
	return fmt.Sprintf("%s/%v", PathDocs, slug)
}

func DocsContent(projectSlug, versionSlug, contentSlug any) string {
	return fmt.Sprintf("%s/%v/%v/%v", PathDocs, projectSlug, versionSlug, contentSlug)
}
