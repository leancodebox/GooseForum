package oidcprovider

import (
	"context"
	"errors"
	"sync"
	"time"
)

// MemoryStore is a concurrency-safe reference Store for tests and local
// development. Values are cloned at its boundary so callers cannot mutate
// internal state without holding the lock.
type MemoryStore struct {
	mu              sync.Mutex
	clients         map[string]*Client
	consents        map[string]*Consent
	codes           map[string]*AuthorizationCode
	access, refresh map[string]*Token
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{clients: map[string]*Client{}, consents: map[string]*Consent{}, codes: map[string]*AuthorizationCode{}, access: map[string]*Token{}, refresh: map[string]*Token{}}
}
func (s *MemoryStore) GetClient(_ context.Context, id string) (*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneClient(s.clients[id]), nil
}
func (s *MemoryStore) PutClient(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c.ID] = cloneClient(c)
}
func (s *MemoryStore) GetConsent(_ context.Context, userID, clientID string) (*Consent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneConsent(s.consents[userID+"\x00"+clientID]), nil
}
func (s *MemoryStore) SaveConsent(_ context.Context, v *Consent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consents[v.UserID+"\x00"+v.ClientID] = cloneConsent(v)
	return nil
}
func (s *MemoryStore) SaveAuthorizationCode(_ context.Context, v *AuthorizationCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[v.Hash] = cloneCode(v)
	return nil
}
func (s *MemoryStore) GetAuthorizationCode(_ context.Context, hash string) (*AuthorizationCode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneCode(s.codes[hash]), nil
}

func (s *MemoryStore) ExchangeAuthorizationCode(_ context.Context, x AuthorizationCodeExchange) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	code := s.codes[x.CodeHash]
	if code != nil && code.Used {
		s.revokeAuthorizationGrantLocked(code.GrantID, x.Now)
		return errors.Join(ErrInvalidGrant, ErrAuthorizationCodeReplay)
	}
	if code == nil || !x.Now.Before(code.ExpiresAt) || code.ClientID != x.ClientID || code.RedirectURI != x.RedirectURI || code.CodeChallenge != "" && !verifyPKCE(x.CodeVerifier, code.CodeChallenge) || x.AccessToken == nil {
		return ErrInvalidGrant
	}
	code.Used = true
	s.access[x.AccessToken.Hash] = cloneToken(x.AccessToken)
	if x.RefreshToken != nil {
		s.refresh[x.RefreshToken.Hash] = cloneToken(x.RefreshToken)
	}
	return nil
}

func (s *MemoryStore) GetAccessToken(_ context.Context, hash string) (*Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneToken(s.access[hash]), nil
}
func (s *MemoryStore) GetRefreshToken(_ context.Context, hash string) (*Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneToken(s.refresh[hash]), nil
}

func (s *MemoryStore) RotateRefreshToken(_ context.Context, x RefreshTokenRotation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.refresh[x.OldHash]
	if old == nil || old.ClientID != x.ClientID || !x.Now.Before(old.ExpiresAt) || !old.FamilyExpiresAt.IsZero() && !x.Now.Before(old.FamilyExpiresAt) || x.AccessToken == nil || x.RefreshToken == nil {
		return ErrInvalidGrant
	}
	if old.RevokedAt != nil {
		for _, token := range s.access {
			if token.FamilyID == old.FamilyID {
				token.RevokedAt = timePointer(x.Now)
			}
		}
		for _, token := range s.refresh {
			if token.FamilyID == old.FamilyID {
				revokedAt := x.Now
				token.RevokedAt = &revokedAt
			}
		}
		return ErrInvalidGrant
	}
	revokedAt := x.Now
	old.RevokedAt = &revokedAt
	s.access[x.AccessToken.Hash] = cloneToken(x.AccessToken)
	s.refresh[x.RefreshToken.Hash] = cloneToken(x.RefreshToken)
	return nil
}

func (s *MemoryStore) RevokeToken(_ context.Context, hash, clientID string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v := s.access[hash]; v != nil && v.ClientID == clientID {
		v.RevokedAt = timePointer(now)
		return nil
	}
	if v := s.refresh[hash]; v != nil && v.ClientID == clientID {
		for _, token := range s.access {
			if token.FamilyID == v.FamilyID {
				token.RevokedAt = timePointer(now)
			}
		}
		for _, token := range s.refresh {
			if token.FamilyID == v.FamilyID {
				token.RevokedAt = timePointer(now)
			}
		}
	}
	return nil
}
func (s *MemoryStore) RevokeGrant(_ context.Context, userID, clientID string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.consents, userID+"\x00"+clientID)
	for _, code := range s.codes {
		if code.UserID == userID && code.ClientID == clientID && !code.Used {
			code.Used = true
		}
	}
	for _, token := range s.access {
		if token.UserID == userID && token.ClientID == clientID {
			token.RevokedAt = timePointer(now)
		}
	}
	for _, token := range s.refresh {
		if token.UserID == userID && token.ClientID == clientID {
			token.RevokedAt = timePointer(now)
		}
	}
	return nil
}
func (s *MemoryStore) RevokeAuthorizationCodeGrant(_ context.Context, codeHash string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	code := s.codes[codeHash]
	if code == nil || code.GrantID == "" {
		return nil
	}
	s.revokeAuthorizationGrantLocked(code.GrantID, now)
	return nil
}

func (s *MemoryStore) revokeAuthorizationGrantLocked(grantID string, now time.Time) {
	if grantID == "" {
		return
	}
	for _, token := range s.access {
		if token.GrantID == grantID {
			token.RevokedAt = timePointer(now)
		}
	}
	for _, token := range s.refresh {
		if token.GrantID == grantID {
			token.RevokedAt = timePointer(now)
		}
	}
}
func (s *MemoryStore) RevokeTokenFamily(_ context.Context, family string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, token := range s.access {
		if token.FamilyID == family {
			token.RevokedAt = timePointer(now)
		}
	}
	for _, token := range s.refresh {
		if token.FamilyID == family {
			token.RevokedAt = timePointer(now)
		}
	}
	return nil
}

func (s *MemoryStore) Purge(_ context.Context, request PurgeRequest) (PurgeResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := PurgeResult{}
	codes := make(map[string]*AuthorizationCode)
	for hash, code := range s.codes {
		if !code.Used && !request.Now.Before(code.ExpiresAt) {
			result.AuthorizationCodes++
			continue
		}
		if code.Used && !s.grantNeedsReplayTombstoneLocked(code.GrantID, request.Now) {
			result.AuthorizationCodes++
			continue
		}
		codes[hash] = code
	}
	access := make(map[string]*Token)
	for hash, token := range s.access {
		if !request.Now.Before(token.ExpiresAt) || token.RevokedAt != nil && !token.RevokedAt.After(request.RevokedBefore) {
			result.AccessTokens++
			continue
		}
		access[hash] = token
	}
	refresh := make(map[string]*Token)
	for hash, token := range s.refresh {
		familyExpired := !token.FamilyExpiresAt.IsZero() && !request.Now.Before(token.FamilyExpiresAt)
		if familyExpired || token.RevokedAt == nil && !request.Now.Before(token.ExpiresAt) {
			result.RefreshTokens++
			continue
		}
		refresh[hash] = token
	}
	s.codes, s.access, s.refresh = codes, access, refresh
	return result, nil
}

func (s *MemoryStore) grantNeedsReplayTombstoneLocked(grantID string, now time.Time) bool {
	if grantID == "" {
		return false
	}
	for _, token := range s.access {
		if token.GrantID == grantID && token.RevokedAt == nil && now.Before(token.ExpiresAt) {
			return true
		}
	}
	for _, token := range s.refresh {
		if token.GrantID == grantID && (token.FamilyExpiresAt.IsZero() || now.Before(token.FamilyExpiresAt)) {
			return true
		}
	}
	return false
}

func cloneClient(v *Client) *Client {
	if v == nil {
		return nil
	}
	copy := *v
	copy.RedirectURIs, copy.Scopes, copy.GrantTypes = cloneStrings(v.RedirectURIs), cloneStrings(v.Scopes), cloneStrings(v.GrantTypes)
	return &copy
}
func cloneConsent(v *Consent) *Consent {
	if v == nil {
		return nil
	}
	copy := *v
	copy.Scopes = cloneStrings(v.Scopes)
	return &copy
}
func cloneCode(v *AuthorizationCode) *AuthorizationCode {
	if v == nil {
		return nil
	}
	copy := *v
	copy.Scopes = cloneStrings(v.Scopes)
	return &copy
}
func cloneToken(v *Token) *Token {
	if v == nil {
		return nil
	}
	copy := *v
	copy.Scopes = cloneStrings(v.Scopes)
	if v.RevokedAt != nil {
		copy.RevokedAt = timePointer(*v.RevokedAt)
	}
	return &copy
}
func timePointer(value time.Time) *time.Time { return &value }
