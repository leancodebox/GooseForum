package datamigration

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/posts"
	"gorm.io/gorm"
)

type PostMarkdownResult struct {
	Processed  int
	Failed     int
	LastFailed string
}

func RebuildPostMarkdown() PostMarkdownResult {
	return RebuildPostMarkdownWithDB(dbconnect.Connect())
}

func RebuildPostMarkdownWithDB(conn *gorm.DB) PostMarkdownResult {
	const batchSize = 200
	version := markdown2html.GetPostVersion()
	result := PostMarkdownResult{}
	if !conn.Migrator().HasTable(&posts.Entity{}) {
		return result
	}

	var cursor uint64
	for {
		batch, err := posts.PendingMarkdownBatchWithDB(conn, cursor, version, batchSize)
		if err != nil {
			result.Failed++
			result.LastFailed = err.Error()
			return result
		}
		for i := range batch {
			item := &batch[i]
			cursor = item.Id
			renderedHTML := markdown2html.PostMarkdownToHTMLVersion(item.Content, item.SourceVersion)
			if err := posts.UpdateRenderedMarkdownWithDB(conn, item.Id, renderedHTML, version); err != nil {
				result.Failed++
				result.LastFailed = err.Error()
				continue
			}
			result.Processed++
		}
		if len(batch) < batchSize {
			return result
		}
	}
}
