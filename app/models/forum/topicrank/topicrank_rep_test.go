package topicrank

import (
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestForumUpsertsAcrossDialects(t *testing.T) {
	tests := []struct {
		name, dsnEnv string
		open         func(string) gorm.Dialector
	}{
		{name: "sqlite", open: sqlite.Open},
		{name: "postgres", dsnEnv: "GOOSEFORUM_POSTGRES_TEST_DSN", open: postgres.Open},
		{name: "mysql", dsnEnv: "GOOSEFORUM_MYSQL_TEST_DSN", open: func(dsn string) gorm.Dialector {
			// The fixture uses second-precision columns for the model's CURRENT_TIMESTAMP default.
			return mysql.New(mysql.Config{DSN: dsn, DisableDatetimePrecision: true})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := ":memory:"
			if test.dsnEnv != "" {
				dsn = os.Getenv(test.dsnEnv)
				if dsn == "" {
					t.Skipf("%s is not set", test.dsnEnv)
				}
			}
			db, err := gorm.Open(test.open(dsn), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = sqlDB.Close() })
			if err := db.AutoMigrate(&Entity{}, &topicUserStat.Entity{}); err != nil {
				t.Fatal(err)
			}

			tx := db.Begin()
			if tx.Error != nil {
				t.Fatal(tx.Error)
			}
			defer tx.Rollback()

			topicID := uint64(time.Now().UnixNano())
			const userID = 2
			for reply := 1; reply <= 2; reply++ {
				if err := topicUserStat.IncrementUserPostWithDB(tx, topicID, userID); err != nil {
					t.Fatalf("record reply %d: %v", reply, err)
				}
			}
			var stat topicUserStat.Entity
			if err := tx.Where("topic_id = ? AND user_id = ?", topicID, userID).Take(&stat).Error; err != nil {
				t.Fatal(err)
			}
			if stat.ReplyCount != 2 {
				t.Fatalf("reply count = %d, want 2", stat.ReplyCount)
			}

			earliest := time.Date(2026, time.November, 22, 12, 0, 0, 0, time.UTC)
			for _, next := range []time.Time{earliest.Add(time.Hour), earliest, earliest.Add(time.Minute)} {
				if err := markAtWithDB(tx, topicID, next); err != nil {
					t.Fatalf("schedule at %s: %v", next, err)
				}
			}
			var schedule Entity
			if err := tx.Take(&schedule, "topic_id = ?", topicID).Error; err != nil {
				t.Fatal(err)
			}
			if schedule.NextRunAt == nil || !schedule.NextRunAt.Equal(earliest) || schedule.Version != 3 {
				t.Fatalf("schedule = %+v, want earliest time %s and version 3", schedule, earliest)
			}
		})
	}
}
