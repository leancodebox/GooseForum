package cmd

import (
	db "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "fixUpdateTime",
		Short: "hexo迁移工具",
		Run:   fixUpdateTime,
	})
}

func fixUpdateTime(_ *cobra.Command, _ []string) {

	for _, entity := range articles.All() {
		entity.UpdatedAt = entity.CreatedAt
		db.Connect().Exec("UPDATE articles SET updated_at = created_at where id = ?", entity.Id)
	}
}
