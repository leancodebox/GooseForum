package accountrestrictionservice

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/leancodebox/GooseForum/app/models/forum/accountrestrictions"
	"github.com/leancodebox/GooseForum/app/models/forum/oidcProviderStore"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/adminpolicyservice"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

var (
	ErrInvalid   = errors.New("invalid account restriction")
	ErrProtected = errors.New("protected administrator account")
)

type Change struct {
	UserId   uint64     `json:"userId"`
	Status   string     `json:"restrictionStatus"`
	Until    *time.Time `json:"restrictionUntil"`
	Reason   string     `json:"restrictionReason"`
	Note     string     `json:"restrictionNote"`
	RoleId   uint64     `json:"roleId"`
	Validate int8       `json:"validate"`
}

func Edit(actorID uint64, change Change) error {
	err := edit(actorID, change, time.Now())
	authsessionservice.InvalidateUser(change.UserId)
	if user, err := users.Get(change.UserId); err == nil {
		userservice.RefreshUserCaches(&user)
	}
	return err
}

func edit(actorID uint64, change Change, now time.Time) error {
	change.Reason = strings.TrimSpace(change.Reason)
	change.Note = strings.TrimSpace(change.Note)
	if change.Status != users.RestrictionNormal && change.Status != users.RestrictionSuspended && change.Status != users.RestrictionBanned {
		return ErrInvalid
	}
	if change.Validate != 0 && change.Validate != 1 {
		return ErrInvalid
	}
	if utf8.RuneCountInString(change.Reason) > 500 || utf8.RuneCountInString(change.Note) > 2000 {
		return ErrInvalid
	}
	if change.Status == users.RestrictionNormal {
		change.Until = nil
	}
	if change.Until != nil && !change.Until.After(now) {
		return ErrInvalid
	}
	apply := func() error {
		snapshot, err := adminpolicyservice.ReadSnapshot()
		if err != nil {
			return err
		}
		verificationEnabled := snapshot.EmailVerificationEnabled
		target, err := users.GetRestrictionAccount(change.UserId)
		if err != nil {
			return err
		}
		actor := users.AccountState{}
		if actorID != 0 {
			actor, err = users.GetAccountState(actorID)
			if err != nil {
				return err
			}
		}
		permissions, err := adminpolicyservice.Permissions(actor.RoleId, target.RoleId, change.RoleId)
		if err != nil {
			return err
		}
		actorAdmin := actorID == 0 || permissions[actor.RoleId][permission.Admin.Id()]
		if actor.EffectiveRestriction(now) != users.RestrictionNormal || !actorAdmin && !permissions[actor.RoleId][permission.UserManager.Id()] {
			return ErrProtected
		}
		targetAdmin := permissions[target.RoleId][permission.Admin.Id()]
		nextAdmin := permissions[change.RoleId][permission.Admin.Id()]
		if !actorAdmin {
			for _, granted := range []map[uint64]bool{permissions[target.RoleId], permissions[change.RoleId]} {
				for permissionID := range granted {
					if !permissions[actor.RoleId][permissionID] {
						return ErrProtected
					}
				}
			}
		}
		if actorID == change.UserId && (change.Status != users.RestrictionNormal || change.RoleId != target.RoleId) {
			return ErrProtected
		}
		if change.RoleId != 0 {
			exists, err := role.ExistsEffective(change.RoleId)
			if err != nil {
				return err
			}
			if !exists {
				return ErrInvalid
			}
		}
		verificationPending := change.Validate == users.ActivationPending && (verificationEnabled || target.RequiresEmailVerification)
		if targetAdmin && (!nextAdmin || change.Status != users.RestrictionNormal || verificationPending) {
			remaining, err := users.HasNormalAdministrator(snapshot.AdminRoleIDs, target.Id, 0, now, verificationEnabled)
			if err != nil {
				return err
			}
			if !remaining {
				return ErrProtected
			}
		}
		oldStatus := target.RestrictionStatus
		if oldStatus == "" {
			oldStatus = target.EffectiveRestriction(now)
		}
		changed := oldStatus != change.Status || !sameTime(target.RestrictionUntil, change.Until) || target.RestrictionReason != change.Reason || target.RestrictionNote != change.Note
		if changed && change.Reason == "" {
			return ErrInvalid
		}
		if err := users.UpdateRestriction(target.Id, users.RestrictionUpdate{RoleId: change.RoleId, Activation: change.Validate, Status: change.Status, Until: change.Until, Reason: change.Reason, Note: change.Note, RevokeSessions: changed && change.Status == users.RestrictionBanned || target.RoleId != change.RoleId}); err != nil {
			return err
		}
		if change.Status == users.RestrictionBanned {
			subject := strconv.FormatUint(target.Id, 10)
			if err := oidcProviderStore.RevokeAccountTokens(subject, now); err != nil {
				return err
			}
		}
		if changed || change.Reason != "" {
			if err := accountrestrictions.EnsureHistory(&accountrestrictions.History{UserId: target.Id, ActorId: actorID, Status: change.Status, Until: change.Until, Reason: change.Reason, Note: change.Note, CreatedAt: now}); err != nil {
				return err
			}
		}
		return nil
	}
	return apply()
}

func sameTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}
