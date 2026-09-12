package oidcProviderStore

import "time"

const (
	clientTable            = "oidc_clients"
	consentTable           = "oidc_consents"
	codeTable              = "oidc_authorization_codes"
	tokenTable             = "oidc_tokens"
	signingKeyTable        = "oidc_signing_keys"
	signingKeyHistoryTable = "oidc_signing_key_history"
	interactionTable       = "oidc_interactions"

	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type ClientEntity struct {
	ID                      string   `gorm:"primaryKey;column:client_id;type:varchar(255);not null"`
	Name                    string   `gorm:"column:name;type:varchar(255);not null;default:''"`
	SecretHash              string   `gorm:"column:secret_hash;type:varchar(64);not null;default:''"`
	RedirectURIs            []string `gorm:"column:redirect_uris;type:text;not null;serializer:json"`
	Scopes                  []string `gorm:"column:scopes;type:text;not null;serializer:json"`
	GrantTypes              []string `gorm:"column:grant_types;type:text;not null;serializer:json"`
	TokenEndpointAuthMethod string   `gorm:"column:token_endpoint_auth_method;type:varchar(32);not null"`
	RequirePKCE             bool     `gorm:"column:require_pkce;not null;default:false"`
	Public                  bool     `gorm:"column:is_public;not null;default:false"`
	Enabled                 bool     `gorm:"column:enabled;not null;default:true;index:idx_oidc_clients_enabled"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func (ClientEntity) TableName() string { return clientTable }

type ConsentEntity struct {
	UserID    string   `gorm:"primaryKey;column:user_id;type:varchar(255);not null"`
	ClientID  string   `gorm:"primaryKey;column:client_id;type:varchar(255);not null;index:idx_oidc_consents_client"`
	Scopes    []string `gorm:"column:scopes;type:text;not null;serializer:json"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ConsentEntity) TableName() string { return consentTable }

type AuthorizationCodeEntity struct {
	Hash                string    `gorm:"primaryKey;column:code_hash;type:varchar(64);not null"`
	GrantID             string    `gorm:"column:grant_id;type:varchar(64);not null;default:'';index:idx_oidc_codes_grant"`
	UserID              string    `gorm:"column:user_id;type:varchar(255);not null;index:idx_oidc_codes_grant_owner,priority:1"`
	ClientID            string    `gorm:"column:client_id;type:varchar(255);not null;index:idx_oidc_codes_grant_owner,priority:2"`
	RedirectURI         string    `gorm:"column:redirect_uri;type:text;not null"`
	Scopes              []string  `gorm:"column:scopes;type:text;not null;serializer:json"`
	Nonce               string    `gorm:"column:nonce;type:varchar(255);not null;default:''"`
	CodeChallenge       string    `gorm:"column:code_challenge;type:varchar(128);not null;default:''"`
	CodeChallengeMethod string    `gorm:"column:code_challenge_method;type:varchar(16);not null;default:''"`
	ExpiresAt           time.Time `gorm:"column:expires_at;not null;index:idx_oidc_codes_expiry"`
	AuthTime            time.Time `gorm:"column:auth_time;not null"`
	Used                bool      `gorm:"column:used;not null;default:false;index:idx_oidc_codes_grant_owner,priority:3"`
	CreatedAt           time.Time
}

func (AuthorizationCodeEntity) TableName() string { return codeTable }

type TokenEntity struct {
	Hash            string     `gorm:"primaryKey;column:token_hash;type:varchar(64);not null"`
	Type            string     `gorm:"column:token_type;type:varchar(8);not null;index:idx_oidc_tokens_type_expiry,priority:1"`
	GrantID         string     `gorm:"column:grant_id;type:varchar(64);not null;default:'';index:idx_oidc_tokens_grant"`
	FamilyID        string     `gorm:"column:family_id;type:varchar(64);not null;default:'';index:idx_oidc_tokens_family"`
	UserID          string     `gorm:"column:user_id;type:varchar(255);not null;index:idx_oidc_tokens_owner,priority:1"`
	ClientID        string     `gorm:"column:client_id;type:varchar(255);not null;index:idx_oidc_tokens_owner,priority:2"`
	Scopes          []string   `gorm:"column:scopes;type:text;not null;serializer:json"`
	ExpiresAt       time.Time  `gorm:"column:expires_at;not null;index:idx_oidc_tokens_type_expiry,priority:2"`
	FamilyExpiresAt *time.Time `gorm:"column:family_expires_at;index:idx_oidc_tokens_family_expiry"`
	RevokedAt       *time.Time `gorm:"column:revoked_at;index:idx_oidc_tokens_revoked"`
	AuthTime        time.Time  `gorm:"column:auth_time;not null"`
	CreatedAt       time.Time
}

func (TokenEntity) TableName() string { return tokenTable }

// SigningKeyEntity stores the single active OIDC signing key. The private key
// is encrypted by the runtime service before it crosses this persistence
// boundary.
type SigningKeyEntity struct {
	ID                  uint8  `gorm:"primaryKey;column:id;not null"`
	KID                 string `gorm:"column:kid;type:varchar(64);not null;uniqueIndex"`
	EncryptedPrivateKey []byte `gorm:"column:encrypted_private_key;not null"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (SigningKeyEntity) TableName() string { return signingKeyTable }

type SigningKeyHistoryEntity struct {
	KID       string    `gorm:"primaryKey;column:kid;type:varchar(64);not null"`
	PublicKey []byte    `gorm:"column:public_key;not null"`
	RetireAt  time.Time `gorm:"column:retire_at;not null;index:idx_oidc_signing_key_history_retire"`
	CreatedAt time.Time
}

func (SigningKeyHistoryEntity) TableName() string { return signingKeyHistoryTable }

// InteractionEntity is a short-lived, single-use browser interaction. Hash is
// the digest of the random handle shown to the browser; the handle itself is
// never persisted. RequestJSON is the server-validated authorization request.
type InteractionEntity struct {
	Hash           string     `gorm:"primaryKey;column:interaction_hash;type:varchar(64);not null"`
	RequestJSON    []byte     `gorm:"column:request_json;not null"`
	Purpose        string     `gorm:"column:purpose;type:varchar(16);not null;default:consent;index:idx_oidc_interactions_purpose"`
	UserID         string     `gorm:"column:user_id;type:varchar(255);not null;index:idx_oidc_interactions_user"`
	AuthTime       time.Time  `gorm:"column:auth_time;not null"`
	AuthFresh      bool       `gorm:"column:auth_fresh;not null;default:false"`
	PreviousAuthID string     `gorm:"column:previous_auth_id;type:varchar(64);not null;default:''"`
	StartedAt      time.Time  `gorm:"column:started_at;not null;default:CURRENT_TIMESTAMP"`
	ExpiresAt      time.Time  `gorm:"column:expires_at;not null;index:idx_oidc_interactions_expiry"`
	Used           bool       `gorm:"column:used;not null;default:false"`
	UsedAt         *time.Time `gorm:"column:used_at"`
	CreatedAt      time.Time
}

func (InteractionEntity) TableName() string { return interactionTable }

// Models returns every schema object required by Store.
func Models() []any {
	return []any{&ClientEntity{}, &ConsentEntity{}, &AuthorizationCodeEntity{}, &TokenEntity{}, &SigningKeyEntity{}, &SigningKeyHistoryEntity{}, &InteractionEntity{}}
}
