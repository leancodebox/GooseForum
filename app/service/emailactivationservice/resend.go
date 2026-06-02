package emailactivationservice

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/kvstore"
)

const (
	resendDailyLimit = 3
	resendInterval   = time.Minute
)

var (
	ErrDisabled        = errors.New("email activation resend disabled")
	ErrAlreadyVerified = errors.New("email activation already verified")
	ErrCooldown        = errors.New("email activation resend cooldown")
	ErrDailyLimit      = errors.New("email activation resend daily limit")
)

type ResendResult struct {
	RemainingToday    int
	RetryAfterSeconds int64
	DailyLimit        int
}

type resendState struct {
	Count      int   `json:"count"`
	LastSentAt int64 `json:"lastSentAt"`
}

func Resend(userEntity users.EntityComplete) (ResendResult, error) {
	result := ResendResult{DailyLimit: resendDailyLimit}

	securityConfig := hotdataserve.GetSecuritySettingsConfigCache()
	if !securityConfig.EnableEmailVerification {
		return result, ErrDisabled
	}
	if userEntity.IsActivated != users.ActivationPending {
		return result, ErrAlreadyVerified
	}

	result, err := consumeQuota(userEntity.Id, time.Now())
	if err != nil {
		return result, err
	}

	if err = SendActivationEmail(&userEntity); err != nil {
		slog.Info("验证邮件重发失败", "userId", userEntity.Id, "error", err)
		return result, err
	}

	return result, nil
}

func consumeQuota(userID uint64, now time.Time) (ResendResult, error) {
	result := ResendResult{DailyLimit: resendDailyLimit}
	key, ttl := resendKey(userID, now)
	err := kvstore.UpdateBytes(key, ttl, func(current []byte, exists bool) (kvstore.UpdateAction, []byte, error) {
		state := resendState{}
		if exists && len(current) > 0 {
			if err := json.Unmarshal(current, &state); err != nil {
				state = resendState{}
			}
		}

		if state.LastSentAt > 0 {
			nextAllowedAt := time.Unix(state.LastSentAt, 0).Add(resendInterval)
			if now.Before(nextAllowedAt) {
				result.RetryAfterSeconds = int64(nextAllowedAt.Sub(now).Seconds())
				if result.RetryAfterSeconds < 1 {
					result.RetryAfterSeconds = 1
				}
				return kvstore.UpdateKeep, nil, ErrCooldown
			}
		}
		if state.Count >= resendDailyLimit {
			return kvstore.UpdateKeep, nil, ErrDailyLimit
		}

		state.Count++
		state.LastSentAt = now.Unix()
		data, err := json.Marshal(state)
		if err == nil {
			result.RemainingToday = max(resendDailyLimit-state.Count, 0)
		}
		return kvstore.UpdateSet, data, err
	})
	return result, err
}

func resendKey(userID uint64, now time.Time) (string, time.Duration) {
	day := now.Format("20060102")
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	ttl := time.Until(tomorrow) + time.Hour
	if ttl < time.Hour {
		ttl = time.Hour
	}
	return "activation-email:resend:" + strconv.FormatUint(userID, 10) + ":" + day, ttl
}
