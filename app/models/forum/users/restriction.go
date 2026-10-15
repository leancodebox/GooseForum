package users

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
)

const (
	RestrictionNormal    = "normal"
	RestrictionSuspended = "suspended"
	RestrictionBanned    = "banned"
)

// AccountState is read directly for authorization, independently of display caches.
type AccountState struct {
	TrustLevel                trustlevel.Level
	TrustLevelLocked          bool
	CreatedAt                 time.Time
	Id                        uint64
	RoleId                    uint64
	TokenVersion              uint64
	IsActivated               int8
	RequiresEmailVerification bool
	RestrictionStatus         string
	RestrictionUntil          *time.Time
	RestrictionReason         string
}

func effectiveRestriction(status string, until *time.Time, now time.Time) string {
	if status == "" {
		return RestrictionNormal
	}
	if until != nil && !until.After(now) {
		return RestrictionNormal
	}
	switch status {
	case RestrictionNormal, RestrictionSuspended, RestrictionBanned:
		return status
	default:
		return RestrictionBanned
	}
}

func (state AccountState) EffectiveRestriction(now time.Time) string {
	return effectiveRestriction(state.RestrictionStatus, state.RestrictionUntil, now)
}

func (state AccountState) NeedsEmailVerification(enabled bool) bool {
	return state.IsActivated == ActivationPending && (enabled || state.RequiresEmailVerification)
}

func (user EntityComplete) EffectiveRestriction(now time.Time) string {
	return effectiveRestriction(user.RestrictionStatus, user.RestrictionUntil, now)
}
