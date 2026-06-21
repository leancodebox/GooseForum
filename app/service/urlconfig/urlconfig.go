// Package urlconfig centralizes public URL and route path helpers.
package urlconfig

import (
	"fmt"
	"path"

	"github.com/leancodebox/GooseForum/app/bundles/setting"
)

// GetDefaultAvatar returns the default avatar URL, using the CDN URL when configured.
func GetDefaultAvatar() string {
	cdnURL := setting.GetCDNURL()
	if cdnURL != "" {
		return cdnURL + `/static/pic/default-avatar.webp`
	}
	return `/static/pic/default-avatar.webp`
}

// GetBannedAvatar returns the fixed avatar shown for frozen accounts.
func GetBannedAvatar() string {
	cdnURL := setting.GetCDNURL()
	if cdnURL != "" {
		return cdnURL + `/static/pic/banned-avatar.png`
	}
	return `/static/pic/banned-avatar.png`
}

// FilePath returns the public image route for filename.
func FilePath(filename string) string {
	return path.Join("/file/img", filename)
}

// Public page route constants.
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
	PathDrafts        = "/drafts"
	PathSettings      = "/settings"
	PathNotifications = "/notifications"
	PathActivate      = "/activate"
	PathResetPassword = "/reset-password"
	PathAdmin         = "/admin"
)

// Home returns the home page path.
func Home() string { return PathHome }

// Post returns the post listing path.
func Post() string { return PathPost }

// Docs returns the documentation path.
func Docs() string { return PathDocs }

// Links returns the links page path.
func Links() string { return PathLinks }

// Sponsors returns the sponsors page path.
func Sponsors() string { return PathSponsors }

// About returns the about page path.
func About() string { return PathAbout }

// Publish returns the article publishing path.
func Publish() string { return PathPublish }

// Search returns the search page path.
func Search() string { return PathSearch }

// Register returns the registration path.
func Register() string { return PathRegister }

// Login returns the login path.
func Login() string { return PathLogin }

// Messages returns the messages page path.
func Messages() string { return PathMessages }

// Drafts returns the drafts page path.
func Drafts() string { return PathDrafts }

// Settings returns the settings page path.
func Settings() string { return PathSettings }

// Notifications returns the notifications page path.
func Notifications() string { return PathNotifications }

// Activate returns the account activation path.
func Activate() string { return PathActivate }

// ResetPassword returns the password reset path.
func ResetPassword() string { return PathResetPassword }

// Admin returns the admin console path.
func Admin() string { return PathAdmin }

// Rss returns the RSS feed path.
func Rss() string {
	return PathRss
}

// PostDetail returns the article detail path for id.
func PostDetail(id any) string {
	return fmt.Sprintf("%s/%v", PathPost, id)
}

// User returns the public user profile path for id.
func User(id any) string {
	return fmt.Sprintf("/u/%v", id)
}

// DocsProject returns the documentation project path for slug.
func DocsProject(slug any) string {
	return fmt.Sprintf("%s/%v", PathDocs, slug)
}

// DocsContent returns the documentation content path.
func DocsContent(projectSlug, versionSlug, contentSlug any) string {
	return fmt.Sprintf("%s/%v/%v/%v", PathDocs, projectSlug, versionSlug, contentSlug)
}
