package reports

import (
	"errors"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"gorm.io/gorm"
)

func CreateOpen(entity Entity) (Entity, bool, error) {
	var existing Entity
	err := builder().
		Where(queryopt.Eq(fieldReporterId, entity.ReporterId)).
		Where(queryopt.Eq(fieldTargetType, entity.TargetType)).
		Where(queryopt.Eq(fieldTargetId, entity.TargetId)).
		Where(queryopt.Eq(fieldStatus, StatusOpen)).
		First(&existing).Error
	if err == nil && existing.Id > 0 {
		return existing, false, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return Entity{}, false, err
	}
	entity.Status = StatusOpen
	if err := builder().Create(&entity).Error; err != nil {
		return Entity{}, false, err
	}
	return entity, true, nil
}

func Get(id uint64) (entity Entity) {
	if id == 0 {
		return entity
	}
	builder().First(&entity, id)
	return
}

type CursorPageQuery struct {
	TargetType       string
	Status           string
	Statuses         []string
	ScopeCategoryIDs []uint64
	Cursor, PageSize uint64
}

func CursorPage(q CursorPageQuery) []Entity {
	var list []Entity
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	b := builder()
	if q.Status != "" {
		b = b.Where(queryopt.Eq(fieldStatus, q.Status))
	} else if len(q.Statuses) > 0 {
		b = b.Where(queryopt.In(fieldStatus, q.Statuses))
	}
	if q.TargetType != "" {
		b = b.Where(queryopt.Eq(fieldTargetType, q.TargetType))
	}
	if len(q.ScopeCategoryIDs) > 0 {
		b = b.Where(`EXISTS (
			SELECT 1 FROM topic_category_index idx
			WHERE idx.topic_id = reports.topic_id
				AND idx.category_id IN ?
				AND idx.effective = ?
		)`, q.ScopeCategoryIDs, 1)
	}
	if q.Cursor > 0 {
		b = b.Where(queryopt.Lt("id", q.Cursor))
	}
	b.Limit(int(q.PageSize)).Order(queryopt.Desc("id")).Find(&list)
	return list
}

func UpdateStatus(id uint64, status string, resolution string, handlerId uint64) error {
	now := time.Now()
	return builder().Where(queryopt.Eq("id", id)).Updates(map[string]any{
		"status":     status,
		"resolution": resolution,
		"handler_id": handlerId,
		"handled_at": &now,
	}).Error
}
