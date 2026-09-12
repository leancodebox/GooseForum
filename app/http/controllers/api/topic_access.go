package api

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

var ErrTopicUnavailable = errors.New("topic is unavailable")

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
