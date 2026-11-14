package topicaccessservice

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

var ErrUnavailable = errors.New("topic unavailable")

// AuthorizePublished checks the shared read/reply/create/manage rule for a
// published topic. Callers may translate ErrUnavailable into their own API or
// domain error without duplicating the visibility and category checks.
func AuthorizePublished(userID uint64, topic topics.Entity, required accesscontrol.Capability) error {
	if topic.Id == 0 || !topic.IsPublished() {
		return ErrUnavailable
	}
	snapshot, err := accesscontrol.Resolve(userID)
	if err != nil {
		return err
	}
	var allowed bool
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
		return ErrUnavailable
	}
	return nil
}
