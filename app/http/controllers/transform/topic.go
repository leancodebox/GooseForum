package transform

import (
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/vo"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/urlconfig"
)

func Topics2Vo(data []*topics.Entity, categoryMap map[uint64]*category.Entity) []*vo.TopicsSimpleVo {
	userIDs := make([]uint64, 0, len(data)*2)
	seenUserIDs := make(map[uint64]struct{}, len(data)*2)
	for _, topic := range data {
		if topic == nil {
			continue
		}
		if _, ok := seenUserIDs[topic.UserId]; !ok {
			seenUserIDs[topic.UserId] = struct{}{}
			userIDs = append(userIDs, topic.UserId)
		}
		for _, poster := range topic.GetPosters() {
			if _, ok := seenUserIDs[poster.UserID]; ok {
				continue
			}
			seenUserIDs[poster.UserID] = struct{}{}
			userIDs = append(userIDs, poster.UserID)
		}
	}
	userMap := users.GetMapByIds(userIDs)
	return TopicsWithUser2Vo(data, categoryMap, userMap)
}

func TopicsWithUser2Vo(data []*topics.Entity, categoryMap map[uint64]*category.Entity, userMap map[uint64]*users.EntityComplete) []*vo.TopicsSimpleVo {
	res := make([]*vo.TopicsSimpleVo, 0, len(data))
	for _, t := range data {
		if t == nil {
			continue
		}

		categoryNames := make([]string, 0, len(t.CategoryIds))
		for _, item := range t.CategoryIds {
			if category, ok := categoryMap[item]; ok && category != nil {
				categoryNames = append(categoryNames, category.Name)
				continue
			}
			categoryNames = append(categoryNames, "")
		}

		username := ""
		avatarUrl := urlconfig.GetDefaultAvatar()
		if user, ok := userMap[t.UserId]; ok {
			username = user.Username
			avatarUrl = user.GetWebAvatarUrl()
		}

		posters := t.GetPosters()
		postersVo := make([]vo.PosterVo, 0, len(posters))
		for _, poster := range posters {
			posterUsername := ""
			posterAvatarUrl := urlconfig.GetDefaultAvatar()
			if user, ok := userMap[poster.UserID]; ok {
				posterUsername = user.Username
				posterAvatarUrl = user.GetWebAvatarUrl()
			}
			postersVo = append(postersVo, vo.PosterVo{
				Id:        poster.UserID,
				Username:  posterUsername,
				AvatarUrl: posterAvatarUrl,
			})
		}

		res = append(res, &vo.TopicsSimpleVo{
			Id:             t.Id,
			Title:          t.Title,
			Description:    t.Excerpt,
			FirstImageURL:  t.FirstImageURL,
			LastUpdateTime: t.UpdatedAt.Format(time.DateTime),
			CreateTime:     t.CreatedAt.Format(time.DateTime),
			AuthorId:       t.UserId,
			Username:       username,
			AvatarUrl:      avatarUrl,
			ViewCount:      t.ViewCount,
			CommentCount:   t.ReplyCount,
			PinWeight:      t.PinWeight,
			Categories:     categoryNames,
			CategoriesId:   t.CategoryIds,
			ProcessStatus:  t.ProcessStatus,
			Posters:        postersVo,
			LastPostId:     t.LastPostId,
			LastPostedAt:   timeValue(t.LastPostedAt),
		})
	}
	return res
}

func timeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
