package trustlevelservice

import (
	"errors"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/leancodebox/GooseForum/app/bundles/sharedcache"
	"github.com/leancodebox/GooseForum/app/bundles/trustlevel"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"github.com/leancodebox/GooseForum/app/models/forum/trusthistory"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/notificationservice"
	"github.com/leancodebox/GooseForum/app/service/userservice"
)

var ErrConflict = errors.New("trust level changed; reload and retry")
var ErrInvalid = errors.New("invalid trust level adjustment; levels cannot be lowered")

var reviewCache = &sharedcache.Cache[bool]{Name: "trust-review", MaxEntries: 4096}

func ClearReviewCache() { reviewCache.Clear() }

// Only no-progress reviews are throttled; topic visits always evaluate current policy.
func ReviewAuthenticated(state users.AccountState, now time.Time) error {
	if state.TrustLevelLocked || state.TrustLevel >= trustlevel.Regular {
		return nil
	}
	_, err := reviewCache.GetOrLoadE(strconv.FormatUint(state.Id, 10), func() (bool, error) { return true, ReviewWithState(state, 0, now) }, time.Minute)
	return err
}

func Next(s users.TrustState, c pageConfig.TrustConfig, now time.Time, topicID uint64) users.TrustState {
	n := s
	if n.TrustLevelLocked || n.TrustLevel >= trustlevel.Regular || !trustlevel.Valid(n.TrustLevel) {
		return n
	}
	if topicID > 0 {
		day := now.UTC().Format(time.DateOnly)
		if day > n.TrustLastVisitDay {
			n.TrustLastVisitDay = day
			if n.TrustVisitDays < math.MaxUint32 {
				n.TrustVisitDays++
			}
		}
		if len(n.TrustTopicIds) < trustlevel.MaxTopics && !slices.Contains(n.TrustTopicIds, topicID) {
			n.TrustTopicIds = append(slices.Clone(n.TrustTopicIds), topicID)
		}
	}
	for _, r := range c.PromotionRules {
		if r.Enabled && r.Level > n.TrustLevel && r.Level <= trustlevel.Regular && !now.Before(s.CreatedAt.Add(time.Duration(r.MinAccountAgeHours)*time.Hour)) && n.TrustVisitDays >= r.MinVisitDays && len(n.TrustTopicIds) >= r.MinTopicsVisited {
			n.TrustLevel = r.Level
		}
	}
	return n
}

func Review(userID, topicID uint64, now time.Time) error {
	if userID == 0 {
		return nil
	}
	state, err := users.GetAccountState(userID)
	if err != nil {
		return err
	}
	return ReviewWithState(state, topicID, now)
}

func ReviewWithState(state users.AccountState, topicID uint64, now time.Time) error {
	userID := state.Id
	if state.TrustLevelLocked || state.TrustLevel >= trustlevel.Regular {
		return nil
	}
	c, err := hotdataserve.GetTrustSettings()
	if err != nil {
		return err
	}
	if topicID == 0 {
		possible := false
		for _, r := range c.PromotionRules {
			if r.Enabled && r.Level > state.TrustLevel && !now.Before(state.CreatedAt.Add(time.Duration(r.MinAccountAgeHours)*time.Hour)) {
				possible = true
				break
			}
		}
		if !possible {
			return nil
		}
	}
	for attempt := 0; attempt < 2; attempt++ {
		s, err := users.GetTrustState(userID)
		if err != nil {
			return err
		}
		if !trustlevel.Valid(s.TrustLevel) || s.EffectiveRestriction(now) != users.RestrictionNormal || s.NeedsEmailVerification(hotdataserve.GetSecuritySettingsConfigCache().EnableEmailVerification) {
			return nil
		}
		n := Next(s, c, now, topicID)
		if s.TrustLevel == n.TrustLevel && s.TrustVisitDays == n.TrustVisitDays && slices.Equal(s.TrustTopicIds, n.TrustTopicIds) && s.TrustLastVisitDay == n.TrustLastVisitDay {
			return nil
		}
		ok, err := users.UpdateTrust(s, n)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if s.TrustLevel != n.TrustLevel {
			userservice.InvalidateUserCaches(userID)
			if err := record(s, n, "auto", 0, ""); err != nil {
				slog.Warn("trust promotion history failed", "userId", userID, "error", err)
			}
			notifyPromotion(s, n)
		}
		return nil
	}
	return ErrConflict
}

func notifyPromotion(old, next users.TrustState) {
	if next.TrustLevel <= old.TrustLevel {
		return
	}
	if err := notificationservice.SendTrustPromotionNotification(next.Id, int16(next.TrustLevel)); err != nil {
		slog.Warn("trust promotion notification failed", "userId", next.Id, "level", next.TrustLevel, "error", err)
	}
}

func record(old, next users.TrustState, source string, actor uint64, reason string) error {
	return trusthistory.Create(&trusthistory.History{UserId: old.Id, FromLevel: old.TrustLevel, ToLevel: next.TrustLevel, FromLocked: old.TrustLevelLocked, ToLocked: next.TrustLevelLocked, Source: source, OperatorId: actor, Reason: reason})
}

func SetLevel(actor, userID uint64, level trustlevel.Level, locked bool, version uint64, reason string) error {
	if actor == userID || !trustlevel.Valid(level) || utf8.RuneCountInString(reason) > 500 || strings.TrimSpace(reason) == "" {
		return ErrInvalid
	}
	s, err := users.GetTrustState(userID)
	if err != nil {
		return err
	}
	if s.TrustVersion != version {
		return ErrConflict
	}
	if level < s.TrustLevel {
		return ErrInvalid
	}
	if level == s.TrustLevel && locked == s.TrustLevelLocked {
		return nil
	}
	n := s
	n.TrustLevel = level
	n.TrustLevelLocked = locked
	ok, err := users.UpdateTrust(s, n)
	if err != nil {
		return err
	}
	if !ok {
		return ErrConflict
	}
	userservice.InvalidateUserCaches(userID)
	reviewCache.Delete(strconv.FormatUint(userID, 10))
	notifyPromotion(s, n)
	if err := record(s, n, "admin", actor, strings.TrimSpace(reason)); err != nil {
		return errors.New("等级已更新，但历史记录写入失败")
	}
	return nil
}

type Progress struct {
	Limits                    pageConfig.TrustLimits          `json:"limits"`
	NextLimits                *pageConfig.TrustLimits         `json:"nextLimits"`
	RegistrationCooldown      pageConfig.RegistrationCooldown `json:"registrationCooldown"`
	CommunityUploadRoleExempt bool                            `json:"communityUploadRoleExempt"`
	CommunityUploadsEnabled   bool                            `json:"communityUploadsEnabled"`
	RestrictionStatus         string                          `json:"restrictionStatus"`
	Level                     trustlevel.Level                `json:"level"`
	Name                      string                          `json:"name"`
	Locked                    bool                            `json:"locked"`
	Version                   uint64                          `json:"version"`
	VisitDays                 uint32                          `json:"visitDays"`
	TopicsVisited             int                             `json:"topicsVisited"`
	CreatedAt                 time.Time                       `json:"createdAt"`
	Next                      *pageConfig.PromotionRule       `json:"next"`
	Definitions               []trustlevel.Definition         `json:"definitions"`
}

func GetProgress(id uint64) (Progress, error) {
	s, err := users.GetTrustState(id)
	if err != nil {
		return Progress{}, err
	}
	if !trustlevel.Valid(s.TrustLevel) {
		return Progress{}, ErrInvalid
	}
	c, err := hotdataserve.GetTrustSettings()
	if err != nil {
		return Progress{}, err
	}
	p := Progress{Level: s.TrustLevel, Name: trustlevel.Name(s.TrustLevel), Locked: s.TrustLevelLocked, Version: s.TrustVersion, VisitDays: s.TrustVisitDays, TopicsVisited: len(s.TrustTopicIds), CreatedAt: s.CreatedAt, Definitions: trustlevel.Definitions()}
	for _, r := range c.PromotionRules {
		if r.Enabled && r.Level > s.TrustLevel && (p.Next == nil || r.Level < p.Next.Level) {
			p.Next = &r
		}
	}
	p.RegistrationCooldown = c.RegistrationCooldown
	p.CommunityUploadRoleExempt = s.RoleId > 0
	p.CommunityUploadsEnabled = hotdataserve.GetPostingSettingsConfigCache().UploadControl.AllowAttachments || p.CommunityUploadRoleExempt
	p.RestrictionStatus = s.EffectiveRestriction(time.Now())
	for _, limits := range c.LevelLimits {
		if limits.Level == s.TrustLevel {
			p.Limits = limits
		}
		if p.Next != nil && limits.Level == p.Next.Level {
			next := limits
			p.NextLimits = &next
		}
	}
	return p, nil
}
