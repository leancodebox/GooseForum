package badgeservice

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/userBadges"
	"github.com/leancodebox/GooseForum/app/models/forum/userStatistics"
)

type Trigger string

const (
	TriggerSignUp  Trigger = "sign_up"
	TriggerPost    Trigger = "post"
	TriggerComment Trigger = "comment"
	TriggerLike    Trigger = "like"
	TriggerFollow  Trigger = "follow"
)

func CheckAndGrant(userID uint64, trigger Trigger) {
	if userID == 0 {
		return
	}
	stats := userStatistics.Get(userID)
	for _, code := range candidateCodes(trigger) {
		if !shouldGrant(code, stats) {
			continue
		}
		granted, err := Grant(userID, code, userBadges.SourceAuto, "", 0)
		if err != nil {
			slog.Debug("badge auto grant failed", "user_id", userID, "badge", code, "err", err)
			continue
		}
		if granted {
			slog.Debug("badge auto granted", "user_id", userID, "badge", code)
		}
	}
}

func candidateCodes(trigger Trigger) []string {
	switch trigger {
	case TriggerPost:
		return []string{CodeFirstPost, CodeWriter10}
	case TriggerComment:
		return []string{CodeFirstComment, CodeCommenter50}
	case TriggerLike:
		return []string{CodeFirstLikeGiven, CodeLiked10, CodePopular100}
	case TriggerFollow:
		return []string{CodeFirstFollower, CodeSocial10}
	default:
		return []string{}
	}
}

// GrantRule describes the same fixed rule used by the event handler.
type GrantRule struct {
	Metric    string `json:"metric"`
	Threshold int64  `json:"threshold"`
}

func RuleForCode(code string) *GrantRule {
	var rule GrantRule
	switch code {
	case CodeFirstPost:
		rule = GrantRule{"topics", 1}
	case CodeWriter10:
		rule = GrantRule{"topics", 10}
	case CodeFirstComment:
		rule = GrantRule{"replies", 1}
	case CodeCommenter50:
		rule = GrantRule{"replies", 50}
	case CodeFirstLikeGiven:
		rule = GrantRule{"likesGiven", 1}
	case CodeLiked10:
		rule = GrantRule{"likesReceived", 10}
	case CodePopular100:
		rule = GrantRule{"likesReceived", 100}
	case CodeFirstFollower:
		rule = GrantRule{"followers", 1}
	case CodeSocial10:
		rule = GrantRule{"followers", 10}
	default:
		return nil
	}
	return &rule
}

func shouldGrant(code string, stats userStatistics.Entity) bool {
	rule := RuleForCode(code)
	if rule == nil {
		return false
	}
	var count int64
	switch rule.Metric {
	case "topics":
		count = int64(stats.TopicCount)
	case "replies":
		count = int64(stats.ReplyCount)
	case "likesGiven":
		count = int64(stats.LikeGivenCount)
	case "likesReceived":
		count = int64(stats.LikeReceivedCount)
	case "followers":
		count = int64(stats.FollowerCount)
	}
	return count >= rule.Threshold
}
