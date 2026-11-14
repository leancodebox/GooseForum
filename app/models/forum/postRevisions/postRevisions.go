package postRevisions

import (
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/pageutil"
	"gorm.io/gorm"
)

const tableName = "post_revisions"

// Entity is an immutable source-content snapshot. It is intentionally separate
// from posts.Entity so renderer changes and attachment metadata do not rewrite
// historical source content.
type Entity struct {
	Id                 uint64    `gorm:"primaryKey;column:id;autoIncrement;not null" json:"id"`
	PostId             uint64    `gorm:"column:post_id;not null;index:idx_post_revisions_post_created,priority:1;uniqueIndex:uniq_post_revision,priority:1" json:"postId"`
	RevisionNo         uint64    `gorm:"column:revision_no;not null;uniqueIndex:uniq_post_revision,priority:2" json:"revisionNo"`
	AuthorId           uint64    `gorm:"column:author_id;not null;default:0;index:idx_post_revisions_author" json:"authorId"`
	Content            string    `gorm:"column:content;type:text;not null" json:"content"`
	SourceVersion      uint8     `gorm:"column:source_version;not null;default:0" json:"sourceVersion"`
	ModerationVersion  uint64    `gorm:"column:moderation_version;not null;default:0" json:"moderationVersion"`
	VisibilitySnapshot string    `gorm:"column:visibility_snapshot;type:varchar(32);not null;default:''" json:"visibilitySnapshot"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime;not null;index:idx_post_revisions_post_created,priority:2,sort:desc" json:"createdAt"`
}

// SnapshotInput is the stable, source-only data needed to append a revision.
// Keeping this DTO independent from posts.Entity lets multiple write services
// share the insert boundary without creating an import cycle.
type SnapshotInput struct {
	PostId             uint64
	RevisionNo         uint64
	AuthorId           uint64
	Content            string
	SourceVersion      uint8
	ModerationVersion  uint64
	VisibilitySnapshot string
}

func (itself *Entity) TableName() string { return tableName }

func CreateSnapshotWithDB(db *gorm.DB, input SnapshotInput) error {
	return db.Create(&Entity{
		PostId:             input.PostId,
		RevisionNo:         input.RevisionNo,
		AuthorId:           input.AuthorId,
		Content:            input.Content,
		SourceVersion:      input.SourceVersion,
		ModerationVersion:  input.ModerationVersion,
		VisibilitySnapshot: input.VisibilitySnapshot,
	}).Error
}

type PageQuery struct {
	PostId   uint64
	Cursor   uint64
	PageSize uint64
}

type Page struct {
	Items      []Entity `json:"items"`
	NextCursor uint64   `json:"nextCursor"`
	HasMore    bool     `json:"hasMore"`
}

func ListPageWithDB(db *gorm.DB, query PageQuery) (Page, error) {
	if db == nil || query.PostId == 0 {
		return Page{}, gorm.ErrInvalidData
	}
	query.PageSize = pageutil.DefaultPageSize(query.PageSize, 20)
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	pageSize := int(query.PageSize)
	b := db.Model(&Entity{}).Where("post_id = ?", query.PostId)
	if query.Cursor > 0 {
		b = b.Where("revision_no < ?", query.Cursor)
	}
	var items []Entity
	if err := b.Order("revision_no DESC").Limit(pageSize + 1).Find(&items).Error; err != nil {
		return Page{}, err
	}
	items, nextCursor, hasMore := pageutil.TrimCursorPage(items, pageSize, func(item Entity) uint64 {
		return item.RevisionNo
	})
	return Page{Items: items, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func GetWithDB(db *gorm.DB, postID, revisionNo uint64) (Entity, error) {
	if db == nil || postID == 0 || revisionNo == 0 {
		return Entity{}, gorm.ErrInvalidData
	}
	var entity Entity
	err := db.Where("post_id = ? AND revision_no = ?", postID, revisionNo).First(&entity).Error
	return entity, err
}

func VisibilitySnapshot(processStatus int8) string {
	if processStatus == 0 {
		return "visible"
	}
	return "hidden"
}
