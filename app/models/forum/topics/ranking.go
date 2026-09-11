package topics

import (
	"context"
	"time"

	"github.com/leancodebox/GooseForum/app/bundles/queryopt"
)

// GetForRanking includes hidden/deleted topics so their scores can settle.
// Old decimal scores are not read during conversion to integer ranking.
func GetForRanking(ctx context.Context, id uint64) (Entity, error) {
	var topic Entity
	err := builder().WithContext(ctx).
		Unscoped().
		Select("id", "status", "process_status", "deleted_at", "created_at", "published_at", "like_count", "reply_count", "last_posted_at").
		Where(queryopt.Eq("id", id)).
		First(&topic).Error
	return topic, err
}

// These writes deliberately use a table/map, bypassing read-only entity fields.
func SaveRankScore(ctx context.Context, id uint64, score int64) error {
	return builder().WithContext(ctx).Where(queryopt.Eq("id", id)).UpdateColumn("rank_score", score).Error
}

func SetPublishedAt(ctx context.Context, id uint64, at time.Time) error {
	return builder().WithContext(ctx).Where(queryopt.Eq("id", id)).Where(queryopt.IsNull("published_at")).UpdateColumn("published_at", at).Error
}

func MaxRankingID(ctx context.Context) (uint64, error) {
	var ids []uint64
	err := builder().WithContext(ctx).Order(queryopt.Desc("id")).Limit(1).Pluck("id", &ids).Error
	if err != nil || len(ids) == 0 {
		return 0, err
	}
	return ids[0], nil
}

func RankingIDsAfter(ctx context.Context, cursor, upper uint64, limit int) ([]uint64, error) {
	var ids []uint64
	err := builder().WithContext(ctx).Where(queryopt.Gt("id", cursor)).Where(queryopt.Le("id", upper)).Order(queryopt.Asc("id")).Limit(limit).Pluck("id", &ids).Error
	return ids, err
}
