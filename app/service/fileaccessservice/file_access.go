package fileaccessservice

import (
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
)

type Decision struct {
	Allowed bool
	Public  bool
}

func Resolve(userID uint64, fileName string, uploaderID uint64) (Decision, error) {
	usages, err := fileUsage.GetByFileName(fileName)
	if err != nil {
		return Decision{}, err
	}
	// Existing untracked uploads predate file_usage. Keep them public for
	// compatibility; newly linked topic images are governed below.
	if len(usages) == 0 {
		return Decision{Allowed: true, Public: true}, nil
	}
	topicIDs := make(map[uint64]struct{})
	postIDs := make([]uint64, 0)
	for _, usage := range usages {
		switch usage.TargetType {
		case fileUsage.TargetTopic:
			topicIDs[usage.TargetId] = struct{}{}
		case fileUsage.TargetPost:
			postIDs = append(postIDs, usage.TargetId)
		case fileUsage.TargetPendingUpload, fileUsage.TargetUploadOwner:
			continue
		default:
			return Decision{Allowed: true, Public: true}, nil
		}
	}
	for _, post := range posts.GetByIds(postIDs) {
		if post != nil && post.TopicId != 0 {
			topicIDs[post.TopicId] = struct{}{}
		}
	}
	topicIDList := make([]uint64, 0, len(topicIDs))
	for topicID := range topicIDs {
		topicIDList = append(topicIDList, topicID)
	}
	topicMap := topics.GetMapByIds(topicIDList)
	guest, err := accesscontrol.Resolve(0)
	if err != nil {
		return Decision{}, err
	}
	viewer, err := accesscontrol.Resolve(userID)
	if err != nil {
		return Decision{}, err
	}
	for _, topic := range topicMap {
		if topic.Id == 0 {
			continue
		}
		if topic.IsPublished() && guest.CanReadCategory(topic.MainCategoryId) {
			return Decision{Allowed: true, Public: true}, nil
		}
		if userID != 0 && ((topic.IsPublished() && viewer.CanReadCategory(topic.MainCategoryId)) || topic.UserId == userID || viewer.CanManageAnyCategory(topic.CategoryIds)) {
			return Decision{Allowed: true}, nil
		}
	}
	return Decision{Allowed: userID != 0 && userID == uploaderID}, nil
}
