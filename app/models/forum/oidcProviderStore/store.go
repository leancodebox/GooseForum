package oidcProviderStore

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"time"

	core "github.com/leancodebox/GooseForum/app/bundles/oidcprovider"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrInteractionInvalid = errors.New("OIDC interaction is missing, expired, used, or belongs to another user")

const (
	InteractionPurposeConsent = "consent"
	InteractionPurposeLogin   = "login"
)

type Store struct{ db *gorm.DB }

type UserGrant struct {
	ClientID  string
	Name      string
	Scopes    []string
	GrantedAt time.Time
	Enabled   bool
}

func New(db *gorm.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("oidc provider store requires a database")
	}
	return &Store{db: db}, nil
}

func (s *Store) Migrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(Models()...)
}

// GetSigningKey returns the active singleton signing-key record.
func (s *Store) GetSigningKey(ctx context.Context) (*SigningKeyEntity, error) {
	var entity SigningKeyEntity
	err := s.db.WithContext(ctx).Where("id = ?", uint8(1)).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// CreateSigningKeyIfAbsent atomically installs the first signing key. A
// concurrent process that lost the insert race receives created=false and
// must reload the winner.
func (s *Store) CreateSigningKeyIfAbsent(ctx context.Context, entity *SigningKeyEntity) (created bool, err error) {
	if entity == nil || entity.KID == "" || len(entity.EncryptedPrivateKey) == 0 {
		return false, errors.New("complete signing key is required")
	}
	owned := SigningKeyEntity{
		ID:                  1,
		KID:                 entity.KID,
		EncryptedPrivateKey: append([]byte(nil), entity.EncryptedPrivateKey...),
	}
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&owned)
	return result.RowsAffected == 1, result.Error
}

func (s *Store) ListSigningKeyHistory(ctx context.Context) ([]SigningKeyHistoryEntity, error) {
	var entities []SigningKeyHistoryEntity
	err := s.db.WithContext(ctx).Order("created_at ASC, kid ASC").Find(&entities).Error
	for i := range entities {
		entities[i].PublicKey = append([]byte(nil), entities[i].PublicKey...)
	}
	return entities, err
}

func (s *Store) RotateSigningKey(ctx context.Context, previousKID string, next *SigningKeyEntity, history *SigningKeyHistoryEntity) error {
	if previousKID == "" || next == nil || next.KID == "" || len(next.EncryptedPrivateKey) == 0 || history == nil || history.KID != previousKID || len(history.PublicKey) == 0 || history.RetireAt.IsZero() {
		return errors.New("complete OIDC signing-key rotation is required")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(history).Error; err != nil {
			return err
		}
		result := tx.Model(&SigningKeyEntity{}).Where("id = ? AND kid = ?", uint8(1), previousKID).Updates(map[string]any{
			"kid": next.KID, "encrypted_private_key": next.EncryptedPrivateKey,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("active OIDC signing key changed concurrently")
		}
		return nil
	})
}

func (s *Store) PurgeSigningKeyHistory(ctx context.Context, now time.Time) (int64, error) {
	result := s.db.WithContext(ctx).Where("retire_at <= ?", now).Delete(&SigningKeyHistoryEntity{})
	return result.RowsAffected, result.Error
}

func (s *Store) SaveClient(ctx context.Context, client *core.Client) error {
	if client == nil {
		return errors.New("client is required")
	}
	if err := core.ValidateClient(*client); err != nil {
		return err
	}
	entity := clientToEntity(client)
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&entity).Error
}

// CreateClient inserts a registration without overwriting an existing client.
// Administrative creation uses this so an ID collision cannot replace an
// existing client's credentials or redirect URI allowlist.
func (s *Store) CreateClient(ctx context.Context, client *core.Client) error {
	if client == nil {
		return errors.New("client is required")
	}
	if err := core.ValidateClient(*client); err != nil {
		return err
	}
	entity := clientToEntity(client)
	return s.db.WithContext(ctx).Create(&entity).Error
}

func (s *Store) ListClients(ctx context.Context) ([]core.Client, error) {
	var entities []ClientEntity
	if err := s.db.WithContext(ctx).Order("created_at ASC, client_id ASC").Find(&entities).Error; err != nil {
		return nil, err
	}
	result := make([]core.Client, 0, len(entities))
	for _, entity := range entities {
		result = append(result, *entity.client())
	}
	return result, nil
}

func (s *Store) GetClient(ctx context.Context, id string) (*core.Client, error) {
	var entity ClientEntity
	err := s.db.WithContext(ctx).Where("client_id = ?", id).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return entity.client(), nil
}

func (s *Store) GetConsent(ctx context.Context, userID, clientID string) (*core.Consent, error) {
	var entity ConsentEntity
	err := s.db.WithContext(ctx).Where("user_id = ? AND client_id = ?", userID, clientID).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &core.Consent{UserID: entity.UserID, ClientID: entity.ClientID, Scopes: cloneStrings(entity.Scopes)}, nil
}

func (s *Store) ListUserGrants(ctx context.Context, userID string) ([]UserGrant, error) {
	var consents []ConsentEntity
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at DESC, client_id ASC").Find(&consents).Error; err != nil {
		return nil, err
	}
	clientIDs := make([]string, 0, len(consents))
	for _, consent := range consents {
		clientIDs = append(clientIDs, consent.ClientID)
	}
	var clients []ClientEntity
	if len(clientIDs) > 0 {
		if err := s.db.WithContext(ctx).Where("client_id IN ?", clientIDs).Find(&clients).Error; err != nil {
			return nil, err
		}
	}
	clientsByID := make(map[string]ClientEntity, len(clients))
	for _, client := range clients {
		clientsByID[client.ID] = client
	}
	result := make([]UserGrant, 0, len(consents))
	for _, consent := range consents {
		client := clientsByID[consent.ClientID]
		name := client.Name
		if name == "" {
			name = consent.ClientID
		}
		result = append(result, UserGrant{
			ClientID: consent.ClientID, Name: name, Scopes: cloneStrings(consent.Scopes),
			GrantedAt: consent.UpdatedAt, Enabled: client.Enabled,
		})
	}
	return result, nil
}

func (s *Store) SaveConsent(ctx context.Context, consent *core.Consent) error {
	if consent == nil {
		return errors.New("consent is required")
	}
	entity := ConsentEntity{UserID: consent.UserID, ClientID: consent.ClientID, Scopes: cloneStrings(consent.Scopes)}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "client_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"scopes", "updated_at"}),
	}).Create(&entity).Error
}

func (s *Store) CreateInteraction(ctx context.Context, interaction *InteractionEntity) error {
	if interaction == nil || len(interaction.Hash) != sha256.Size*2 || len(interaction.RequestJSON) == 0 || interaction.ExpiresAt.IsZero() || interaction.StartedAt.IsZero() {
		return errors.New("complete OIDC interaction is required")
	}
	if interaction.Purpose != InteractionPurposeConsent && interaction.Purpose != InteractionPurposeLogin {
		return errors.New("valid OIDC interaction purpose is required")
	}
	if interaction.Purpose == InteractionPurposeConsent && (interaction.UserID == "" || interaction.AuthTime.IsZero()) {
		return errors.New("consent interaction requires authentication")
	}
	owned := *interaction
	owned.RequestJSON = append([]byte(nil), interaction.RequestJSON...)
	owned.Used = false
	owned.UsedAt = nil
	return s.db.WithContext(ctx).Create(&owned).Error
}

// GetInteraction returns only a live interaction belonging to userID.
func (s *Store) GetInteraction(ctx context.Context, hash, userID string, now time.Time) (*InteractionEntity, error) {
	var entity InteractionEntity
	err := s.db.WithContext(ctx).
		Where("interaction_hash = ? AND purpose = ? AND user_id = ? AND used = ? AND expires_at > ?", hash, InteractionPurposeConsent, userID, false, now).
		Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInteractionInvalid
	}
	if err != nil {
		return nil, err
	}
	entity.RequestJSON = append([]byte(nil), entity.RequestJSON...)
	return &entity, nil
}

// ConsumeInteraction atomically marks one live, user-bound interaction used
// and returns its trusted request snapshot. Exactly one concurrent caller can
// succeed on SQLite, MySQL, or any GORM database with transactional updates.
func (s *Store) ConsumeInteraction(ctx context.Context, hash, userID string, now time.Time) (*InteractionEntity, error) {
	var entity InteractionEntity
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&InteractionEntity{}).
			Where("interaction_hash = ? AND purpose = ? AND user_id = ? AND used = ? AND expires_at > ?", hash, InteractionPurposeConsent, userID, false, now).
			Updates(map[string]any{"used": true, "used_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrInteractionInvalid
		}
		if err := tx.Where("interaction_hash = ?", hash).Take(&entity).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	entity.RequestJSON = append([]byte(nil), entity.RequestJSON...)
	return &entity, nil
}

// ConsumeLoginInteraction atomically marks a live login interaction used only
// after a distinct authentication ceremony completed after it began.
func (s *Store) ConsumeLoginInteraction(ctx context.Context, hash, authID string, authTime, now time.Time) (*InteractionEntity, error) {
	if authID == "" || authTime.IsZero() {
		return nil, ErrInteractionInvalid
	}
	var entity InteractionEntity
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&InteractionEntity{}).
			Where("interaction_hash = ? AND purpose = ? AND used = ? AND expires_at > ? AND started_at <= ? AND (previous_auth_id = ? OR previous_auth_id <> ?)",
				hash, InteractionPurposeLogin, false, now, authTime, "", authID).
			Updates(map[string]any{"used": true, "used_at": now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrInteractionInvalid
		}
		return tx.Where("interaction_hash = ?", hash).Take(&entity).Error
	})
	if err != nil {
		return nil, err
	}
	entity.RequestJSON = append([]byte(nil), entity.RequestJSON...)
	return &entity, nil
}

func (s *Store) PurgeInteractions(ctx context.Context, now time.Time) (int64, error) {
	result := s.db.WithContext(ctx).Where("expires_at <= ?", now).Delete(&InteractionEntity{})
	return result.RowsAffected, result.Error
}

func (s *Store) SaveAuthorizationCode(ctx context.Context, code *core.AuthorizationCode) error {
	if code == nil {
		return errors.New("authorization code is required")
	}
	entity := codeToEntity(code)
	return s.db.WithContext(ctx).Create(&entity).Error
}

func (s *Store) GetAuthorizationCode(ctx context.Context, hash string) (*core.AuthorizationCode, error) {
	var entity AuthorizationCodeEntity
	err := s.db.WithContext(ctx).Where("code_hash = ?", hash).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return entity.code(), nil
}

func (s *Store) ExchangeAuthorizationCode(ctx context.Context, exchange core.AuthorizationCodeExchange) error {
	invalid := false
	replay := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var code AuthorizationCodeEntity
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("code_hash = ?", exchange.CodeHash).Take(&code).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			invalid = true
			return nil
		}
		if err != nil {
			return err
		}
		if code.Used {
			if err := tx.Model(&TokenEntity{}).Where("grant_id = ?", code.GrantID).Update("revoked_at", exchange.Now).Error; err != nil {
				return err
			}
			replay = true
			return nil
		}
		if !exchange.Now.Before(code.ExpiresAt) || code.ClientID != exchange.ClientID || code.RedirectURI != exchange.RedirectURI || code.CodeChallenge != "" && !verifyPKCE(exchange.CodeVerifier, code.CodeChallenge) || exchange.AccessToken == nil {
			invalid = true
			return nil
		}
		result := tx.Model(&AuthorizationCodeEntity{}).
			Where("code_hash = ? AND used = ? AND expires_at > ?", exchange.CodeHash, false, exchange.Now).
			Update("used", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			if err := tx.Model(&TokenEntity{}).Where("grant_id = ?", code.GrantID).Update("revoked_at", exchange.Now).Error; err != nil {
				return err
			}
			replay = true
			return nil
		}
		if err := createToken(tx, tokenTypeAccess, exchange.AccessToken); err != nil {
			return err
		}
		if exchange.RefreshToken != nil {
			if err := createToken(tx, tokenTypeRefresh, exchange.RefreshToken); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if replay {
		return errors.Join(core.ErrInvalidGrant, core.ErrAuthorizationCodeReplay)
	}
	if invalid {
		return core.ErrInvalidGrant
	}
	return nil
}

func (s *Store) GetAccessToken(ctx context.Context, hash string) (*core.Token, error) {
	return s.getToken(ctx, hash, tokenTypeAccess)
}

func (s *Store) GetRefreshToken(ctx context.Context, hash string) (*core.Token, error) {
	return s.getToken(ctx, hash, tokenTypeRefresh)
}

func (s *Store) getToken(ctx context.Context, hash, tokenType string) (*core.Token, error) {
	var entity TokenEntity
	err := s.db.WithContext(ctx).Where("token_hash = ? AND token_type = ?", hash, tokenType).Take(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return entity.token(), nil
}

func (s *Store) RotateRefreshToken(ctx context.Context, rotation core.RefreshTokenRotation) error {
	invalid := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var old TokenEntity
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash = ? AND token_type = ?", rotation.OldHash, tokenTypeRefresh).Take(&old).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			invalid = true
			return nil
		}
		if err != nil {
			return err
		}
		familyExpired := old.FamilyExpiresAt != nil && !rotation.Now.Before(*old.FamilyExpiresAt)
		if old.ClientID != rotation.ClientID || !rotation.Now.Before(old.ExpiresAt) || familyExpired || rotation.AccessToken == nil || rotation.RefreshToken == nil {
			invalid = true
			return nil
		}
		if old.RevokedAt != nil {
			if err := revokeFamily(tx, old.FamilyID, rotation.Now); err != nil {
				return err
			}
			invalid = true
			return nil
		}
		result := tx.Model(&TokenEntity{}).
			Where("token_hash = ? AND token_type = ? AND revoked_at IS NULL", rotation.OldHash, tokenTypeRefresh).
			Update("revoked_at", rotation.Now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			if err := revokeFamily(tx, old.FamilyID, rotation.Now); err != nil {
				return err
			}
			invalid = true
			return nil
		}
		if err := createToken(tx, tokenTypeAccess, rotation.AccessToken); err != nil {
			return err
		}
		return createToken(tx, tokenTypeRefresh, rotation.RefreshToken)
	})
	if err != nil {
		return err
	}
	if invalid {
		return core.ErrInvalidGrant
	}
	return nil
}

func (s *Store) RevokeToken(ctx context.Context, hash, clientID string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&TokenEntity{}).Where("token_hash = ? AND token_type = ? AND client_id = ?", hash, tokenTypeAccess, clientID).Update("revoked_at", now)
		if result.Error != nil || result.RowsAffected != 0 {
			return result.Error
		}
		var refresh TokenEntity
		err := tx.Where("token_hash = ? AND token_type = ? AND client_id = ?", hash, tokenTypeRefresh, clientID).Take(&refresh).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		return revokeFamily(tx, refresh.FamilyID, now)
	})
}

func (s *Store) RevokeTokenFamily(ctx context.Context, family string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return revokeFamily(tx, family, now) })
}

func (s *Store) RevokeGrant(ctx context.Context, userID, clientID string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND client_id = ?", userID, clientID).Delete(&ConsentEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&AuthorizationCodeEntity{}).Where("user_id = ? AND client_id = ? AND used = ?", userID, clientID, false).Update("used", true).Error; err != nil {
			return err
		}
		return tx.Model(&TokenEntity{}).Where("user_id = ? AND client_id = ?", userID, clientID).Update("revoked_at", now).Error
	})
}

func (s *Store) RevokeAuthorizationCodeGrant(ctx context.Context, codeHash string, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var code AuthorizationCodeEntity
		err := tx.Select("grant_id").Where("code_hash = ?", codeHash).Take(&code).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if code.GrantID == "" {
			return nil
		}
		return tx.Model(&TokenEntity{}).Where("grant_id = ?", code.GrantID).Update("revoked_at", now).Error
	})
}

func (s *Store) Purge(ctx context.Context, request core.PurgeRequest) (core.PurgeResult, error) {
	result := core.PurgeResult{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Used codes are replay-detection tombstones while any credential from
		// their grant can still be used. Purge them before tokens so both tests
		// observe the same transaction-start credential state.
		codes := tx.Where(`
			(used = ? AND expires_at <= ?)
			OR
			(used = ? AND (grant_id = '' OR NOT EXISTS (
				SELECT 1 FROM oidc_tokens
				WHERE oidc_tokens.grant_id = oidc_authorization_codes.grant_id
				AND (
					(oidc_tokens.token_type = ? AND oidc_tokens.revoked_at IS NULL AND oidc_tokens.expires_at > ?)
					OR
					(oidc_tokens.token_type = ? AND (oidc_tokens.family_expires_at IS NULL OR oidc_tokens.family_expires_at > ?))
				)
			)))`, false, request.Now, true, tokenTypeAccess, request.Now, tokenTypeRefresh, request.Now).
			Delete(&AuthorizationCodeEntity{})
		if codes.Error != nil {
			return codes.Error
		}
		result.AuthorizationCodes = int(codes.RowsAffected)
		access := tx.Where("token_type = ? AND (expires_at <= ? OR (revoked_at IS NOT NULL AND revoked_at <= ?))", tokenTypeAccess, request.Now, request.RevokedBefore).Delete(&TokenEntity{})
		if access.Error != nil {
			return access.Error
		}
		result.AccessTokens = int(access.RowsAffected)
		// A rotated/revoked refresh token is retained until its family expires;
		// it is the evidence needed to detect replay and revoke that family.
		refresh := tx.Where("token_type = ? AND ((family_expires_at IS NOT NULL AND family_expires_at <= ?) OR (revoked_at IS NULL AND expires_at <= ?))", tokenTypeRefresh, request.Now, request.Now).Delete(&TokenEntity{})
		if refresh.Error != nil {
			return refresh.Error
		}
		result.RefreshTokens = int(refresh.RowsAffected)
		return nil
	})
	if err != nil {
		return core.PurgeResult{}, err
	}
	return result, err
}

func createToken(tx *gorm.DB, tokenType string, token *core.Token) error {
	if token == nil {
		return errors.New("token is required")
	}
	entity := tokenToEntity(tokenType, token)
	return tx.Create(&entity).Error
}

func revokeFamily(tx *gorm.DB, family string, now time.Time) error {
	if family == "" {
		return nil
	}
	return tx.Model(&TokenEntity{}).Where("family_id = ?", family).Update("revoked_at", now).Error
}

func verifyPKCE(verifier, challenge string) bool {
	if len(verifier) < 43 || len(verifier) > 128 || len(challenge) != 43 {
		return false
	}
	for _, char := range verifier {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '.' || char == '_' || char == '~') {
			return false
		}
	}
	sum := sha256.Sum256([]byte(verifier))
	actual := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(actual), []byte(challenge)) == 1
}

func clientToEntity(client *core.Client) ClientEntity {
	return ClientEntity{ID: client.ID, Name: client.Name, SecretHash: client.SecretHash, RedirectURIs: cloneStrings(client.RedirectURIs), Scopes: cloneStrings(client.Scopes), GrantTypes: cloneStrings(client.GrantTypes), TokenEndpointAuthMethod: string(client.TokenEndpointAuthMethod), RequirePKCE: client.RequirePKCE, Public: client.Public, Enabled: client.Enabled}
}

func (entity ClientEntity) client() *core.Client {
	return &core.Client{ID: entity.ID, Name: entity.Name, SecretHash: entity.SecretHash, RedirectURIs: cloneStrings(entity.RedirectURIs), Scopes: cloneStrings(entity.Scopes), GrantTypes: cloneStrings(entity.GrantTypes), TokenEndpointAuthMethod: core.ClientAuthenticationMethod(entity.TokenEndpointAuthMethod), RequirePKCE: entity.RequirePKCE, Public: entity.Public, Enabled: entity.Enabled}
}

func codeToEntity(code *core.AuthorizationCode) AuthorizationCodeEntity {
	return AuthorizationCodeEntity{Hash: code.Hash, GrantID: code.GrantID, UserID: code.UserID, ClientID: code.ClientID, RedirectURI: code.RedirectURI, Scopes: cloneStrings(code.Scopes), Nonce: code.Nonce, CodeChallenge: code.CodeChallenge, CodeChallengeMethod: code.CodeChallengeMethod, ExpiresAt: code.ExpiresAt, AuthTime: code.AuthTime, Used: code.Used}
}

func (entity AuthorizationCodeEntity) code() *core.AuthorizationCode {
	return &core.AuthorizationCode{Hash: entity.Hash, GrantID: entity.GrantID, UserID: entity.UserID, ClientID: entity.ClientID, RedirectURI: entity.RedirectURI, Scopes: cloneStrings(entity.Scopes), Nonce: entity.Nonce, CodeChallenge: entity.CodeChallenge, CodeChallengeMethod: entity.CodeChallengeMethod, ExpiresAt: entity.ExpiresAt, AuthTime: entity.AuthTime, Used: entity.Used}
}

func tokenToEntity(tokenType string, token *core.Token) TokenEntity {
	var familyExpiry *time.Time
	if !token.FamilyExpiresAt.IsZero() {
		value := token.FamilyExpiresAt
		familyExpiry = &value
	}
	return TokenEntity{Hash: token.Hash, Type: tokenType, GrantID: token.GrantID, FamilyID: token.FamilyID, UserID: token.UserID, ClientID: token.ClientID, Scopes: cloneStrings(token.Scopes), ExpiresAt: token.ExpiresAt, FamilyExpiresAt: familyExpiry, RevokedAt: cloneTime(token.RevokedAt), AuthTime: token.AuthTime}
}

func (entity TokenEntity) token() *core.Token {
	var familyExpiry time.Time
	if entity.FamilyExpiresAt != nil {
		familyExpiry = *entity.FamilyExpiresAt
	}
	return &core.Token{Hash: entity.Hash, GrantID: entity.GrantID, FamilyID: entity.FamilyID, UserID: entity.UserID, ClientID: entity.ClientID, Scopes: cloneStrings(entity.Scopes), ExpiresAt: entity.ExpiresAt, FamilyExpiresAt: familyExpiry, RevokedAt: cloneTime(entity.RevokedAt), AuthTime: entity.AuthTime}
}

func cloneStrings(values []string) []string { return append([]string(nil), values...) }

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

var _ core.Store = (*Store)(nil)
