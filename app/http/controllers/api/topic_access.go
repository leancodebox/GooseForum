package api

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"github.com/leancodebox/GooseForum/app/service/topicaccessservice"
)

var ErrTopicUnavailable = errors.New("topic is unavailable")

func authorizePublishedTopic(userID uint64, topic topics.Entity, required accesscontrol.Capability) error {
	if err := topicaccessservice.AuthorizePublished(userID, topic, required); err != nil {
		if errors.Is(err, topicaccessservice.ErrUnavailable) {
			return ErrTopicUnavailable
		}
		return err
	}
	return nil
}
