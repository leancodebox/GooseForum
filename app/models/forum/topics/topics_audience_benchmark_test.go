package topics

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func BenchmarkAudienceTopicList100K(b *testing.B) {
	conn, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		b.Fatalf("open sqlite: %v", err)
	}
	if err := conn.AutoMigrate(&Entity{}, &topicCategoryIndex.Entity{}); err != nil {
		b.Fatalf("migrate benchmark tables: %v", err)
	}
	seedAudienceBenchmark(b, conn, 100000, 100)

	for _, visiblePercent := range []int{95, 50, 5} {
		readable := make([]uint64, visiblePercent)
		for i := range readable {
			readable[i] = uint64(i + 1)
		}
		b.Run(fmt.Sprintf("visible_%d_percent", visiblePercent), func(b *testing.B) {
			b.ReportAllocs()
			samples := make([]time.Duration, b.N)
			for i := 0; i < b.N; i++ {
				started := time.Now()
				var result []Entity
				query := conn.Model(&Entity{}).
					Where("status = ? AND process_status = ?", 1, 0)
				query = applyAudienceFilter(query, readable, true)
				if err := query.Order("pin_weight DESC, updated_at DESC, id DESC").Limit(20).Find(&result).Error; err != nil {
					b.Fatal(err)
				}
				samples[i] = time.Since(started)
			}
			slices.Sort(samples)
			if len(samples) > 0 {
				p95Index := min((len(samples)*95+99)/100-1, len(samples)-1)
				b.ReportMetric(float64(samples[p95Index].Nanoseconds()), "p95-ns")
			}
		})
	}
}

func seedAudienceBenchmark(b *testing.B, conn *gorm.DB, topicCount int, categoryCount int) {
	b.Helper()
	now := time.Now()
	for start := 1; start <= topicCount; start += 1000 {
		end := min(start+1000, topicCount+1)
		topicsBatch := make([]Entity, 0, end-start)
		indexBatch := make([]topicCategoryIndex.Entity, 0, end-start)
		for id := start; id < end; id++ {
			categoryID := uint64((id-1)%categoryCount + 1)
			topicsBatch = append(topicsBatch, Entity{Id: uint64(id), Title: "benchmark", CategoryIds: []uint64{categoryID}, Status: 1, UpdatedAt: now.Add(time.Duration(id) * time.Microsecond)})
			indexBatch = append(indexBatch, topicCategoryIndex.Entity{TopicId: uint64(id), CategoryId: categoryID, Effective: 1})
		}
		if err := conn.Create(&topicsBatch).Error; err != nil {
			b.Fatalf("seed topics: %v", err)
		}
		if err := conn.Create(&indexBatch).Error; err != nil {
			b.Fatalf("seed category indexes: %v", err)
		}
	}
}
