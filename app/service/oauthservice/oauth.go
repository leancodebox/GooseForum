package oauthservice

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/randopt"
	"github.com/leancodebox/GooseForum/app/bundles/sessionstore"
	"github.com/leancodebox/GooseForum/app/service/emailactivationservice"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/leancodebox/GooseForum/app/service/filestorage"
	"github.com/leancodebox/GooseForum/app/service/registrationservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"

	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/models/forum/userOAuth"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/samber/lo"
)

// InitOAuth configures available OAuth providers.
func InitOAuth() {
	gothic.Store = sessionstore.GetSession()
	if err := ReloadProviders(loadSettings()); err != nil {
		slog.Error("OAuth provider initialization failed", "err", err)
		return
	}
	if count := len(EnabledProviders()); count > 0 {
		slog.Info("OAuth providers initialized", "count", count)
	} else {
		slog.Warn("no OAuth providers enabled")
	}
}

// OAuthUserInfo is the normalized user data from an OAuth provider.
type OAuthUserInfo struct {
	ID            string `json:"id"`
	Login         string `json:"login"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	AvatarURL     string `json:"avatar_url"`
	Bio           string `json:"bio"`
	Blog          string `json:"blog"`
	Location      string `json:"location"`
	Provider      string `json:"provider"`
	EmailVerified bool   `json:"emailVerified"`
}

var oauthRegistrationMu sync.Mutex

// ProcessOAuthCallback logs in an existing OAuth user or creates a new one.
func ProcessOAuthCallback(gothUser goth.User, clientIP ...string) (*users.EntityComplete, error) {
	userInfo := parseOAuthUserInfo(gothUser)
	if err := validateOAuthUserInfo(userInfo); err != nil {
		return nil, err
	}
	// Recheck the provider identity while serializing first-time callbacks.
	oauthRegistrationMu.Lock()
	ip := ""
	if len(clientIP) > 0 {
		ip = clientIP[0]
	}
	// Recover before the existing-binding path or new-signup restrictions.
	pending, err := users.FindOAuthRegistration(users.OAuthRegistrationKey(userInfo.Provider, userInfo.ID))
	if err != nil {
		oauthRegistrationMu.Unlock()
		return nil, err
	}
	if pending != nil {
		recovered, err := userservice.CreateUserWithBinding(pending.Username, "", pending.Email, pending.RequiresEmailVerification, &userOAuth.Entity{Provider: userInfo.Provider, ProviderUid: userInfo.ID})
		oauthRegistrationMu.Unlock()
		if err != nil {
			return nil, err
		}
		return finishOAuthRegistration(recovered, userInfo, ip), nil
	}

	existingOAuth := userOAuth.GetByProviderAndUID(userInfo.Provider, userInfo.ID)
	if existingOAuth != nil {
		oauthRegistrationMu.Unlock()
		user, err := users.Get(existingOAuth.UserId)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %w", err)
		}
		if user.NeedsEmailVerification(hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification) && user.Email != "" {
			if registrationservice.AllowMail(ip, user.Email) == nil {
				if _, err := emailactivationservice.Resend(user); err != nil && !errors.Is(err, emailactivationservice.ErrCooldown) && !errors.Is(err, emailactivationservice.ErrDailyLimit) {
					slog.Warn("OAuth verification email could not be sent", "userId", user.Id, "error", err)
				}
			}
		}
		return &user, nil
	}
	email, config, err := registrationservice.Check(userInfo.Email, ip, true)
	if err != nil {
		oauthRegistrationMu.Unlock()
		return nil, err
	}
	userInfo.Email = email
	needVerification := (config.EnableEmailVerification || len(config.AllowedDomains) > 0) && !userInfo.EmailVerified
	if err := registrationservice.CheckVerificationMail(needVerification); err != nil {
		oauthRegistrationMu.Unlock()
		return nil, err
	}
	newUser, err := createUserFromOAuth(userInfo, needVerification)
	oauthRegistrationMu.Unlock()
	if err != nil {
		return nil, err
	}
	return finishOAuthRegistration(newUser, userInfo, ip), nil
}

func finishOAuthRegistration(newUser *users.EntityComplete, userInfo OAuthUserInfo, ip string) *users.EntityComplete {
	if err := populateOAuthProfile(newUser, userInfo); err != nil {
		slog.Warn("OAuth profile update failed", "userId", newUser.Id, "error", err)
	}

	if newUser.IsActivated == users.ActivationPending && newUser.Email != "" {
		if registrationservice.AllowMail(ip, newUser.Email) == nil {
			if err := emailactivationservice.SendActivationEmail(newUser); err != nil {
				slog.Warn("OAuth verification email could not be sent", "userId", newUser.Id, "error", err)
			}
		}
	}

	eventbus.Publish(context.Background(), &eventhandlers.UserSignUpEvent{
		UserId:   newUser.Id,
		Username: newUser.Username,
	})

	return newUser
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
		userInfo.EmailVerified, _ = gothUser.RawData["email_verified"].(bool)
		if gothUser.Provider == "discord" {
			userInfo.EmailVerified, _ = gothUser.RawData["verified"].(bool)
		}
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

func validateOAuthUserInfo(userInfo OAuthUserInfo) error {
	if strings.TrimSpace(userInfo.Provider) == "" {
		return errors.New("OAuth provider is missing")
	}
	if strings.TrimSpace(userInfo.ID) == "" {
		return errors.New("OAuth provider user ID is missing")
	}
	return nil
}

// createUserFromOAuth creates a local account from OAuth user data.
func createUserFromOAuth(userInfo OAuthUserInfo, needVerification bool) (*users.EntityComplete, error) {
	username := oauthUsername(userInfo)
	originalUsername := username
	counter := 1
	for users.ExistUsername(username) {
		suffix := fmt.Sprintf("_%d", counter)
		base := originalUsername
		if len(base)+len(suffix) > 32 {
			base = base[:32-len(suffix)]
		}
		username = base + suffix
		counter++
	}

	binding := &userOAuth.Entity{Provider: userInfo.Provider, ProviderUid: userInfo.ID}
	var userEntity *users.EntityComplete
	var err error
	for range 10 {
		userEntity, err = userservice.CreateUserWithBinding(username, randopt.RandomString(32), userInfo.Email, needVerification, binding)
		if !errors.Is(err, users.ErrUsernameExists) {
			break
		}
		username = fmt.Sprintf("%.20s_%d", originalUsername, counter)
		counter++
	}
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	return userEntity, nil
}

func populateOAuthProfile(userEntity *users.EntityComplete, userInfo OAuthUserInfo) error {

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

	userEntity.Nickname = strings.TrimSpace(userInfo.Name)
	if userEntity.Nickname == "" {
		userEntity.Nickname = userEntity.Username
	}
	userEntity.Bio = userInfo.Bio
	userEntity.Website = userInfo.Blog
	if err := userservice.SaveUser(userEntity); err != nil {
		return err
	}

	return nil
}

var invalidUsernameCharacters = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func oauthUsername(userInfo OAuthUserInfo) string {
	base := strings.Trim(invalidUsernameCharacters.ReplaceAllString(strings.TrimSpace(userInfo.Login), "_"), "_-")
	if base == "" {
		base = strings.Trim(invalidUsernameCharacters.ReplaceAllString(strings.TrimSpace(userInfo.Name), "_"), "_-")
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(userInfo.Provider+":"+userInfo.ID)))[:8]
	if base == "" {
		base = "oauth"
	}
	if len(base) > 23 {
		base = base[:23]
	}
	if len(base) < 4 {
		base += "_" + digest
	}
	return base
}

// createOAuthRecord stores a provider account binding.
func createOAuthRecord(userID uint64, userInfo OAuthUserInfo) error {
	oauthEntity := &userOAuth.Entity{
		UserId:       userID,
		Provider:     userInfo.Provider,
		ProviderUid:  userInfo.ID,
		AccessToken:  "",
		RefreshToken: "",
		Scopes:       "",
		RawUserData:  "",
	}

	return userOAuth.Create(oauthEntity)
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
		return fmt.Errorf("获取用户信息失败: %w", err)
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
	if err := validateOAuthUserInfo(userInfo); err != nil {
		return err
	}
	oauthRegistrationMu.Lock()
	defer oauthRegistrationMu.Unlock()

	existingOAuth := userOAuth.GetByProviderAndUID(userInfo.Provider, userInfo.ID)
	if existingOAuth != nil {
		if existingOAuth.UserId != userID {
			return errors.New("该OAuth账户已被其他用户绑定")
		}
		return nil
	}

	existingUserOAuth := userOAuth.GetByUserIDAndProvider(userID, userInfo.Provider)
	if existingUserOAuth != nil {
		return errors.New("您已绑定该平台账户")
	}

	return createOAuthRecord(userID, userInfo)
}

// GetUserOAuthBindings returns active OAuth bindings keyed by provider.
func GetUserOAuthBindings(userID uint64) map[string]*userOAuth.Entity {
	providers := enabledProviderKeys()
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, avatarURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载头像失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载头像失败，状态码: %d", resp.StatusCode)
	}

	const maxFileSize = 2 * 1024 * 1024
	limitedReader := io.LimitReader(resp.Body, maxFileSize+1)

	avatarData, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("读取头像数据失败: %w", err)
	}

	if len(avatarData) > maxFileSize {
		return "", errors.New("头像文件过大，最大允许2MB")
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

	fileEntity, err := filestorage.SaveAvatar(context.Background(), userID, avatarData, filename)
	if err != nil {
		return "", fmt.Errorf("保存头像失败: %w", err)
	}

	return fileEntity.Name, nil
}
