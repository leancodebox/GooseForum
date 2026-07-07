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

func shouldGrant(code string, stats userStatistics.Entity) bool {
	switch code {
	case CodeFirstPost:
		return stats.TopicCount >= 1
	case CodeWriter10:
		return stats.TopicCount >= 10
	case CodeFirstComment:
		return stats.ReplyCount >= 1
	case CodeCommenter50:
		return stats.ReplyCount >= 50
	case CodeFirstLikeGiven:
		return stats.LikeGivenCount >= 1
	case CodeLiked10:
		return stats.LikeReceivedCount >= 10
	case CodePopular100:
		return stats.LikeReceivedCount >= 100
	case CodeFirstFollower:
		return stats.FollowerCount >= 1
	case CodeSocial10:
		return stats.FollowerCount >= 10
	default:
		return false
	}
}
