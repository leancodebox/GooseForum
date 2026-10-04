package users

import (
	"fmt"
	"testing"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"gorm.io/gorm"
)

func BenchmarkMentionCandidates(b *testing.B) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&EntityComplete{}); err != nil {
		b.Fatal(err)
	}
	for _, size := range []int{10000, 100000} {
		b.Run(fmt.Sprintf("users_%d", size), func(b *testing.B) {
			prefix := fmt.Sprintf("benchmention%d_", size)
			b.Cleanup(func() {
				db.Unscoped().Where("username_lower >= ? AND username_lower < ?", prefix, prefix+"~").Delete(&EntityComplete{})
			})
			for start := 0; start < size; start += 500 {
				rows := make([]EntityComplete, 500)
				for i := range rows {
					group, status := "active", RestrictionNormal
					if start+i < size/2 {
						group, status = "blocked", RestrictionBanned
					}
					rows[i] = EntityComplete{Username: fmt.Sprintf("%s%s%06d", prefix, group, start+i), RestrictionStatus: status}
				}
				if err := db.Session(&gorm.Session{SkipDefaultTransaction: true}).Create(&rows).Error; err != nil {
					b.Fatal(err)
				}
			}
			for _, tc := range []struct {
				name, query string
				want        int
			}{
				{"common", prefix + "active", 8},
				{"missing", prefix + "missing", 0},
				{"all_filtered", prefix + "blocked", 0},
				{"empty_prefix", "", 8},
			} {
				b.Run(tc.name, func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						rows, err := MentionCandidates(tc.query)
						if err != nil || len(rows) != tc.want {
							b.Fatalf("candidates=%d want=%d error=%v", len(rows), tc.want, err)
						}
					}
				})
			}
		})
	}
}
