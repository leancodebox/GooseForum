package mfaservice

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/authsessions"
	"github.com/leancodebox/GooseForum/app/models/forum/usermfa"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/adminpolicyservice"
	"github.com/leancodebox/GooseForum/app/service/authsessionservice"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

// AdminResetBy checks the operator permission again before resetting MFA.
func AdminResetBy(actorID, userID uint64, reason string) error {
	if actorID == 0 || actorID == userID {
		return ErrVerification
	}
	return adminReset(actorID, userID, reason)
}

func adminReset(actorID, userID uint64, reason string) error {
	if strings.TrimSpace(reason) == "" || len(reason) > 256 {
		return errors.New("a recovery reason of at most 256 bytes is required")
	}
	err := func() error {
		if actorID != 0 {
			snapshot, err := adminpolicyservice.ReadSnapshot()
			if err != nil {
				return err
			}
			actor, err := users.GetAccountState(actorID)
			if err != nil {
				return err
			}
			grants, err := adminpolicyservice.Permissions(actor.RoleId)
			if err != nil {
				return err
			}
			if actor.EffectiveRestriction(time.Now()) != users.RestrictionNormal || actor.NeedsEmailVerification(snapshot.EmailVerificationEnabled) || !grants[actor.RoleId][permission.Admin.Id()] {
				return ErrVerification
			}
			// Keep the audit identity within the existing bounded reason field.
			reason = fmt.Sprintf("admin:%d %s", actorID, strings.TrimSpace(reason))
			if len(reason) > 256 {
				return ErrVerification
			}
		}
		advanced, err := users.AdvanceTokenVersion(userID)
		if err != nil {
			return err
		}
		if !advanced {
			return errors.New("user not found")
		}
		defer func() {
			authsessionservice.InvalidateUser(userID)
			if user, err := users.Get(userID); err == nil {
				userservice.RefreshUserCaches(&user)
			}
		}()
		if err := authsessions.RevokeAllTokens(userID, time.Now()); err != nil {
			return err
		}
		if err := usermfa.DeleteRecoveryCodes(userID); err != nil {
			return err
		}
		if err := clearUserState(userID); err != nil {
			return err
		}
		if _, err := usermfa.DeleteFactor(userID); err != nil {
			return err
		}
		return authsessions.CreateSecurityLog(&authsessions.Log{UserId: userID, Action: "mfa_admin_reset", Result: "success", AuthMethod: "operator", Reason: strings.TrimSpace(reason), CreatedAt: time.Now()})
	}()
	if err != nil {
		return err
	}
	return nil
}
