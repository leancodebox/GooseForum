package topics

import (
	"fmt"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	categoryBenchmarkTopicCount        = 100000
	categoryBenchmarkMainCategoryCount = 100
	categoryBenchmarkCommonCategoryID  = 101
	categoryBenchmarkRareCategoryID    = 102
)

func BenchmarkCategoryTopicList100K(b *testing.B) {
	benchmarkCategoryTopicList100K(b, ":memory:")
}

func BenchmarkCategoryTopicList100KFile(b *testing.B) {
	benchmarkCategoryTopicList100K(b, filepath.Join(b.TempDir(), "topics.db"))
}

func benchmarkCategoryTopicList100K(b *testing.B, dsn string) {
	b.Helper()
	conn := openCategoryBenchmarkDB(b, dsn)
	seedCategoryBenchmark(b, conn)

	b.Run("main_category_1_percent", func(b *testing.B) {
		benchmarkCategoryTopicQuery(b, conn, 1, []uint64{1})
	})
	for _, visiblePercent := range []int{95, 50, 5} {
		readable := categoryBenchmarkReadableIDs(visiblePercent, categoryBenchmarkCommonCategoryID)
		b.Run(fmt.Sprintf("common_auxiliary_visible_%d_percent", visiblePercent), func(b *testing.B) {
			benchmarkCategoryTopicQuery(b, conn, categoryBenchmarkCommonCategoryID, readable)
		})
	}
	b.Run("rare_category_10_oldest_topics", func(b *testing.B) {
		readable := categoryBenchmarkReadableIDs(categoryBenchmarkMainCategoryCount, categoryBenchmarkRareCategoryID)
		benchmarkCategoryTopicQuery(b, conn, categoryBenchmarkRareCategoryID, readable)
	})
}

func benchmarkCategoryTopicQuery(b *testing.B, conn *gorm.DB, categoryID uint64, readable []uint64) {
	b.Helper()
	b.ReportAllocs()
	samples := make([]time.Duration, b.N)
	for i := 0; i < b.N; i++ {
		started := time.Now()
		var result []Entity
		query := conn.Model(&Entity{}).
			Where("status = ? AND process_status = ?", 1, 0).
			Where(
				`EXISTS (SELECT 1 FROM topic_category_index idx WHERE idx.topic_id = topics.id AND idx.category_id = ? AND idx.effective = ?)`,
				categoryID,
				1,
			)
		query = applyAudienceFilter(query, readable, true)
		if err := query.
			Order("pin_weight DESC, updated_at DESC, id DESC").
			Limit(21).
			Find(&result).Error; err != nil {
			b.Fatal(err)
		}
		samples[i] = time.Since(started)
	}
	reportCategoryBenchmarkP95(b, samples)
}

func openCategoryBenchmarkDB(b *testing.B, dsn string) *gorm.DB {
	b.Helper()
	conn, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		b.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		b.Fatalf("get sqlite database: %v", err)
	}
	b.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			b.Errorf("close sqlite: %v", err)
		}
	})
	if err := conn.AutoMigrate(&Entity{}, &topicCategoryIndex.Entity{}); err != nil {
		b.Fatalf("migrate benchmark tables: %v", err)
	}
	return conn
}

func seedCategoryBenchmark(b *testing.B, conn *gorm.DB) {
	b.Helper()
	now := time.Now()
	for start := 1; start <= categoryBenchmarkTopicCount; start += 1000 {
		end := min(start+1000, categoryBenchmarkTopicCount+1)
		topicsBatch := make([]Entity, 0, end-start)
		indexBatch := make([]topicCategoryIndex.Entity, 0, (end-start)*2)
		for id := start; id < end; id++ {
			mainCategoryID := uint64((id-1)%categoryBenchmarkMainCategoryCount + 1)
			categoryIDs := []uint64{mainCategoryID, categoryBenchmarkCommonCategoryID}
			if id <= 10 {
				categoryIDs = append(categoryIDs, categoryBenchmarkRareCategoryID)
			}
			topicsBatch = append(topicsBatch, Entity{
				Id:             uint64(id),
				Title:          "benchmark",
				CategoryIds:    categoryIDs,
				MainCategoryId: mainCategoryID,
				Status:         1,
				UpdatedAt:      now.Add(time.Duration(id) * time.Microsecond),
			})
			for _, categoryID := range categoryIDs {
				indexBatch = append(indexBatch, topicCategoryIndex.Entity{
					TopicId: uint64(id), CategoryId: categoryID, Effective: 1,
				})
			}
		}
		if err := conn.Create(&topicsBatch).Error; err != nil {
			b.Fatalf("seed topics: %v", err)
		}
		if err := conn.Create(&indexBatch).Error; err != nil {
			b.Fatalf("seed category indexes: %v", err)
		}
	}
}

func categoryBenchmarkReadableIDs(visibleMainCategories int, listedCategoryID uint64) []uint64 {
	readable := make([]uint64, visibleMainCategories, visibleMainCategories+1)
	for i := range readable {
		readable[i] = uint64(i + 1)
	}
	return append(readable, listedCategoryID)
}

func reportCategoryBenchmarkP95(b *testing.B, samples []time.Duration) {
	b.Helper()
	slices.Sort(samples)
	if len(samples) == 0 {
		return
	}
	p95Index := min((len(samples)*95+99)/100-1, len(samples)-1)
	b.ReportMetric(float64(samples[p95Index].Nanoseconds()), "p95-ns")
}
