package datamigration

import (
	"log/slog"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
)

type ReplySequenceResult struct {
	Articles   int
	Replies    int
	Skipped    int
	Failed     int
	LastFailed uint64
}

func BackfillReplySequence() ReplySequenceResult {
	conn := db.Connect()
	result := ReplySequenceResult{}

	var articleIDs []uint64
	if err := conn.Table("reply").
		Distinct("article_id").
		Where("article_id > 0").
		Order("article_id asc").
		Pluck("article_id", &articleIDs).Error; err != nil {
		result.Failed++
		slog.Error("reply sequence article scan failed", "err", err)
		return result
	}

	for _, articleID := range articleIDs {
		var rows []struct {
			Id      uint64
			ReplyNo uint64
		}
		if err := conn.Table("reply").
			Select("id", "reply_no").
			Where("article_id = ?", articleID).
			Order("id asc").
			Find(&rows).Error; err != nil {
			result.Failed++
			result.LastFailed = articleID
			slog.Error("reply sequence row scan failed", "articleId", articleID, "err", err)
			continue
		}

		var maxReplyNo uint64
		for index, row := range rows {
			replyNo := uint64(index + 1)
			maxReplyNo = replyNo
			if row.ReplyNo == replyNo {
				result.Skipped++
				continue
			}
			if err := conn.Table("reply").
				Where("id = ?", row.Id).
				Update("reply_no", replyNo).Error; err != nil {
				result.Failed++
				result.LastFailed = articleID
				slog.Error("reply sequence backfill failed", "articleId", articleID, "replyId", row.Id, "replyNo", replyNo, "err", err)
				continue
			}
			result.Replies++
		}

		if maxReplyNo > 0 {
			if err := conn.Table("articles").
				Where("id = ? AND reply_seq < ?", articleID, maxReplyNo).
				Update("reply_seq", maxReplyNo).Error; err != nil {
				result.Failed++
				result.LastFailed = articleID
				slog.Error("article reply sequence sync failed", "articleId", articleID, "replySeq", maxReplyNo, "err", err)
				continue
			}
		}
		result.Articles++
	}

	return result
}
