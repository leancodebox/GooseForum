package api

import (
	"log/slog"
	"sync"

	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"github.com/leancodebox/GooseForum/app/models/hotdataserve"
	"github.com/leancodebox/GooseForum/app/service/contentmoderationservice"
	"github.com/leancodebox/GooseForum/app/service/fileusageservice"
	"github.com/leancodebox/GooseForum/app/service/postservice"
	"github.com/leancodebox/GooseForum/app/service/topicservice"
)

type contentReviewJob struct {
	topic       bool
	id, version uint64
}

// A bounded in-memory queue keeps requests independent of word matching without
// creating a goroutine per submission. A full queue applies backpressure.
var contentReviews = make(chan contentReviewJob, 256)
var startContentReviews sync.Once
var contentReviewsIdle sync.WaitGroup

func enqueueContentReview(topic bool, id, version uint64, status string) {
	if status != "pending" {
		return
	}
	startContentReviews.Do(func() {
		// One worker also preserves the ordering of review side effects on SQLite.
		go func() {
			for job := range contentReviews {
				if err := runContentReview(job); err != nil {
					slog.Error("content review failed", "topic", job.topic, "id", job.id, "version", job.version, "error", err)
				}
				contentReviewsIdle.Done()
			}
		}()
	})
	contentReviewsIdle.Add(1)
	contentReviews <- contentReviewJob{topic: topic, id: id, version: version}
}

func runContentReview(job contentReviewJob) error {
	if job.topic {
		topic := topics.Get(job.id)
		if topic.Id == 0 || topic.ModerationVersion != job.version || topic.ModerationStatus != "pending" {
			return nil
		}
		post := posts.Get(topic.FirstPostId)
		if post.Id == 0 {
			return nil
		}
		firstPublication := topic.PublishedAt == nil
		// The pending state is only assigned to an explicit publication request.
		topic.Status = 1
		contentmoderationservice.ReviewTopic(&topic, &post)
		if err := topicservice.SaveTopicAndFirstPost(topicservice.FirstPostWrite{Topic: &topic, FirstPost: &post, CategoryIDs: topic.CategoryIds, ExpectedVersion: &job.version, ReviewOnly: true}); err != nil {
			return err
		}
		fileusageservice.ReplaceTopic(topic.Id, topic.UserId, post.Content)
		hotdataserve.ClearTopicCategoryCache()
		publishTopicReviewResult(&topic, &post, firstPublication)
		return nil
	}
	post := posts.Get(job.id)
	if post.Id == 0 || post.ModerationVersion != job.version || post.ModerationStatus != "pending" {
		return nil
	}
	topic := topics.Get(post.TopicId)
	if topic.Id == 0 {
		return nil
	}
	wasVisible, wasPublished := post.ProcessStatus == 0, post.WasPublished()
	contentmoderationservice.ReviewPost(&post)
	if err := posts.SaveReviewed(&post, job.version); err != nil {
		return err
	}
	if wasVisible != (post.ProcessStatus == 0) {
		postservice.SyncTopicPostStats(topic, post, post.ProcessStatus != 0)
	}
	fileusageservice.ReplacePost(post.Id, post.UserId, post.Content)
	hotdataserve.ClearTopicListCache()
	if !wasPublished && post.ProcessStatus == 0 {
		publishVisiblePost(topic, post)
	}
	return nil
}
