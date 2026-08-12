package api

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

var ErrTopicUnavailable = errors.New("topic is unavailable")

func authorizeTopicCategoryWrite(userID uint64, topic *topics.Entity, next []uint64, newTopic bool, publishing bool) ([]uint64, error) {
	actor, err := accesscontrol.Resolve(userID)
	if err != nil {
		return nil, err
	}
	everyone, err := accesscontrol.Resolve(0)
	if err != nil {
		return nil, err
	}
	var current []uint64
	if topic != nil {
		current = topic.CategoryIds
	}
	categoryIDs, err := accesscontrol.ValidateTopicCategoryWrite(actor, everyone, accesscontrol.TopicCategoryWrite{
		Current:    current,
		Next:       next,
		Publishing: publishing,
		NewTopic:   newTopic,
	})
	if err != nil {
		return nil, err
	}
	for _, categoryID := range categoryIDs {
		if category.Get(categoryID).Id == 0 {
			return nil, accesscontrol.ErrCategoryPermissionDenied
		}
	}
	return categoryIDs, nil
}

func authorizePublishedTopic(userID uint64, topic topics.Entity, required accesscontrol.Capability) error {
	if topic.Id == 0 || topic.Status != 1 || topic.ProcessStatus != 0 {
		return ErrTopicUnavailable
	}
	snapshot, err := accesscontrol.Resolve(userID)
	if err != nil {
		return err
	}
	allowed := false
	switch required {
	case accesscontrol.CapabilityReply:
		allowed = snapshot.CanReplyCategory(topic.MainCategoryId)
	case accesscontrol.CapabilityCreate:
		allowed = snapshot.CanCreateCategory(topic.MainCategoryId)
	case accesscontrol.CapabilityManage:
		allowed = snapshot.CanManageCategory(topic.MainCategoryId)
	default:
		allowed = snapshot.CanReadCategory(topic.MainCategoryId)
	}
	if !allowed {
		return ErrTopicUnavailable
	}
	return nil
}
