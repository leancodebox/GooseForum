package posts

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

type MarkdownSource struct {
	Id            uint64
	Content       string
	SourceVersion uint8
}

func PendingMarkdownBatch(afterID uint64, version uint32, limit int) ([]MarkdownSource, error) {
	return PendingMarkdownBatchWithDB(dbconnect.Connect(), afterID, version, limit)
}

func PendingMarkdownBatchWithDB(db *gorm.DB, afterID uint64, version uint32, limit int) ([]MarkdownSource, error) {
	var batch []MarkdownSource
	err := db.Model(&Entity{}).Where("id > ?", afterID).
		Where("rendered_html = ? OR rendered_version < ?", "", version).
		Order("id ASC").Limit(limit).Find(&batch).Error
	return batch, err
}

func UpdateRenderedMarkdown(id uint64, html string, version uint32) error {
	return UpdateRenderedMarkdownWithDB(dbconnect.Connect(), id, html, version)
}

func UpdateRenderedMarkdownWithDB(db *gorm.DB, id uint64, html string, version uint32) error {
	return db.Model(&Entity{}).Where("id = ?", id).
		UpdateColumns(map[string]any{"rendered_html": html, "rendered_version": version}).Error
}
