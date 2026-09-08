// Package oidcproviderservice assembles GooseForum persistence and users with
// the transport-independent OIDC provider core.
package oidcproviderservice

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"github.com/leancodebox/GooseForum/app/bundles/preferences"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"gorm.io/gorm"
)

const (
	issuerPath       = "/oauth2"
	signingKeyBits   = 2048
	defaultRetention = 24 * time.Hour
)

type Options struct {
	DB                      *gorm.DB
	SiteURL                 string
	KeyEncryptionSecret     []byte
	AllowInsecureIssuer     bool
	Now                     func() time.Time
	Random                  io.Reader
	AuthorizationCodeTTL    time.Duration
	AccessTokenTTL          time.Duration
	RefreshTokenTTL         time.Duration
	RefreshTokenMaxLifetime time.Duration
	IDTokenTTL              time.Duration
	SupportedScopes         []string
}

type Service struct {
	provider   *core.Provider
	store      *oidcProviderStore.Store
	signer     *core.RSAKeySet
	now        func() time.Time
	random     io.Reader
	idTokenTTL time.Duration
}

// New builds a runtime service without changing the database schema. Database
// migrations must run before this function is called.
func New(ctx context.Context, opts Options) (*Service, error) {
	if ctx == nil {
		return nil, errors.New("OIDC provider context is required")
	}
	if opts.DB == nil {
		return nil, errors.New("OIDC provider database is required")
	}
	issuer, siteBase, err := deriveIssuer(opts.SiteURL, opts.AllowInsecureIssuer)
	if err != nil {
		return nil, err
	}
	if len(opts.KeyEncryptionSecret) < 32 {
		return nil, errors.New("OIDC signing-key encryption secret must contain at least 32 bytes")
	}
	store, err := oidcProviderStore.New(opts.DB)
	if err != nil {
		return nil, fmt.Errorf("create OIDC store: %w", err)
	}
	signer, err := loadOrCreateSigner(ctx, store, opts.KeyEncryptionSecret, opts.Random)
	if err != nil {
		return nil, fmt.Errorf("initialize OIDC signing key: %w", err)
	}
	provider, err := core.New(core.Config{
		Issuer:                  issuer,
		AuthorizationCodeTTL:    opts.AuthorizationCodeTTL,
		AccessTokenTTL:          opts.AccessTokenTTL,
		RefreshTokenTTL:         opts.RefreshTokenTTL,
		RefreshTokenMaxLifetime: opts.RefreshTokenMaxLifetime,
		IDTokenTTL:              opts.IDTokenTTL,
		Store:                   store,
		Users:                   &userResolver{db: opts.DB, siteBase: siteBase},
		Signer:                  signer,
		Now:                     opts.Now,
		AllowInsecureIssuer:     opts.AllowInsecureIssuer,
		Random:                  opts.Random,
		SupportedScopes:         opts.SupportedScopes,
	})
	if err != nil {
		return nil, fmt.Errorf("create OIDC provider: %w", err)
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	random := opts.Random
	if random == nil {
		random = rand.Reader
	}
	idTokenTTL := opts.IDTokenTTL
	if idTokenTTL <= 0 {
		idTokenTTL = 10 * time.Minute
	}
	return &Service{provider: provider, store: store, signer: signer, now: now, random: random, idTokenTTL: idTokenTTL}, nil
}

func (s *Service) InteractionStore() *oidcProviderStore.Store {
	if s == nil {
		return nil
	}
	return s.store
}

func (s *Service) Now() func() time.Time {
	if s == nil {
		return nil
	}
	return s.now
}

func (s *Service) Random() io.Reader {
	if s == nil {
		return nil
	}
	return s.random
}

func (s *Service) Provider() *core.Provider {
	if s == nil {
		return nil
	}
	return s.provider
}

// Purge deletes expired protocol state. It is deliberately caller-driven so
// the service owns no ticker or goroutine and can be scheduled by GooseForum's
// existing maintenance facilities.
func (s *Service) Purge(ctx context.Context, revokedRetention time.Duration) (core.PurgeResult, error) {
	if s == nil || s.provider == nil {
		return core.PurgeResult{}, errors.New("OIDC provider service is not initialized")
	}
	if revokedRetention < 0 {
		return core.PurgeResult{}, errors.New("OIDC revoked-token retention cannot be negative")
	}
	result, err := s.provider.Purge(ctx, revokedRetention)
	if err != nil {
		return result, err
	}
	_, err = s.store.PurgeInteractions(ctx, s.now())
	if err != nil {
		return result, err
	}
	now := s.now()
	if _, err = s.store.PurgeSigningKeyHistory(ctx, now); err != nil {
		return result, err
	}
	s.signer.Prune(now)
	return result, nil
}

func (s *Service) RotateSigningKey(ctx context.Context, secret []byte) error {
	stored, err := s.store.GetSigningKey(ctx)
	if err != nil || stored == nil {
		return errors.Join(errors.New("load active OIDC signing key"), err)
	}
	plaintext, err := decryptPrivateKey(secret, stored.KID, stored.EncryptedPrivateKey)
	if err != nil {
		return err
	}
	defer clear(plaintext)
	oldPrivate, err := core.ParseRSAPrivateKeyPEM(plaintext)
	if err != nil {
		return err
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&oldPrivate.PublicKey)
	if err != nil {
		return err
	}
	next, err := buildSigningKey(secret, s.random)
	if err != nil {
		return err
	}
	defer clear(next.EncryptedPrivateKey)
	return s.store.RotateSigningKey(ctx, stored.KID, next, &oidcProviderStore.SigningKeyHistoryEntity{
		KID: stored.KID, PublicKey: publicDER, RetireAt: s.now().Add(s.idTokenTTL),
	})
}

func (s *Service) PurgeDefault(ctx context.Context) (core.PurgeResult, error) {
	return s.Purge(ctx, defaultRetention)
}

var defaultRuntime runtimeState

var ErrUnavailable = errors.New("OIDC provider is not configured")

type runtimeState struct {
	reloadMu sync.Mutex
	mu       sync.RWMutex
	service  *Service
	err      error
}

type RuntimeStatus struct {
	Enabled   bool   `json:"enabled"`
	Available bool   `json:"available"`
	Issuer    string `json:"issuer,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (s *runtimeState) get() (*Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.service != nil {
		return s.service, nil
	}
	if s.err != nil {
		return nil, s.err
	}
	return nil, ErrUnavailable
}

func (s *runtimeState) replace(service *Service, err error) {
	s.mu.Lock()
	s.service = service
	s.err = err
	s.mu.Unlock()
}

func newDefaultService() (*Service, error) {
	siteURL := strings.TrimSpace(hotdataserve.GetSiteSettingsConfigCache().SiteUrl)
	return New(context.Background(), Options{
		DB:                  dbconnect.Connect(),
		SiteURL:             siteURL,
		KeyEncryptionSecret: []byte(preferences.GetString("app.signingKey")),
		AllowInsecureIssuer: setting.IsLocal(),
	})
}

// InitOIDC initializes the configured provider during application startup.
func InitOIDC() {
	if !Configured() {
		return
	}
	if err := ReloadDefault(); err != nil {
		slog.Error("OIDC provider initialization failed", "err", err)
	}
}

// ReloadDefault rebuilds the process-wide runtime from current configuration.
// Failed initialization makes only the OIDC endpoints unavailable; callers can
// fix the configuration and invoke ReloadDefault again without restarting.
func ReloadDefault() error {
	defaultRuntime.reloadMu.Lock()
	defer defaultRuntime.reloadMu.Unlock()

	if !Configured() {
		defaultRuntime.replace(nil, ErrUnavailable)
		return nil
	}
	service, err := newDefaultService()
	defaultRuntime.replace(service, err)
	return err
}

func RotateDefaultSigningKey(ctx context.Context) error {
	defaultRuntime.reloadMu.Lock()
	defer defaultRuntime.reloadMu.Unlock()
	service, err := defaultRuntime.get()
	if err != nil {
		return err
	}
	if err := service.RotateSigningKey(ctx, []byte(preferences.GetString("app.signingKey"))); err != nil {
		return err
	}
	next, err := newDefaultService()
	defaultRuntime.replace(next, err)
	return err
}

// Default returns the currently active process-wide runtime.
func Default() (*Service, error) { return defaultRuntime.get() }

func Status() RuntimeStatus {
	status := RuntimeStatus{Enabled: Configured()}
	service, err := Default()
	if service != nil && err == nil {
		status.Available = true
		status.Issuer = service.Provider().Metadata().Issuer
		return status
	}
	if status.Enabled && err != nil {
		status.Error = err.Error()
	}
	return status
}

// DefaultProvider is the narrow entry point used by transport adapters.
func DefaultProvider() (*core.Provider, error) {
	service, err := Default()
	if err != nil {
		return nil, err
	}
	return service.Provider(), nil
}

// DefaultInteractions is a stateless lazy proxy used while routes are
// registered before migrations and runtime initialization complete.
type DefaultInteractions struct{}

func (DefaultInteractions) CreateInteraction(ctx context.Context, entity *oidcProviderStore.InteractionEntity) error {
	service, err := Default()
	if err != nil {
		return err
	}
	return service.store.CreateInteraction(ctx, entity)
}
func (DefaultInteractions) GetInteraction(ctx context.Context, hash, userID string, now time.Time) (*oidcProviderStore.InteractionEntity, error) {
	service, err := Default()
	if err != nil {
		return nil, err
	}
	return service.store.GetInteraction(ctx, hash, userID, now)
}
func (DefaultInteractions) ConsumeInteraction(ctx context.Context, hash, userID string, now time.Time) (*oidcProviderStore.InteractionEntity, error) {
	service, err := Default()
	if err != nil {
		return nil, err
	}
	return service.store.ConsumeInteraction(ctx, hash, userID, now)
}
func (DefaultInteractions) ConsumeLoginInteraction(ctx context.Context, hash, authID string, authTime, now time.Time) (*oidcProviderStore.InteractionEntity, error) {
	service, err := Default()
	if err != nil {
		return nil, err
	}
	return service.store.ConsumeLoginInteraction(ctx, hash, authID, authTime, now)
}

// Configured reports whether the OIDC provider is enabled in site settings.
func Configured() bool {
	settings := pageConfig.GetConfigByPageType(pageConfig.OIDCProvider, pageConfig.OIDCProviderSettingsConfig{})
	return settings.Enabled
}

func deriveIssuer(raw string, allowInsecure bool) (issuer string, siteBase *url.URL, err error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	parsed, err := url.Parse(raw)
	validScheme := err == nil && (strings.EqualFold(parsed.Scheme, "https") || allowInsecure && strings.EqualFold(parsed.Scheme, "http"))
	if !validScheme || !parsed.IsAbs() || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || strings.Contains(raw, "#") {
		return "", nil, errors.New("OIDC site URL must be an absolute URL without query, fragment, or user information")
	}
	return raw + issuerPath, parsed, nil
}

type userResolver struct {
	db       *gorm.DB
	siteBase *url.URL
}

func (r *userResolver) ResolveUser(ctx context.Context, subject string) (*core.User, error) {
	id, err := strconv.ParseUint(subject, 10, 64)
	if err != nil || id == 0 {
		return nil, core.ErrUserUnavailable
	}
	var entity users.EntityComplete
	err = r.db.WithContext(ctx).Where("id = ?", id).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, core.ErrUserUnavailable
	}
	if err != nil {
		return nil, err
	}
	if entity.IsFrozen == users.StatusFrozen {
		return nil, core.ErrUserUnavailable
	}
	name := entity.Nickname
	if name == "" {
		name = entity.Username
	}
	picture := absoluteURL(r.siteBase, entity.GetWebAvatarUrl())
	return &core.User{
		Subject:           strconv.FormatUint(entity.Id, 10),
		Name:              name,
		PreferredUsername: entity.Username,
		Email:             entity.Email,
		EmailVerified:     entity.IsActivated == users.ActivationSuccess,
		Picture:           picture,
	}, nil
}

func absoluteURL(base *url.URL, value string) string {
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.IsAbs() {
		return parsed.String()
	}
	return base.ResolveReference(parsed).String()
}

func loadOrCreateSigner(ctx context.Context, store *oidcProviderStore.Store, secret []byte, random io.Reader) (*core.RSAKeySet, error) {
	stored, err := store.GetSigningKey(ctx)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		stored, err = createSigningKey(ctx, store, secret, random)
		if err != nil {
			return nil, err
		}
	}
	plaintext, err := decryptPrivateKey(secret, stored.KID, stored.EncryptedPrivateKey)
	if err != nil {
		return nil, err
	}
	defer clear(plaintext)
	privateKey, err := core.ParseRSAPrivateKeyPEM(plaintext)
	if err != nil {
		return nil, err
	}
	keySet, err := core.NewRSAKeySet(stored.KID, privateKey)
	if err != nil {
		return nil, err
	}
	history, err := store.ListSigningKeyHistory(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range history {
		parsed, parseErr := x509.ParsePKIXPublicKey(item.PublicKey)
		if parseErr != nil {
			return nil, fmt.Errorf("parse retired OIDC signing key %q: %w", item.KID, parseErr)
		}
		publicKey, ok := parsed.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("retired OIDC signing key %q is not RSA", item.KID)
		}
		if err := keySet.AddVerificationKey(item.KID, publicKey); err != nil {
			return nil, err
		}
		if err := keySet.Retire(item.KID, item.RetireAt); err != nil {
			return nil, err
		}
	}
	return keySet, nil
}

func createSigningKey(ctx context.Context, store *oidcProviderStore.Store, secret []byte, random io.Reader) (*oidcProviderStore.SigningKeyEntity, error) {
	key, err := buildSigningKey(secret, random)
	if err != nil {
		return nil, err
	}
	defer clear(key.EncryptedPrivateKey)
	created, err := store.CreateSigningKeyIfAbsent(ctx, key)
	if err != nil {
		return nil, err
	}
	if !created {
		stored, err := store.GetSigningKey(ctx)
		if err != nil {
			return nil, err
		}
		if stored == nil {
			return nil, errors.New("OIDC signing key disappeared after concurrent initialization")
		}
		return stored, nil
	}
	return store.GetSigningKey(ctx)
}

func buildSigningKey(secret []byte, random io.Reader) (*oidcProviderStore.SigningKeyEntity, error) {
	if random == nil {
		random = rand.Reader
	}
	privateKey, err := rsa.GenerateKey(random, signingKeyBits)
	if err != nil {
		return nil, err
	}
	kidBytes := make([]byte, 16)
	if _, err := io.ReadFull(random, kidBytes); err != nil {
		return nil, err
	}
	kid := base64.RawURLEncoding.EncodeToString(kidBytes)
	clear(kidBytes)
	pemData, err := core.MarshalRSAPrivateKeyPEM(privateKey)
	if err != nil {
		return nil, err
	}
	defer clear(pemData)
	encrypted, err := encryptPrivateKey(secret, kid, pemData, random)
	if err != nil {
		return nil, err
	}
	return &oidcProviderStore.SigningKeyEntity{KID: kid, EncryptedPrivateKey: encrypted}, nil
}

func encryptPrivateKey(secret []byte, kid string, plaintext []byte, random io.Reader) ([]byte, error) {
	aead, err := keyAEAD(secret)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(random, nonce); err != nil {
		return nil, err
	}
	return aead.Seal(nonce, nonce, plaintext, []byte(kid)), nil
}

func decryptPrivateKey(secret []byte, kid string, encrypted []byte) ([]byte, error) {
	aead, err := keyAEAD(secret)
	if err != nil {
		return nil, err
	}
	if len(encrypted) < aead.NonceSize()+aead.Overhead() {
		return nil, errors.New("encrypted OIDC signing key is truncated")
	}
	nonce := encrypted[:aead.NonceSize()]
	plaintext, err := aead.Open(nil, nonce, encrypted[aead.NonceSize():], []byte(kid))
	if err != nil {
		return nil, errors.New("decrypt OIDC signing key: encryption secret does not match or data is corrupt")
	}
	return plaintext, nil
}

func keyAEAD(secret []byte) (cipher.AEAD, error) {
	if len(secret) < 32 {
		return nil, errors.New("OIDC signing-key encryption secret must contain at least 32 bytes")
	}
	key := sha256.Sum256(secret)
	block, err := aes.NewCipher(key[:])
	clear(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
