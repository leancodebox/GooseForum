package datamigration

import (
	"encoding/json"
	"log/slog"
	"time"

	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/category"
	"github.com/leancodebox/GooseForum/app/models/forum/fileUsage"
	"github.com/leancodebox/GooseForum/app/models/forum/migrationMapping"
	"github.com/leancodebox/GooseForum/app/models/forum/moderationLog"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/reports"
	"github.com/leancodebox/GooseForum/app/models/forum/topicCategoryIndex"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserAction"
	"github.com/leancodebox/GooseForum/app/models/forum/topicUserStat"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TopicPostMigrationScope = "topic_post_model"

const (
	legacyModerationSubjectArticle        = "article"
	legacyModerationSubjectReply          = "reply"
	legacyModerationActionArticleBlocked  = "articleBlocked"
	legacyModerationActionArticleRestored = "articleUnblocked"
	legacyModerationActionReplyBlocked    = "replyBlocked"
	legacyModerationActionReplyRestored   = "replyUnblocked"
	legacyReportTargetArticle             = "article"
	legacyReportTargetReply               = "reply"
	legacyFileUsageTargetArticle          = "article"
	legacyFileUsageTargetReply            = "reply"
)

type TopicPostMigrationResult struct {
	Topics                int
	Posts                 int
	Categories            int
	TopicCategoryIndexes  int
	TopicUserActions      int
	TopicUserStats        int
	Mappings              int
	Notifications         int
	ReportsChecked        int
	ReportsMissing        int
	FileUsages            int
	FileUsagesMissing     int
	ModerationLogs        int
	ModerationLogsMissing int
	Skipped               int
	Failed                int
	LastFailed            string
}

type legacyArticleRow struct {
	Id              uint64
	Title           string
	Content         string
	Description     string
	FirstImageURL   string
	RenderedHTML    string
	RenderedVersion uint32
	CategoryId      string
	UserId          uint64
	Posters         string
	ArticleStatus   int8
	ProcessStatus   int8
	LikeCount       uint64
	ViewCount       uint64
	ReplyCount      uint64
	ReplySeq        uint64
	PinWeight       int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type legacyReplyRow struct {
	Id              uint64
	ArticleId       uint64
	ReplyNo         uint64
	UserId          uint64
	TargetId        uint64
	Content         string
	RenderedHTML    string
	RenderedVersion uint32
	ProcessStatus   int8
	ReplyId         uint64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type legacyReportRow struct {
	Id         uint64
	TargetType string
	TargetId   uint64
	ArticleId  uint64
	TopicId    uint64
}

func BackfillTopicPostModel() TopicPostMigrationResult {
	return BackfillTopicPostModelWithDB(db.Connect())
}

func BackfillModerationLogsTopicPost() TopicPostMigrationResult {
	return BackfillModerationLogsTopicPostWithDB(db.Connect())
}

func BackfillFileUsagesTopicPost() TopicPostMigrationResult {
	return BackfillFileUsagesTopicPostWithDB(db.Connect())
}

func BackfillFileUsagesTopicPostWithDB(conn *gorm.DB) TopicPostMigrationResult {
	result := TopicPostMigrationResult{}
	migrateFileUsages(conn, &result)
	return result
}

func BackfillModerationLogsTopicPostWithDB(conn *gorm.DB) TopicPostMigrationResult {
	result := TopicPostMigrationResult{}
	migrateModerationLogs(conn, &result)
	return result
}

func BackfillTopicPostModelWithDB(conn *gorm.DB) TopicPostMigrationResult {
	result := TopicPostMigrationResult{}
	if err := conn.AutoMigrate(
		&topics.Entity{},
		&posts.Entity{},
		&category.Entity{},
		&topicCategoryIndex.Entity{},
		&topicUserAction.Entity{},
		&topicUserStat.Entity{},
		&migrationMapping.Entity{},
	); err != nil {
		result.Failed++
		result.LastFailed = "auto_migrate"
		slog.Error("topic post model migration schema failed", "err", err)
		return result
	}

	migrateCategories(conn, &result)
	migrateTopicsAndFirstPosts(conn, &result)
	migrateReplyPosts(conn, &result)
	linkReplyParents(conn, &result)
	syncTopicPostPointers(conn, &result)
	migrateTopicCategoryIndexes(conn, &result)
	migrateTopicUserActions(conn, &result)
	migrateTopicUserStats(conn, &result)
	enrichNotifications(conn, &result)
	migrateReports(conn, &result)
	return result
}

func migrateCategories(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("article_category") {
		return
	}
	var rows []struct {
		Id        uint64
		Category  string
		Desc      string
		Icon      string
		Color     string
		Slug      string
		Sort      int
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	if err := conn.Table("article_category").Find(&rows).Error; err != nil {
		failMigration(result, "category_scan", err)
		return
	}
	for _, row := range rows {
		entity := category.Entity{
			Id:        row.Id,
			Name:      row.Category,
			Desc:      row.Desc,
			Icon:      row.Icon,
			Color:     row.Color,
			Slug:      row.Slug,
			Sort:      row.Sort,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
		if err := upsert(conn, &entity, []string{"id"}); err != nil {
			failMigration(result, "category", err)
			continue
		}
		result.Categories++
		saveMapping(conn, result, "category", row.Id, "category", row.Id)
	}
}

func migrateTopicsAndFirstPosts(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("articles") {
		return
	}
	var rows []legacyArticleRow
	if err := conn.Table("articles").Where("deleted_at IS NULL").Order("id asc").Find(&rows).Error; err != nil {
		failMigration(result, "article_scan", err)
		return
	}
	for _, row := range rows {
		topic := topics.Entity{
			Id:            row.Id,
			Title:         row.Title,
			CategoryIds:   parseUint64JSON(row.CategoryId),
			UserId:        row.UserId,
			Status:        row.ArticleStatus,
			ProcessStatus: row.ProcessStatus,
			LikeCount:     row.LikeCount,
			ViewCount:     row.ViewCount,
			ReplyCount:    row.ReplyCount,
			PostSeq:       row.ReplySeq + 1,
			PinWeight:     row.PinWeight,
			Excerpt:       row.Description,
			FirstImageURL: row.FirstImageURL,
			Posters:       parseTopicPosters(row.Posters),
			CreatedAt:     row.CreatedAt,
			UpdatedAt:     row.UpdatedAt,
		}
		if err := upsert(conn, &topic, []string{"id"}); err != nil {
			failMigration(result, "topic", err)
			continue
		}
		result.Topics++
		saveMapping(conn, result, "article", row.Id, "topic", row.Id)

		post := posts.Entity{
			TopicId:         row.Id,
			PostNo:          1,
			UserId:          row.UserId,
			Content:         row.Content,
			RenderedHTML:    row.RenderedHTML,
			RenderedVersion: row.RenderedVersion,
			ProcessStatus:   row.ProcessStatus,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		}
		if err := upsert(conn, &post, []string{"topic_id", "post_no"}); err != nil {
			failMigration(result, "first_post", err)
			continue
		}
		if post.Id == 0 {
			conn.Where("topic_id = ? AND post_no = ?", row.Id, 1).First(&post)
		}
		result.Posts++
		saveMapping(conn, result, "article_first_post", row.Id, "post", post.Id)
	}
}

func migrateReplyPosts(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("reply") {
		return
	}
	var rows []legacyReplyRow
	if err := conn.Table("reply").Where("deleted_at IS NULL").Order("article_id asc, reply_no asc, id asc").Find(&rows).Error; err != nil {
		failMigration(result, "reply_scan", err)
		return
	}
	for _, row := range rows {
		if row.ArticleId == 0 {
			result.Skipped++
			continue
		}
		postNo := row.ReplyNo + 1
		if postNo <= 1 {
			postNo = nextPostNo(conn, row.ArticleId)
		}
		post := posts.Entity{
			TopicId:         row.ArticleId,
			PostNo:          postNo,
			UserId:          row.UserId,
			Content:         row.Content,
			RenderedHTML:    row.RenderedHTML,
			RenderedVersion: row.RenderedVersion,
			ProcessStatus:   row.ProcessStatus,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		}
		if err := upsert(conn, &post, []string{"topic_id", "post_no"}); err != nil {
			failMigration(result, "reply_post", err)
			continue
		}
		if post.Id == 0 {
			conn.Where("topic_id = ? AND post_no = ?", row.ArticleId, postNo).First(&post)
		}
		result.Posts++
		saveMapping(conn, result, "reply", row.Id, "post", post.Id)
	}
}

func linkReplyParents(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("reply") {
		return
	}
	var rows []legacyReplyRow
	if err := conn.Table("reply").Where("reply_id > 0").Find(&rows).Error; err != nil {
		failMigration(result, "reply_parent_scan", err)
		return
	}
	for _, row := range rows {
		child := loadMapping(conn, "reply", row.Id)
		parent := loadMapping(conn, "reply", row.ReplyId)
		if child.TargetId == 0 || parent.TargetId == 0 {
			result.Skipped++
			continue
		}
		if err := conn.Model(&posts.Entity{}).Where("id = ?", child.TargetId).Update("reply_to_post_id", parent.TargetId).Error; err != nil {
			failMigration(result, "reply_parent_link", err)
		}
	}
}

func syncTopicPostPointers(conn *gorm.DB, result *TopicPostMigrationResult) {
	var topicIDs []uint64
	if err := conn.Model(&topics.Entity{}).Pluck("id", &topicIDs).Error; err != nil {
		failMigration(result, "topic_pointer_scan", err)
		return
	}
	for _, topicID := range topicIDs {
		var first posts.Entity
		var last posts.Entity
		var count int64
		conn.Where("topic_id = ?", topicID).Order("post_no asc, id asc").First(&first)
		conn.Where("topic_id = ?", topicID).Order("post_no desc, id desc").First(&last)
		conn.Model(&posts.Entity{}).Where("topic_id = ?", topicID).Count(&count)
		if first.Id == 0 {
			continue
		}
		updates := map[string]any{
			"first_post_id":  first.Id,
			"last_post_id":   last.Id,
			"post_count":     uint64(count),
			"reply_count":    uint64(maxInt64(count-1, 0)),
			"post_seq":       last.PostNo,
			"last_posted_at": last.CreatedAt,
		}
		if err := conn.Model(&topics.Entity{}).Where("id = ?", topicID).Updates(updates).Error; err != nil {
			failMigration(result, "topic_pointer_update", err)
		}
	}
}

func migrateTopicCategoryIndexes(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("article_category_rs") {
		return
	}
	var rows []struct {
		Id                uint64
		ArticleId         uint64
		ArticleCategoryId uint64
		Effective         int
		CreatedAt         time.Time
		UpdatedAt         time.Time
	}
	if err := conn.Table("article_category_rs").Find(&rows).Error; err != nil {
		failMigration(result, "topic_category_index_scan", err)
		return
	}
	for _, row := range rows {
		entity := topicCategoryIndex.Entity{
			TopicId:    row.ArticleId,
			CategoryId: row.ArticleCategoryId,
			Effective:  row.Effective,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		}
		if err := upsert(conn, &entity, []string{"topic_id", "category_id"}); err != nil {
			failMigration(result, "topic_category_index", err)
			continue
		}
		if entity.Id == 0 {
			conn.Where("topic_id = ? AND category_id = ?", row.ArticleId, row.ArticleCategoryId).First(&entity)
		}
		result.TopicCategoryIndexes++
		saveMapping(conn, result, "article_category_rs", row.Id, "topic_category_index", entity.Id)
	}
}

func migrateTopicUserActions(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("article_user_action") {
		return
	}
	var rows []struct {
		Id           uint64
		UserId       uint64
		ArticleId    uint64
		LikedAt      *time.Time
		BookmarkedAt *time.Time
		WatchedAt    *time.Time
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}
	if err := conn.Table("article_user_action").Find(&rows).Error; err != nil {
		failMigration(result, "topic_user_action_scan", err)
		return
	}
	for _, row := range rows {
		entity := topicUserAction.Entity{
			UserId:       row.UserId,
			TopicId:      row.ArticleId,
			LikedAt:      row.LikedAt,
			BookmarkedAt: row.BookmarkedAt,
			WatchedAt:    row.WatchedAt,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
		if err := upsert(conn, &entity, []string{"user_id", "topic_id"}); err != nil {
			failMigration(result, "topic_user_action", err)
			continue
		}
		if entity.Id == 0 {
			conn.Where("user_id = ? AND topic_id = ?", row.UserId, row.ArticleId).First(&entity)
		}
		result.TopicUserActions++
		saveMapping(conn, result, "article_user_action", row.Id, "topic_user_action", entity.Id)
	}
}

func migrateTopicUserStats(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("articles_user_stat") {
		return
	}
	var rows []struct {
		Id          uint64
		ArticleId   uint64
		UserId      uint64
		ReplyCount  uint32
		LastReplyAt time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}
	if err := conn.Table("articles_user_stat").Find(&rows).Error; err != nil {
		failMigration(result, "topic_user_stat_scan", err)
		return
	}
	for _, row := range rows {
		entity := topicUserStat.Entity{
			TopicId:     row.ArticleId,
			UserId:      row.UserId,
			ReplyCount:  row.ReplyCount,
			LastReplyAt: row.LastReplyAt,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
		if err := upsert(conn, &entity, []string{"topic_id", "user_id"}); err != nil {
			failMigration(result, "topic_user_stat", err)
			continue
		}
		if entity.Id == 0 {
			conn.Where("topic_id = ? AND user_id = ?", row.ArticleId, row.UserId).First(&entity)
		}
		result.TopicUserStats++
		saveMapping(conn, result, "articles_user_stat", row.Id, "topic_user_stat", entity.Id)
	}
}

func enrichNotifications(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("event_notification") {
		return
	}
	var rows []struct {
		Id      uint64
		Payload string
	}
	if err := conn.Table("event_notification").Select("id", "payload").Find(&rows).Error; err != nil {
		failMigration(result, "notification_scan", err)
		return
	}
	for _, row := range rows {
		var payload map[string]any
		if row.Payload == "" {
			continue
		}
		if err := json.Unmarshal([]byte(row.Payload), &payload); err != nil {
			failMigration(result, "notification_unmarshal", err)
			continue
		}
		articleID := uint64FromJSONNumber(payload["articleId"])
		commentID := uint64FromJSONNumber(payload["commentId"])
		changed := false
		if articleID > 0 {
			payload["topicId"] = articleID
			delete(payload, "articleId")
			changed = true
		}
		if commentID > 0 {
			mapped := loadMapping(conn, "reply", commentID)
			if mapped.TargetId > 0 {
				payload["postId"] = mapped.TargetId
				delete(payload, "commentId")
				changed = true
			}
		}
		if !changed {
			continue
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			failMigration(result, "notification_remarshal", err)
			continue
		}
		if err := conn.Table("event_notification").Where("id = ?", row.Id).Update("payload", string(payloadBytes)).Error; err != nil {
			failMigration(result, "notification_update", err)
			continue
		}
		result.Notifications++
	}
}

func migrateReports(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("reports") {
		return
	}
	var rows []legacyReportRow
	if err := conn.Table("reports").Find(&rows).Error; err != nil {
		failMigration(result, "report_scan", err)
		return
	}
	for _, row := range rows {
		result.ReportsChecked++
		switch row.TargetType {
		case legacyReportTargetArticle:
			topicID := row.TopicId
			if topicID == 0 {
				topicID = row.ArticleId
			}
			if topicID == 0 {
				topicID = row.TargetId
			}
			if topicID == 0 {
				result.ReportsMissing++
				continue
			}
			if err := conn.Model(&reports.Entity{}).Where("id = ?", row.Id).Updates(map[string]any{
				"target_type": reports.TargetTopic,
				"target_id":   topicID,
				"topic_id":    topicID,
			}).Error; err != nil {
				failMigration(result, "report_topic_update", err)
			}
		case legacyReportTargetReply:
			mapped := loadMapping(conn, "reply", row.TargetId)
			if mapped.TargetId == 0 {
				result.ReportsMissing++
				continue
			}
			topicID := row.TopicId
			if topicID == 0 {
				topicID = row.ArticleId
			}
			if topicID == 0 {
				var post posts.Entity
				conn.First(&post, mapped.TargetId)
				topicID = post.TopicId
			}
			if topicID == 0 {
				result.ReportsMissing++
				continue
			}
			if err := conn.Model(&reports.Entity{}).Where("id = ?", row.Id).Updates(map[string]any{
				"target_type": reports.TargetPost,
				"target_id":   mapped.TargetId,
				"topic_id":    topicID,
			}).Error; err != nil {
				failMigration(result, "report_post_update", err)
			}
		}
	}
}

func migrateModerationLogs(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable("moderation_logs") {
		return
	}
	var rows []moderationLog.Entity
	if err := conn.Find(&rows).Error; err != nil {
		failMigration(result, "moderation_log_scan", err)
		return
	}
	for _, row := range rows {
		changed := false
		subjectType := row.SubjectType
		subjectID := row.SubjectId
		action := row.Action
		params := row.Payload.Params
		if params == nil {
			params = map[string]any{}
		}

		switch row.SubjectType {
		case legacyModerationSubjectArticle:
			subjectType = moderationLog.SubjectTopic
			changed = true
		case legacyModerationSubjectReply:
			mapped := loadMapping(conn, "reply", row.SubjectId)
			if mapped.TargetId == 0 {
				result.ModerationLogsMissing++
				continue
			}
			subjectType = moderationLog.SubjectPost
			subjectID = mapped.TargetId
			params["postId"] = mapped.TargetId
			if post := loadPost(conn, mapped.TargetId); post.Id > 0 {
				params["postNo"] = post.PostNo
				if uint64FromJSONNumber(params["topicId"]) == 0 {
					params["topicId"] = post.TopicId
				}
			}
			delete(params, "replyNo")
			changed = true
		}

		switch row.Action {
		case legacyModerationActionArticleBlocked:
			action = moderationLog.ActionTopicBlocked
			changed = true
		case legacyModerationActionArticleRestored:
			action = moderationLog.ActionTopicUnblocked
			changed = true
		case legacyModerationActionReplyBlocked:
			action = moderationLog.ActionPostBlocked
			changed = true
		case legacyModerationActionReplyRestored:
			action = moderationLog.ActionPostUnblocked
			changed = true
		}

		if migrateModerationLogParams(conn, params, result) {
			changed = true
		}
		if !changed {
			continue
		}
		row.Payload.Params = params
		payloadBytes, err := json.Marshal(row.Payload)
		if err != nil {
			failMigration(result, "moderation_log_remarshal", err)
			continue
		}
		if err := conn.Model(&moderationLog.Entity{}).Where("id = ?", row.Id).Updates(map[string]any{
			"action":       action,
			"subject_type": subjectType,
			"subject_id":   subjectID,
			"payload":      string(payloadBytes),
		}).Error; err != nil {
			failMigration(result, "moderation_log_update", err)
			continue
		}
		result.ModerationLogs++
	}
}

func migrateModerationLogParams(conn *gorm.DB, params map[string]any, result *TopicPostMigrationResult) bool {
	changed := false
	if articleID := uint64FromJSONNumber(params["articleId"]); articleID > 0 {
		params["topicId"] = articleID
		delete(params, "articleId")
		changed = true
	}
	if targetType, ok := params["targetType"].(string); ok {
		switch targetType {
		case legacyReportTargetArticle:
			params["targetType"] = reports.TargetTopic
			changed = true
		case legacyReportTargetReply:
			params["targetType"] = reports.TargetPost
			if targetID := uint64FromJSONNumber(params["targetId"]); targetID > 0 {
				mapped := loadMapping(conn, "reply", targetID)
				if mapped.TargetId > 0 {
					params["targetId"] = mapped.TargetId
					if post := loadPost(conn, mapped.TargetId); post.Id > 0 {
						params["postNo"] = post.PostNo
						if uint64FromJSONNumber(params["topicId"]) == 0 {
							params["topicId"] = post.TopicId
						}
					}
				} else {
					result.ModerationLogsMissing++
				}
			}
			delete(params, "replyNo")
			changed = true
		}
	}
	return changed
}

func migrateFileUsages(conn *gorm.DB, result *TopicPostMigrationResult) {
	if !conn.Migrator().HasTable((&fileUsage.Entity{}).TableName()) {
		return
	}
	var rows []fileUsage.Entity
	if err := conn.
		Where("target_type IN ?", []string{legacyFileUsageTargetArticle, legacyFileUsageTargetReply}).
		Find(&rows).Error; err != nil {
		failMigration(result, "file_usage_scan", err)
		return
	}
	for _, row := range rows {
		targetType := ""
		targetID := uint64(0)
		switch row.TargetType {
		case legacyFileUsageTargetArticle:
			targetType = fileUsage.TargetTopic
			targetID = row.TargetId
		case legacyFileUsageTargetReply:
			mapped := loadMapping(conn, "reply", row.TargetId)
			if mapped.TargetId == 0 {
				result.FileUsagesMissing++
				continue
			}
			targetType = fileUsage.TargetPost
			targetID = mapped.TargetId
		default:
			continue
		}
		if targetID == 0 {
			result.FileUsagesMissing++
			continue
		}
		next := row
		next.Id = 0
		next.TargetType = targetType
		next.TargetId = targetID
		if err := conn.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "target_type"},
				{Name: "target_id"},
				{Name: "usage_type"},
				{Name: "file_name"},
			},
			DoNothing: true,
		}).Create(&next).Error; err != nil {
			failMigration(result, "file_usage_create", err)
			continue
		}
		if err := conn.Delete(&fileUsage.Entity{}, row.Id).Error; err != nil {
			failMigration(result, "file_usage_delete", err)
			continue
		}
		result.FileUsages++
	}
}

func saveMapping(conn *gorm.DB, result *TopicPostMigrationResult, sourceType string, sourceID uint64, targetType string, targetID uint64) {
	if sourceID == 0 || targetID == 0 {
		result.Skipped++
		return
	}
	entity := migrationMapping.Entity{
		Scope:      TopicPostMigrationScope,
		SourceType: sourceType,
		SourceId:   sourceID,
		TargetType: targetType,
		TargetId:   targetID,
	}
	if err := upsert(conn, &entity, []string{"scope", "source_type", "source_id"}); err != nil {
		failMigration(result, "mapping", err)
		return
	}
	result.Mappings++
}

func loadMapping(conn *gorm.DB, sourceType string, sourceID uint64) migrationMapping.Entity {
	var entity migrationMapping.Entity
	conn.Where("scope = ? AND source_type = ? AND source_id = ?", TopicPostMigrationScope, sourceType, sourceID).First(&entity)
	return entity
}

func loadPost(conn *gorm.DB, postID uint64) posts.Entity {
	var entity posts.Entity
	conn.First(&entity, postID)
	return entity
}

func upsert[T any](conn *gorm.DB, entity *T, columns []string) error {
	conflictColumns := make([]clause.Column, 0, len(columns))
	for _, column := range columns {
		conflictColumns = append(conflictColumns, clause.Column{Name: column})
	}
	return conn.Clauses(clause.OnConflict{
		Columns:   conflictColumns,
		UpdateAll: true,
	}).Create(entity).Error
}

func parseUint64JSON(raw string) []uint64 {
	var values []uint64
	if raw == "" {
		return values
	}
	_ = json.Unmarshal([]byte(raw), &values)
	return values
}

func parseTopicPosters(raw string) []topics.Poster {
	var values []topics.Poster
	if raw == "" {
		return values
	}
	_ = json.Unmarshal([]byte(raw), &values)
	return values
}

func uint64FromJSONNumber(value any) uint64 {
	switch v := value.(type) {
	case float64:
		if v > 0 {
			return uint64(v)
		}
	case int:
		if v > 0 {
			return uint64(v)
		}
	case int64:
		if v > 0 {
			return uint64(v)
		}
	case uint64:
		return v
	case json.Number:
		parsed, _ := v.Int64()
		if parsed > 0 {
			return uint64(parsed)
		}
	}
	return 0
}

func nextPostNo(conn *gorm.DB, topicID uint64) uint64 {
	var maxPostNo uint64
	conn.Model(&posts.Entity{}).Where("topic_id = ?", topicID).Select("COALESCE(MAX(post_no), 0)").Scan(&maxPostNo)
	return maxPostNo + 1
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func failMigration(result *TopicPostMigrationResult, step string, err error) {
	result.Failed++
	result.LastFailed = step
	slog.Error("topic post model migration failed", "step", step, "err", err)
}
