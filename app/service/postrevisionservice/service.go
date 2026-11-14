package postrevisionservice

import (
	"errors"

	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"github.com/leancodebox/GooseForum/app/models/forum/topics"
	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.New("post revision resource not found")
	ErrForbidden = errors.New("post revision access forbidden")
)

// Authorize restricts revision access to the post author. Moderation privileges
// do not grant access to another user's edit history.
func Authorize(userID uint64, post posts.Entity, topic topics.Entity) error {
	if post.Id == 0 || topic.Id == 0 {
		return ErrNotFound
	}
	if userID == 0 || userID != post.UserId {
		return ErrForbidden
	}
	return nil
}

func List(userID, postID, cursor, pageSize uint64) (postRevisions.Page, error) {
	post := posts.Get(postID)
	if post.Id == 0 {
		return postRevisions.Page{}, ErrNotFound
	}
	topic := topics.GetSimple(post.TopicId)
	if err := Authorize(userID, post, topic); err != nil {
		return postRevisions.Page{}, err
	}
	page, err := postRevisions.ListPageWithDB(dbconnect.Connect(), postRevisions.PageQuery{
		PostId: postID, Cursor: cursor, PageSize: pageSize,
	})
	if err != nil {
		return postRevisions.Page{}, err
	}
	return page, nil
}

func Get(userID, postID, revisionNo uint64) (postRevisions.Entity, error) {
	post := posts.Get(postID)
	if post.Id == 0 {
		return postRevisions.Entity{}, ErrNotFound
	}
	topic := topics.GetSimple(post.TopicId)
	if err := Authorize(userID, post, topic); err != nil {
		return postRevisions.Entity{}, err
	}
	entity, err := postRevisions.GetWithDB(dbconnect.Connect(), postID, revisionNo)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return postRevisions.Entity{}, ErrNotFound
	}
	return entity, err
}
