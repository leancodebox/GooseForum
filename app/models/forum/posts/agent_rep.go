package posts

import "time"

type AgentPost struct {
	Id, TopicId, UserId, PostNo, ReplyToPostId uint64
	Content                                    string
	SourceVersion                              uint8
	CreatedAt, UpdatedAt                       time.Time
}

type AgentReplyTarget struct{ Id, TopicId uint64 }
type AgentSubmission struct {
	Id, TopicId     uint64
	ClientRequestID *string
}

func GetAgentPost(id uint64) (AgentPost, bool, error) {
	var row AgentPost
	result := builder().Model(&Entity{}).Where("id = ? AND process_status = 0", id).Limit(1).Find(&row)
	return row, result.RowsAffected > 0, result.Error
}

func ListAgentPosts(topicID, after uint64, limit int) ([]AgentPost, error) {
	var rows []AgentPost
	err := builder().Model(&Entity{}).Where("topic_id = ? AND process_status = 0 AND post_no > ?", topicID, after).Order("post_no ASC").Limit(limit).Find(&rows).Error
	return rows, err
}

func GetAgentReplyTargets(ids []uint64) ([]AgentReplyTarget, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []AgentReplyTarget
	err := builder().Model(&Entity{}).Where("id IN ? AND process_status = 0", ids).Find(&rows).Error
	return rows, err
}

func AgentPostExists(id, topicID uint64) (bool, error) {
	q := builder().Model(&Entity{}).Where("id = ? AND process_status = 0", id)
	if topicID > 0 {
		q = q.Where("topic_id = ?", topicID)
	}
	var row struct{ Id uint64 }
	result := q.Limit(1).Find(&row)
	return result.RowsAffected > 0, result.Error
}

func GetAgentSubmission(source string, userID, id uint64, requestID string) (Entity, error) {
	q := builder().Unscoped().Where("agent_source = ? AND user_id = ?", source, userID)
	if id > 0 {
		q = q.Where("id = ?", id)
	} else {
		q = q.Where("client_request_id = ?", requestID)
	}
	var row Entity
	err := q.Take(&row).Error
	return row, err
}

func GetTopicSubmission(topicID uint64) (Entity, error) {
	var row Entity
	err := builder().Where("topic_id = ? AND post_no = 1", topicID).Take(&row).Error
	return row, err
}

func ListAgentSubmissions(source string, userID, before uint64, categoryIDs []uint64, globalManage bool, limit int) ([]AgentSubmission, error) {
	q := builder().Unscoped().Model(&Entity{}).Where("agent_source = ? AND user_id = ?", source, userID)
	if before > 0 {
		q = q.Where("posts.id < ?", before)
	}
	// Apply current category permissions before pagination, including deleted submissions.
	if !globalManage {
		q = q.Where("EXISTS (SELECT 1 FROM topics WHERE topics.id = posts.topic_id AND topics.main_category_id IN ?)", categoryIDs)
	}
	var rows []AgentSubmission
	err := q.Order("posts.id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
