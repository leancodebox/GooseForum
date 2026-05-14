package oauthservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/randopt"
	"github.com/leancodebox/GooseForum/app/bundles/sessionstore"
	"github.com/leancodebox/GooseForum/app/models/filemodel/filedata"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/userservice"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/samber/lo"
)

const (
	ProviderGitHub   = "github"
	ProviderGoogle   = "google"
	ProviderFacebook = "facebook"
	ProviderTwitter  = "twitter"
)

// InitOAuth configures available OAuth providers.
func InitOAuth() {
	gothic.Store = sessionstore.GetSession()

	var providers []goth.Provider

	if provider := initGitHubProvider(); provider != nil {
		providers = append(providers, provider)
	}

	if provider := initGoogleProvider(); provider != nil {
		providers = append(providers, provider)
	}

	if len(providers) > 0 {
		goth.UseProviders(providers...)
		slog.Info("OAuth提供商初始化完成", "count", len(providers))
	} else {
		slog.Warn("未配置任何OAuth提供商")
	}
}

// initGitHubProvider returns a GitHub provider when configured.
func initGitHubProvider() goth.Provider {
	clientID := preferences.GetString("github.client_id", "")
	clientSecret := preferences.GetString("github.client_secret", "")
	callbackURL := hotdataserve.GetSiteSettingsConfigCache().SiteUrl + "/api/auth/github/callback"
	if clientID == "" || clientSecret == "" {
		slog.Warn("GitHub OAuth配置缺失，跳过初始化")
		return nil
	}

	slog.Info("GitHub OAuth提供商初始化完成")
	return github.New(clientID, clientSecret, callbackURL)
}

// initGoogleProvider returns a Google provider when configured.
func initGoogleProvider() *google.Provider {
	clientID := preferences.GetString("google.client_id")
	clientSecret := preferences.GetString("google.client_secret")
	callbackURL := hotdataserve.GetSiteSettingsConfigCache().SiteUrl + "/api/auth/google/callback"
	if clientID != "" && clientSecret != "" && callbackURL != "" {
		// goth.UseProviders(googleProvider)
		slog.Info("Google OAuth provider configuration found (implementation pending)")
		return nil
	}
	return google.New(clientID, clientSecret, callbackURL)
}

// OAuthUserInfo is the normalized user data from an OAuth provider.
type OAuthUserInfo struct {
	ID        string `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
	Blog      string `json:"blog"`
	Location  string `json:"location"`
	Provider  string `json:"provider"`
}

// ProcessOAuthCallback logs in an existing OAuth user or creates a new one.
func ProcessOAuthCallback(gothUser goth.User) (*users.EntityComplete, error) {
	userInfo := parseOAuthUserInfo(gothUser)

	existingOAuth := userOAuth.GetByProviderAndUID(userInfo.Provider, userInfo.ID)
	if existingOAuth != nil {
		user, err := users.Get(existingOAuth.UserId)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %v", err)
		}
		updateOAuthRecord(existingOAuth, gothUser)
		return &user, nil
	}

	newUser, err := createUserFromOAuth(gothUser, userInfo)
	if err != nil {
		return nil, err
	}

	eventbus.Publish(context.Background(), &eventhandlers.UserSignUpEvent{
		UserId:   newUser.Id,
		Username: newUser.Username,
	})

	err = createOAuthRecord(newUser.Id, gothUser, userInfo)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// parseOAuthUserInfo normalizes provider-specific user data.
func parseOAuthUserInfo(gothUser goth.User) OAuthUserInfo {
	userInfo := OAuthUserInfo{
		ID:        gothUser.UserID,
		Login:     gothUser.NickName,
		Name:      gothUser.Name,
		Email:     gothUser.Email,
		AvatarURL: gothUser.AvatarURL,
		Provider:  gothUser.Provider,
	}

	if gothUser.RawData != nil {
		if bio, ok := gothUser.RawData["bio"].(string); ok {
			userInfo.Bio = bio
		}
		if blog, ok := gothUser.RawData["blog"].(string); ok {
			userInfo.Blog = blog
		}
		if location, ok := gothUser.RawData["location"].(string); ok {
			userInfo.Location = location
		}
		if login, ok := gothUser.RawData["login"].(string); ok && login != "" {
			userInfo.Login = login
		}
	}

	return userInfo
}

// createUserFromOAuth creates a local account from OAuth user data.
func createUserFromOAuth(gothUser goth.User, userInfo OAuthUserInfo) (*users.EntityComplete, error) {
	username := userInfo.Login
	originalUsername := username
	counter := 1
	for users.ExistUsername(username) {
		username = fmt.Sprintf("%s_%d", originalUsername, counter)
		counter++
	}

	userEntity, err := userservice.CreateUser(username, randopt.RandomString(32), "", false)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	if userInfo.AvatarURL != "" {
		localAvatarPath, err := downloadAndSaveAvatar(userEntity.Id, userInfo.AvatarURL)
		if err != nil {
			slog.Warn("下载头像失败，使用默认头像", "error", err, "avatarURL", userInfo.AvatarURL)
			userEntity.AvatarUrl = users.RandAvatarUrl()
		} else {
			userEntity.AvatarUrl = localAvatarPath
		}
	} else {
		userEntity.AvatarUrl = users.RandAvatarUrl()
	}

	userEntity.Nickname = username
	userEntity.Bio = userInfo.Bio
	userEntity.Website = userInfo.Blog
	users.Save(userEntity)

	return userEntity, nil
}

// createOAuthRecord stores a provider account binding.
func createOAuthRecord(userID uint64, gothUser goth.User, userInfo OAuthUserInfo) error {
	oauthEntity := &userOAuth.Entity{
		UserId:       userID,
		Provider:     userInfo.Provider,
		ProviderUid:  userInfo.ID,
		AccessToken:  gothUser.AccessToken,
		RefreshToken: gothUser.RefreshToken,
		Scopes:       gothUser.AccessTokenSecret,
		RawUserData:  "",
	}

	if !gothUser.ExpiresAt.IsZero() {
		oauthEntity.TokenExpiry = gothUser.ExpiresAt
	} else {
		oauthEntity.TokenExpiry = time.Now().AddDate(1, 0, 0)
	}

	return userOAuth.Create(oauthEntity)
}

// updateOAuthRecord refreshes stored provider token data.
func updateOAuthRecord(oauthEntity *userOAuth.Entity, gothUser goth.User) {
	oauthEntity.AccessToken = gothUser.AccessToken
	oauthEntity.RefreshToken = gothUser.RefreshToken
	oauthEntity.RawUserData = ""

	if !gothUser.ExpiresAt.IsZero() {
		oauthEntity.TokenExpiry = gothUser.ExpiresAt
	}

	userOAuth.Update(oauthEntity)
}

// GetOAuthByUserID returns a user's OAuth binding for a provider.
func GetOAuthByUserID(userID uint64, provider string) *userOAuth.Entity {
	return userOAuth.GetByUserIDAndProvider(userID, provider)
}

// UnbindOAuth removes one OAuth binding after safety checks.
func UnbindOAuth(userID uint64, provider string) error {
	oauthEntity := userOAuth.GetByUserIDAndProvider(userID, provider)
	if oauthEntity == nil {
		return errors.New("OAuth绑定不存在")
	}

	if err := checkUnbindSafety(userID, provider); err != nil {
		return err
	}

	return userOAuth.Delete(oauthEntity.Id)
}

// checkUnbindSafety ensures the user keeps at least one login method.
func checkUnbindSafety(userID uint64, providerToUnbind string) error {
	user, err := users.Get(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	hasEmail := user.Email != ""

	bindings := GetUserOAuthBindings(userID)

	remainingBindings := lo.CountBy(lo.Keys(bindings), func(p string) bool {
		return p != providerToUnbind
	})

	if !hasEmail && remainingBindings == 0 {
		return errors.New("解绑失败：您必须至少保留一种登录方式（邮箱或其他OAuth绑定）")
	}

	return nil
}

// ProcessOAuthBind binds a provider account to an existing user.
func ProcessOAuthBind(userID uint64, gothUser goth.User) error {
	userInfo := parseOAuthUserInfo(gothUser)

	existingOAuth := userOAuth.GetByProviderAndUID(userInfo.Provider, userInfo.ID)
	if existingOAuth != nil {
		if existingOAuth.UserId != userID {
			return errors.New("该OAuth账户已被其他用户绑定")
		}
		updateOAuthRecord(existingOAuth, gothUser)
		return nil
	}

	existingUserOAuth := userOAuth.GetByUserIDAndProvider(userID, userInfo.Provider)
	if existingUserOAuth != nil {
		return errors.New("您已绑定该平台账户")
	}

	return createOAuthRecord(userID, gothUser, userInfo)
}

// GetUserOAuthBindings returns active OAuth bindings keyed by provider.
func GetUserOAuthBindings(userID uint64) map[string]*userOAuth.Entity {
	providers := []string{ProviderGitHub, ProviderGoogle}
	return lo.PickBy(lo.Associate(providers, func(p string) (string, *userOAuth.Entity) {
		return p, userOAuth.GetByUserIDAndProvider(userID, p)
	}), func(_ string, v *userOAuth.Entity) bool {
		return v != nil
	})
}

// downloadAndSaveAvatar stores an external OAuth avatar locally.
func downloadAndSaveAvatar(userID uint64, avatarURL string) (string, error) {
	if avatarURL == "" {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", avatarURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载头像失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载头像失败，状态码: %d", resp.StatusCode)
	}

	const maxFileSize = 2 * 1024 * 1024
	limitedReader := io.LimitReader(resp.Body, maxFileSize+1)

	avatarData, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("读取头像数据失败: %v", err)
	}

	if len(avatarData) > maxFileSize {
		return "", fmt.Errorf("头像文件过大，最大允许2MB")
	}

	filename := "avatar"
	if urlPath := resp.Request.URL.Path; urlPath != "" {
		ext := path.Ext(urlPath)
		if ext != "" {
			filename = "avatar" + ext
		} else {
			contentType := resp.Header.Get("Content-Type")
			switch {
			case strings.Contains(contentType, "jpeg"):
				filename = "avatar.jpg"
			case strings.Contains(contentType, "png"):
				filename = "avatar.png"
			case strings.Contains(contentType, "gif"):
				filename = "avatar.gif"
			case strings.Contains(contentType, "webp"):
				filename = "avatar.webp"
			default:
				filename = "avatar.jpg"
			}
		}
	}

	fileEntity, err := filedata.SaveAvatar(userID, avatarData, filename)
	if err != nil {
		return "", fmt.Errorf("保存头像失败: %v", err)
	}

	return fileEntity.Name, nil
}
