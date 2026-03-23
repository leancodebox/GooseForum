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

// 支持的 OAuth 提供商
const (
	ProviderGitHub   = "github"
	ProviderGoogle   = "google"
	ProviderFacebook = "facebook"
	ProviderTwitter  = "twitter"
)

// InitOAuth 初始化OAuth配置
func InitOAuth() {
	// 设置goth的session store
	gothic.Store = sessionstore.GetSession()

	// 初始化所有配置的OAuth提供商
	var providers []goth.Provider

	// 配置GitHub OAuth
	if provider := initGitHubProvider(); provider != nil {
		providers = append(providers, provider)
	}

	// 初始化其他OAuth提供商
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

// initGitHubProvider 初始化GitHub OAuth提供商
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

// initGoogleProvider 初始化Google OAuth提供商
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

// OAuthUserInfo 通用OAuth用户信息结构
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

// ProcessOAuthCallback 处理OAuth回调
func ProcessOAuthCallback(gothUser goth.User) (*users.EntityComplete, error) {
	// 解析用户信息
	userInfo := parseOAuthUserInfo(gothUser)

	// 检查是否已存在OAuth记录
	existingOAuth := userOAuth.GetByProviderAndUID(userInfo.Provider, userInfo.ID)
	if existingOAuth != nil {
		// 已存在OAuth记录，直接登录
		user, err := users.Get(existingOAuth.UserId)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %v", err)
		}
		// 更新OAuth信息
		updateOAuthRecord(existingOAuth, gothUser)
		return &user, nil
	}

	// 创建新用户
	newUser, err := createUserFromOAuth(gothUser, userInfo)
	if err != nil {
		return nil, err
	}

	// 发布注册事件
	eventbus.Publish(context.Background(), &eventhandlers.UserSignUpEvent{
		UserId:   newUser.Id,
		Username: newUser.Username,
	})

	// 创建OAuth记录
	err = createOAuthRecord(newUser.Id, gothUser, userInfo)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// parseOAuthUserInfo 解析OAuth用户信息
func parseOAuthUserInfo(gothUser goth.User) OAuthUserInfo {
	userInfo := OAuthUserInfo{
		ID:        gothUser.UserID,
		Login:     gothUser.NickName,
		Name:      gothUser.Name,
		Email:     gothUser.Email,
		AvatarURL: gothUser.AvatarURL,
		Provider:  gothUser.Provider,
	}

	// 尝试从RawData解析更多信息
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
		// 对于GitHub，login字段可能在RawData中
		if login, ok := gothUser.RawData["login"].(string); ok && login != "" {
			userInfo.Login = login
		}
	}

	return userInfo
}

// createUserFromOAuth 从OAuth信息创建新用户
func createUserFromOAuth(gothUser goth.User, userInfo OAuthUserInfo) (*users.EntityComplete, error) {
	// 生成用户名（如果用户名已存在，添加后缀）
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

	// 下载并保存头像到本地
	if userInfo.AvatarURL != "" {
		localAvatarPath, err := downloadAndSaveAvatar(userEntity.Id, userInfo.AvatarURL)
		if err != nil {
			slog.Warn("下载头像失败，使用默认头像", "error", err, "avatarURL", userInfo.AvatarURL)
			// 如果下载失败，使用默认头像
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

// createOAuthRecord 创建OAuth记录
func createOAuthRecord(userID uint64, gothUser goth.User, userInfo OAuthUserInfo) error {
	oauthEntity := &userOAuth.Entity{
		UserId:       userID,
		Provider:     userInfo.Provider,
		ProviderUid:  userInfo.ID,
		AccessToken:  gothUser.AccessToken,
		RefreshToken: gothUser.RefreshToken,
		Scopes:       gothUser.AccessTokenSecret, // GitHub使用这个字段存储scope
		RawUserData:  "",                         // RawData暂时设为空字符串
	}

	// 设置token过期时间
	if !gothUser.ExpiresAt.IsZero() {
		oauthEntity.TokenExpiry = gothUser.ExpiresAt
	} else {
		// 如果没有过期时间，设置为1年后
		oauthEntity.TokenExpiry = time.Now().AddDate(1, 0, 0)
	}

	return userOAuth.Create(oauthEntity)
}

// updateOAuthRecord 更新OAuth记录
func updateOAuthRecord(oauthEntity *userOAuth.Entity, gothUser goth.User) {
	oauthEntity.AccessToken = gothUser.AccessToken
	oauthEntity.RefreshToken = gothUser.RefreshToken
	oauthEntity.RawUserData = "" // RawData暂时设为空字符串

	if !gothUser.ExpiresAt.IsZero() {
		oauthEntity.TokenExpiry = gothUser.ExpiresAt
	}

	userOAuth.Update(oauthEntity)
}

// GetOAuthByUserID 获取用户的OAuth绑定信息
func GetOAuthByUserID(userID uint64, provider string) *userOAuth.Entity {
	return userOAuth.GetByUserIDAndProvider(userID, provider)
}

// UnbindOAuth 解绑OAuth账户
func UnbindOAuth(userID uint64, provider string) error {
	oauthEntity := userOAuth.GetByUserIDAndProvider(userID, provider)
	if oauthEntity == nil {
		return errors.New("OAuth绑定不存在")
	}

	// 安全检查：确保用户解绑后仍有登录方式
	if err := checkUnbindSafety(userID, provider); err != nil {
		return err
	}

	return userOAuth.Delete(oauthEntity.Id)
}

// checkUnbindSafety 检查解绑安全性
func checkUnbindSafety(userID uint64, providerToUnbind string) error {
	// 获取用户信息
	user, err := users.Get(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 检查用户是否有邮箱
	hasEmail := user.Email != ""

	// 获取用户的所有OAuth绑定
	bindings := GetUserOAuthBindings(userID)

	// 计算解绑后剩余的OAuth绑定数量
	remainingBindings := lo.CountBy(lo.Keys(bindings), func(p string) bool {
		return p != providerToUnbind
	})

	// 如果用户既没有邮箱，解绑后也没有其他OAuth绑定，则禁止解绑
	if !hasEmail && remainingBindings == 0 {
		return errors.New("解绑失败：您必须至少保留一种登录方式（邮箱或其他OAuth绑定）")
	}

	return nil
}

// ProcessOAuthBind 处理OAuth绑定（用于已登录用户）
func ProcessOAuthBind(userID uint64, gothUser goth.User) error {
	// 解析用户信息
	userInfo := parseOAuthUserInfo(gothUser)

	// 检查该OAuth账户是否已被其他用户绑定
	existingOAuth := userOAuth.GetByProviderAndUID(userInfo.Provider, userInfo.ID)
	if existingOAuth != nil {
		if existingOAuth.UserId != userID {
			return errors.New("该OAuth账户已被其他用户绑定")
		}
		// 如果是同一用户，更新OAuth信息
		updateOAuthRecord(existingOAuth, gothUser)
		return nil
	}

	// 检查用户是否已绑定该提供商
	existingUserOAuth := userOAuth.GetByUserIDAndProvider(userID, userInfo.Provider)
	if existingUserOAuth != nil {
		return errors.New("您已绑定该平台账户")
	}

	// 创建新的OAuth绑定记录
	return createOAuthRecord(userID, gothUser, userInfo)
}

// GetUserOAuthBindings 获取用户的所有OAuth绑定
func GetUserOAuthBindings(userID uint64) map[string]*userOAuth.Entity {
	// 检查各个提供商的绑定状态
	providers := []string{ProviderGitHub, ProviderGoogle}
	return lo.PickBy(lo.Associate(providers, func(p string) (string, *userOAuth.Entity) {
		return p, userOAuth.GetByUserIDAndProvider(userID, p)
	}), func(_ string, v *userOAuth.Entity) bool {
		return v != nil
	})
}

// downloadAndSaveAvatar 下载外部头像并保存到本地
func downloadAndSaveAvatar(userID uint64, avatarURL string) (string, error) {
	if avatarURL == "" {
		return "", nil
	}

	// 设置下载超时时间（30秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建带超时的请求
	req, err := http.NewRequestWithContext(ctx, "GET", avatarURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 下载头像
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载头像失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载头像失败，状态码: %d", resp.StatusCode)
	}

	// 限制文件大小（最大2MB）
	const maxFileSize = 2 * 1024 * 1024 // 2MB
	limitedReader := io.LimitReader(resp.Body, maxFileSize+1)

	// 读取头像数据
	avatarData, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("读取头像数据失败: %v", err)
	}

	// 检查文件大小是否超过限制
	if len(avatarData) > maxFileSize {
		return "", fmt.Errorf("头像文件过大，最大允许2MB")
	}

	// 从URL中提取文件扩展名
	filename := "avatar"
	if urlPath := resp.Request.URL.Path; urlPath != "" {
		ext := path.Ext(urlPath)
		if ext != "" {
			filename = "avatar" + ext
		} else {
			// 如果没有扩展名，根据Content-Type推断
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
				filename = "avatar.jpg" // 默认为jpg
			}
		}
	}

	// 保存头像到本地
	fileEntity, err := filedata.SaveAvatar(userID, avatarData, filename)
	if err != nil {
		return "", fmt.Errorf("保存头像失败: %v", err)
	}

	return fileEntity.Name, nil
}
