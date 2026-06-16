package moderationLog

import "github.com/leancodebox/GooseForum/app/bundles/queryopt"

func Create(entity *Entity) error {
	return builder().Create(entity).Error
}

type CursorPageQuery struct {
	Cursor, PageSize uint64
}

func CursorPage(q CursorPageQuery) []Entity {
	var list []Entity
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	b := builder()
	if q.Cursor > 0 {
		b.Where(queryopt.Lt("id", q.Cursor))
	}
	b.Limit(int(q.PageSize)).Order(queryopt.Desc("id")).Find(&list)
	return list
}
