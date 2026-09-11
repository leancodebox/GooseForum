package datamigration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/topicrank"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

type legacyRankTopic struct {
	topics.Entity
	NextRankAt *time.Time `gorm:"column:next_rank_at;index:idx_topics_rank_due"`
}

func (legacyRankTopic) TableName() string { return "topics" }

func TestRankScheduleMigrationResumesThenDropsOnlyLegacyField(t *testing.T) {
	db := dbconnect.Connect()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(db.Migrator().DropTable(&topicrank.Entity{}, &topics.Entity{}))
	must(db.AutoMigrate(&legacyRankTopic{}, &topicrank.Entity{}))
	due := time.Now().Add(time.Hour).Truncate(time.Millisecond)
	created := due.Add(-48 * time.Hour)
	rows := make([]legacyRankTopic, 212)
	for i := range rows {
		rows[i] = legacyRankTopic{Entity: topics.Entity{Id: uint64(i + 1), Status: 1, Title: "preserved", CreatedAt: created, UpdatedAt: created}, NextRankAt: &due}
	}
	must(db.CreateInBatches(rows, 100).Error)
	must(db.Create(&legacyRankTopic{Entity: topics.Entity{Id: 213, Title: "settled"}}).Error)
	must(db.Table("topics").Where("id > 0").UpdateColumns(map[string]any{"rank_score": 17000, "published_at": created}).Error)
	earlier := due.Add(-time.Hour)
	must(db.Create(&topicrank.Entity{TopicID: 1, NextRunAt: &earlier, Version: 7}).Error)
	// Real schema migration must preserve the old column until data migration.
	must(db.AutoMigrate(&topics.Entity{}, &topicrank.Entity{}))
	if !db.Migrator().HasColumn(&topics.Entity{}, "next_rank_at") {
		t.Fatal("schema step lost the deadlines")
	}
	batches := 0
	must(db.Callback().Create().Before("gorm:create").Register("test:interrupt_schedule_copy", func(tx *gorm.DB) {
		if tx.Statement.Table == "topic_rank_schedule" {
			batches++
			if batches == 2 {
				tx.AddError(errors.New("interrupted second batch"))
			}
		}
	}))
	t.Cleanup(func() { db.Callback().Create().Remove("test:interrupt_schedule_copy") })
	if err := MigrateTopicRankSchedule(context.Background()); err == nil {
		t.Fatal("expected interrupted copy")
	}
	if !db.Migrator().HasColumn(&topics.Entity{}, "next_rank_at") || !db.Migrator().HasIndex(&topics.Entity{}, "idx_topics_rank_due") {
		t.Fatal("legacy schedule deleted before copying finished")
	}
	must(db.Callback().Create().Remove("test:interrupt_schedule_copy"))
	must(MigrateTopicRankSchedule(context.Background()))
	must(MigrateTopicRankSchedule(context.Background()))
	must(db.AutoMigrate(&topics.Entity{}, &topicrank.Entity{}))
	if db.Migrator().HasColumn(&topics.Entity{}, "next_rank_at") || db.Migrator().HasIndex(&topics.Entity{}, "idx_topics_rank_due") {
		t.Fatal("legacy field/index survived or were recreated")
	}
	for _, name := range []string{"idx_topics_list_rank", "idx_topics_list_default", "idx_topics_user_status"} {
		if !db.Migrator().HasIndex(&topics.Entity{}, name) {
			t.Fatalf("migration dropped unrelated index %s", name)
		}
	}
	var schedules []topicrank.Entity
	must(db.Order("topic_id ASC").Find(&schedules).Error)
	if len(schedules) != 212 {
		t.Fatalf("copied %d schedules", len(schedules))
	}
	if schedules[0].Version != 7 || !schedules[0].NextRunAt.Equal(earlier) {
		t.Fatal("rerun overwrote an existing schedule")
	}
	for _, row := range schedules[1:] {
		if row.Version != 1 || !row.NextRunAt.Equal(due) {
			t.Fatalf("wrong deadline/version: %+v", row)
		}
	}
	var topic topics.Entity
	must(db.First(&topic, 212).Error)
	if topic.RankScore != 17000 || topic.Title != "preserved" || !topic.CreatedAt.Equal(created) || !topic.UpdatedAt.Equal(created) || !topic.PublishedAt.Equal(created) {
		t.Fatalf("migration recalculated or altered topic: %+v", topic)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := MigrateTopicRankSchedule(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation ignored: %v", err)
	}
}

func TestRankScheduleMigrationOnFreshSchema(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.Migrator().DropTable(&topicrank.Entity{}, &topics.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&topics.Entity{}, &topicrank.Entity{}); err != nil {
		t.Fatal(err)
	}
	if err := MigrateTopicRankSchedule(context.Background()); err != nil {
		t.Fatal(err)
	}
	if db.Migrator().HasColumn(&topics.Entity{}, "next_rank_at") {
		t.Fatal("fresh schema has legacy field")
	}
}
