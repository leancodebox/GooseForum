package datamigration

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/postRevisions"
)

func PrunePostHistory() error {
	db := dbconnect.Connect()
	if !db.Migrator().HasTable(&postRevisions.Entity{}) {
		return nil
	}
	return postRevisions.PruneAllWithDB(db)
}
