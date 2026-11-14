package posts

import (
	"log/slog"

	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
	"gorm.io/gorm"
)

// RecordRevisionBestEffort runs only after the content write has committed.
// The old moderation version identifies the replaced source, including when
// concurrent writers finish recording their snapshots out of order.
func RecordRevisionBestEffort(db *gorm.DB, previous, current Entity) {
	db = db.Session(&gorm.Session{SkipDefaultTransaction: true})
	if previous.Content != current.Content {
		err := postRevisions.CreateSnapshotWithDB(db, postRevisions.SnapshotInput{
			PostId: previous.Id, RevisionNo: previous.ModerationVersion + 1,
			AuthorId: previous.UserId, Content: previous.Content,
			SourceVersion: previous.SourceVersion, ModerationVersion: previous.ModerationVersion,
			VisibilitySnapshot: postRevisions.VisibilitySnapshot(previous.ProcessStatus),
		})
		if err != nil {
			slog.Warn("post history write failed", "postId", current.Id, "error", err)
		}
	}
	if err := postRevisions.PruneWithDB(db, current.Id); err != nil {
		slog.Warn("post history cleanup failed", "postId", current.Id, "error", err)
	}
}
