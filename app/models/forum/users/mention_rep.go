package users

import (
	"strings"
	"time"
)

type MentionIdentity struct {
	Id                uint64     `json:"-"`
	Username          string     `json:"username"`
	IsFrozen          int8       `json:"-"`
	RestrictionStatus string     `json:"-"`
	RestrictionUntil  *time.Time `json:"-"`
}

func (user MentionIdentity) EffectiveRestriction(now time.Time) string {
	return AccountState{IsFrozen: user.IsFrozen, RestrictionStatus: user.RestrictionStatus, RestrictionUntil: user.RestrictionUntil}.EffectiveRestriction(now)
}

func MentionIdentities(ids []uint64, names []string) ([]MentionIdentity, error) {
	var result []MentionIdentity
	if len(ids)+len(names) == 0 {
		return result, nil
	}
	keys := make([]string, len(names))
	for i, name := range names {
		keys[i] = identityKey(name)
	}
	err := builder().Model(&EntityComplete{}).Where("id IN ? OR username IN ?", ids, keys).Find(&result).Error
	return result, err
}
func MentionCandidates(prefix string) ([]MentionIdentity, error) {
	var result []MentionIdentity
	prefix = strings.ToLower(prefix)
	err := builder().Model(&EntityComplete{}).Where("username >= ? AND username < ?", prefix, prefix+"~").Where("restriction_status IN ? OR restriction_status = '' OR restriction_until <= ?", []string{RestrictionNormal, RestrictionSuspended}, time.Now()).Order("username ASC").Limit(8).Find(&result).Error
	return result, err
}
