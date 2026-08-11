package topicservice

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

type FirstPostWrite struct {
	Topic       *topics.Entity
	FirstPost   *posts.Entity
	CategoryIDs []uint64
	Create      bool
}

func SaveTopicAndFirstPost(input FirstPostWrite) error {
	return SaveTopicAndFirstPostWithDB(dbconnect.Connect(), input)
}

func SaveTopicAndFirstPostWithDB(conn *gorm.DB, input FirstPostWrite) error {
	if conn == nil || input.Topic == nil || input.FirstPost == nil {
		return errors.New("topic and first post are required")
	}
	if len(input.CategoryIDs) == 0 {
		return errors.New("topic categories are required")
	}
	input.Topic.CategoryIds = append([]uint64(nil), input.CategoryIDs...)
	return conn.Transaction(func(tx *gorm.DB) error {
		if input.Create {
			input.Topic.PostCount = 1
			input.Topic.PostSeq = 1
			if err := tx.Create(input.Topic).Error; err != nil {
				return err
			}
			input.FirstPost.TopicId = input.Topic.Id
			input.FirstPost.PostNo = 1
			if err := tx.Create(input.FirstPost).Error; err != nil {
				return err
			}
			input.Topic.FirstPostId = input.FirstPost.Id
			input.Topic.LastPostId = input.FirstPost.Id
			now := time.Now()
			input.Topic.LastPostedAt = &now
			if err := tx.Save(input.Topic).Error; err != nil {
				return err
			}
		} else {
			if input.Topic.Id == 0 || input.FirstPost.Id == 0 || input.FirstPost.TopicId != input.Topic.Id {
				return errors.New("existing topic and first post are inconsistent")
			}
			if err := tx.Save(input.Topic).Error; err != nil {
				return err
			}
			if err := tx.Save(input.FirstPost).Error; err != nil {
				return err
			}
		}
		return topicCategoryIndex.ReplaceTopicCategoriesWithDB(tx, input.Topic.Id, input.CategoryIDs)
	})
}

func SaveTopicCategories(topic *topics.Entity, categoryIDs []uint64) error {
	if topic == nil || topic.Id == 0 || len(categoryIDs) == 0 {
		return errors.New("existing topic and categories are required")
	}
	topic.CategoryIds = append([]uint64(nil), categoryIDs...)
	return dbconnect.Connect().Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("updated_at").Save(topic).Error; err != nil {
			return err
		}
		return topicCategoryIndex.ReplaceTopicCategoriesWithDB(tx, topic.Id, categoryIDs)
	})
}
