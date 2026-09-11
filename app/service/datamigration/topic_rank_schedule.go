package datamigration

import (
	"context"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
)

// MigrateTopicRankSchedule copies only existing deadlines. No scores or source
// counters are recomputed. Drop the legacy column only after every batch succeeds.
func MigrateTopicRankSchedule(ctx context.Context) error {
	db := dbconnect.Connect().WithContext(ctx)
	if err := ctx.Err(); err != nil {
		return err
	}
	columns, err := db.Migrator().ColumnTypes(&topics.Entity{})
	if err != nil {
		return err
	}
	hasLegacy := false
	for _, column := range columns {
		if column.Name() == "next_rank_at" {
			hasLegacy = true
			break
		}
	}
	if !hasLegacy {
		return nil
	}
	upper, err := topics.MaxRankingID(ctx)
	if err != nil {
		return err
	}
	var cursor uint64
	for {
		var rows []struct {
			ID         uint64     `gorm:"column:id"`
			NextRankAt *time.Time `gorm:"column:next_rank_at"`
		}
		if err := db.Table("topics").Select("id", "next_rank_at").
			Where(queryopt.Gt("id", cursor)).Where(queryopt.Le("id", upper)).
			Where(queryopt.IsNotNull("next_rank_at")).Order(queryopt.Asc("id")).Limit(200).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		entries := make([]topicrank.Entity, 0, len(rows))
		for _, row := range rows {
			entries = append(entries, topicrank.Entity{TopicID: row.ID, NextRunAt: row.NextRankAt, Version: 1})
		}
		if err := topicrank.Import(ctx, entries); err != nil {
			return err
		}
		cursor = rows[len(rows)-1].ID
	}
	if db.Migrator().HasIndex(&topics.Entity{}, "idx_topics_rank_due") {
		if err := db.Migrator().DropIndex(&topics.Entity{}, "idx_topics_rank_due"); err != nil {
			return err
		}
	}
	// Native DROP COLUMN works on supported SQLite/MySQL versions. The SQLite
	// ORM fallback recreates the table and drops unrelated secondary indexes.
	return db.Exec("ALTER TABLE topics DROP COLUMN next_rank_at").Error
}
