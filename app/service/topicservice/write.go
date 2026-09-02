package topicservice

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/service/accesscontrol"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	categoryIDs, err := accesscontrol.CanonicalTopicCategoryIDs(input.CategoryIDs)
	if err != nil {
		return err
	}
	oldPublished := false
	oldCategoryIDs := []uint64{}
	if err := conn.Transaction(func(tx *gorm.DB) error {
		if !input.Create {
			var stored topics.Entity
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id", "status", "process_status").
				First(&stored, input.Topic.Id).Error; err != nil {
				return err
			}
			oldPublished = isCountedTopic(&stored)
			var err error
			oldCategoryIDs, err = topicCategoryIndex.ActiveCategoryIDsByTopicWithDB(tx, input.Topic.Id)
			if err != nil {
				return err
			}
		}
		if err := accesscontrol.ValidateRestrictedCategorySelectionWithDB(tx, categoryIDs); err != nil {
			return err
		}
		input.Topic.CategoryIds = append([]uint64(nil), categoryIDs...)
		input.Topic.MainCategoryId = categoryIDs[0]
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
		return topicCategoryIndex.ReplaceTopicCategoriesWithDB(tx, input.Topic.Id, categoryIDs)
	}); err != nil {
		return err
	}
	return adjustCategoryTopicCounts(conn, oldPublished, oldCategoryIDs, isCountedTopic(input.Topic), categoryIDs)
}

func SaveTopicCategories(topic *topics.Entity, categoryIDs []uint64) error {
	return SaveTopicCategoriesWithDB(dbconnect.Connect(), topic, categoryIDs)
}

func SaveTopicCategoriesWithDB(conn *gorm.DB, topic *topics.Entity, categoryIDs []uint64) error {
	if topic == nil || topic.Id == 0 {
		return errors.New("existing topic and categories are required")
	}
	categoryIDs, err := accesscontrol.CanonicalTopicCategoryIDs(categoryIDs)
	if err != nil {
		return err
	}
	oldPublished := false
	oldCategoryIDs := []uint64{}
	if err := conn.Transaction(func(tx *gorm.DB) error {
		var stored topics.Entity
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "status", "process_status").
			First(&stored, topic.Id).Error; err != nil {
			return err
		}
		oldPublished = isCountedTopic(&stored)
		var err error
		oldCategoryIDs, err = topicCategoryIndex.ActiveCategoryIDsByTopicWithDB(tx, topic.Id)
		if err != nil {
			return err
		}
		if err := accesscontrol.ValidateRestrictedCategorySelectionWithDB(tx, categoryIDs); err != nil {
			return err
		}
		topic.CategoryIds = append([]uint64(nil), categoryIDs...)
		topic.MainCategoryId = categoryIDs[0]
		if err := tx.Omit("updated_at").Save(topic).Error; err != nil {
			return err
		}
		return topicCategoryIndex.ReplaceTopicCategoriesWithDB(tx, topic.Id, categoryIDs)
	}); err != nil {
		return err
	}
	return adjustCategoryTopicCounts(conn, oldPublished, oldCategoryIDs, isCountedTopic(topic), categoryIDs)
}

func UpdateTopicStatus(topic *topics.Entity, nextStatus int8) error {
	return UpdateTopicStatusWithDB(dbconnect.Connect(), topic, nextStatus)
}

func UpdateTopicStatusWithDB(conn *gorm.DB, topic *topics.Entity, nextStatus int8) error {
	if topic == nil || topic.Id == 0 {
		return errors.New("existing topic is required")
	}
	previousStatus := topic.Status
	if previousStatus == nextStatus {
		return nil
	}
	wasCounted := isCountedTopic(topic)
	result := conn.Model(&topics.Entity{}).
		Where("id = ? AND status = ?", topic.Id, previousStatus).
		Update("status", nextStatus)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return refreshConcurrentTopicStatus(conn, topic, nextStatus, false)
	}
	topic.Status = nextStatus
	return adjustCategoryTopicCounts(conn, wasCounted, topic.CategoryIds, isCountedTopic(topic), topic.CategoryIds)
}

func UpdateTopicProcessStatus(topic *topics.Entity, nextStatus int8) error {
	return UpdateTopicProcessStatusWithDB(dbconnect.Connect(), topic, nextStatus)
}

func UpdateTopicProcessStatusWithDB(conn *gorm.DB, topic *topics.Entity, nextStatus int8) error {
	if topic == nil || topic.Id == 0 {
		return errors.New("existing topic is required")
	}
	previousStatus := topic.ProcessStatus
	if previousStatus == nextStatus {
		return nil
	}
	wasCounted := isCountedTopic(topic)
	result := conn.Model(&topics.Entity{}).
		Where("id = ? AND process_status = ?", topic.Id, previousStatus).
		UpdateColumn("process_status", nextStatus)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return refreshConcurrentTopicStatus(conn, topic, nextStatus, true)
	}
	topic.ProcessStatus = nextStatus
	return adjustCategoryTopicCounts(conn, wasCounted, topic.CategoryIds, isCountedTopic(topic), topic.CategoryIds)
}

func refreshConcurrentTopicStatus(conn *gorm.DB, topic *topics.Entity, expected int8, processStatus bool) error {
	var stored topics.Entity
	if err := conn.Select("id", "status", "process_status").First(&stored, topic.Id).Error; err != nil {
		return err
	}
	if processStatus {
		topic.ProcessStatus = stored.ProcessStatus
		if stored.ProcessStatus != expected {
			return errors.New("topic moderation status changed concurrently")
		}
		return nil
	}
	topic.Status = stored.Status
	if stored.Status != expected {
		return errors.New("topic publication status changed concurrently")
	}
	return nil
}

func DeleteTopic(topic *topics.Entity) error {
	return DeleteTopicWithDB(dbconnect.Connect(), topic)
}

func DeleteTopicWithDB(conn *gorm.DB, topic *topics.Entity) error {
	if topic == nil || topic.Id == 0 {
		return errors.New("existing topic is required")
	}
	var stored topics.Entity
	if err := conn.Select("id", "status", "process_status").First(&stored, topic.Id).Error; err != nil {
		return err
	}
	wasCounted := isCountedTopic(&stored)
	categoryIDs, err := topicCategoryIndex.ActiveCategoryIDsByTopicWithDB(conn, topic.Id)
	if err != nil {
		return err
	}
	if _, err := topicCategoryIndex.DeleteByTopicIdWithDB(conn, topic.Id); err != nil {
		return err
	}
	if result := conn.Delete(topic); result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return errors.New("delete topic returned no affected rows")
	}
	return adjustCategoryTopicCounts(conn, wasCounted, categoryIDs, false, nil)
}

func isCountedTopic(topic *topics.Entity) bool {
	return topic != nil && topic.Status == 1 && topic.ProcessStatus == 0 && !topic.DeletedAt.Valid
}

func adjustCategoryTopicCounts(conn *gorm.DB, oldCounted bool, oldIDs []uint64, newCounted bool, newIDs []uint64) error {
	oldSet := make(map[uint64]struct{}, len(oldIDs))
	newSet := make(map[uint64]struct{}, len(newIDs))
	if oldCounted {
		for _, id := range oldIDs {
			if id > 0 {
				oldSet[id] = struct{}{}
			}
		}
	}
	if newCounted {
		for _, id := range newIDs {
			if id > 0 {
				newSet[id] = struct{}{}
			}
		}
	}
	incrementIDs := make([]uint64, 0, len(newSet))
	decrementIDs := make([]uint64, 0, len(oldSet))
	for id := range newSet {
		if _, existed := oldSet[id]; !existed {
			incrementIDs = append(incrementIDs, id)
		}
	}
	for id := range oldSet {
		if _, remains := newSet[id]; !remains {
			decrementIDs = append(decrementIDs, id)
		}
	}
	return category.AdjustTopicCountsWithDB(conn, incrementIDs, decrementIDs)
}
