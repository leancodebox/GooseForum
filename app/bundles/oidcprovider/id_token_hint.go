package oidcprovider

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"

	"github.com/golang-jwt/jwt/v5"
)

// Hints may describe past sessions, so expiration is deliberately not enforced.
// Signature, issuer and client binding still have to be valid.
func (p *Provider) validateIDTokenHint(ctx context.Context, req AuthorizeRequest, auth Authentication) error {
	if req.IDTokenHint == "" {
		return nil
	}
	keys, err := p.cfg.Signer.PublicKeys(ctx)
	if err != nil {
		return authorizationError(req, "server_error", "verification keys unavailable", errors.Join(ErrServer, err))
	}
	claims := &jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(req.IDTokenHint, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		for _, key := range keys {
			if key.Kid != kid || key.KTY != "RSA" || key.Alg != "RS256" {
				continue
			}
			n, ne := base64.RawURLEncoding.DecodeString(key.N)
			e, ee := base64.RawURLEncoding.DecodeString(key.E)
			exponent := new(big.Int).SetBytes(e)
			if ne != nil || ee != nil || !exponent.IsInt64() || exponent.Int64() < 3 || exponent.Int64() > 2147483647 {
				return nil, errors.New("invalid RSA key")
			}
			return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(exponent.Int64())}, nil
		}
		return nil, errors.New("unknown signing key")
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithoutClaimsValidation())
	if err != nil || claims.Issuer != p.cfg.Issuer || !validSubject(claims.Subject) || !contains([]string(claims.Audience), req.ClientID) || claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return authorizationError(req, "invalid_request", "id_token_hint is invalid", ErrInvalidRequest)
	}
	if auth.UserID == "" || auth.AuthTime.IsZero() {
		return nil
	}
	user, err := p.resolveUser(ctx, auth.UserID)
	if err != nil {
		if errors.Is(err, ErrUserUnavailable) {
			return authorizationError(req, "login_required", "hinted user is unavailable", ErrLoginRequired)
		}
		return authorizationError(req, "server_error", "user resolution failed", err)
	}
	if user.Subject != claims.Subject {
		return authorizationError(req, "login_required", "id_token_hint does not match the authenticated user", ErrLoginRequired)
	}
	return nil
}
