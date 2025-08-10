package oauthservice

import (
	"errors"
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"log/slog"
	"time"

	"github.com/gorilla/sessions"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
)

const (
	// 支持的OAuth提供商
	ProviderGitHub   = "github"
	ProviderGoogle   = "google"
	ProviderFacebook = "facebook"
	ProviderTwitter  = "twitter"
)

var store *sessions.CookieStore

// InitOAuth 初始化OAuth配置
func InitOAuth() {
	// 初始化session store
	secretKey := preferences.GetString("app.signingKey", algorithm.SafeGenerateSigningKey(32))
	store = sessions.NewCookieStore([]byte(secretKey))

	// 设置goth的session store
	gothic.Store = store

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
func ProcessOAuthCallback(gothUser goth.User) (*users.EntityComplete, bool, error) {
	// 解析用户信息
	userInfo := parseOAuthUserInfo(gothUser)

	// 检查是否已存在OAuth记录
	existingOAuth := userOAuth.GetByProviderAndUID(userInfo.Provider, userInfo.ID)
	if existingOAuth != nil {
		// 已存在OAuth记录，直接登录
		user, err := users.Get(existingOAuth.UserId)
		if err != nil {
			return nil, false, fmt.Errorf("获取用户信息失败: %v", err)
		}

		// 更新OAuth信息
		updateOAuthRecord(existingOAuth, gothUser)
		return &user, false, nil
	}

	// 检查邮箱是否已被注册
	if gothUser.Email != "" && users.ExistEmail(gothUser.Email) {
		// 邮箱已存在，需要绑定到现有账户
		existingUser, err := users.GetByEmail(gothUser.Email)
		if err == nil {
			// 创建OAuth关联
			err := createOAuthRecord(existingUser.Id, gothUser, userInfo)
			if err != nil {
				return nil, false, err
			}
			return &existingUser, false, nil
		}
	}

	// 创建新用户
	newUser, err := createUserFromOAuth(gothUser, userInfo)
	if err != nil {
		return nil, false, err
	}

	// 创建OAuth记录
	err = createOAuthRecord(newUser.Id, gothUser, userInfo)
	if err != nil {
		return nil, false, err
	}

	return newUser, true, nil
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

	// 创建用户实体
	userEntity := &users.EntityComplete{
		Username: username,
		Email:    gothUser.Email,
		Password: "", // OAuth用户无密码
		Nickname: userInfo.Name,
		Bio:      userInfo.Bio,
		Website:  userInfo.Blog,
		Status:   0, // 正常状态
		Validate: 1, // GitHub用户默认已验证
	}

	// 如果有头像URL，设置头像
	if userInfo.AvatarURL != "" {
		userEntity.AvatarUrl = userInfo.AvatarURL
	}

	// 设置激活时间
	now := time.Now()
	userEntity.ActivatedAt = &now

	// 保存用户
	err := users.Create(userEntity)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

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

	return userOAuth.Delete(oauthEntity.Id)
}

// GetSessionStore 获取session store
func GetSessionStore() *sessions.CookieStore {
	return store
}
