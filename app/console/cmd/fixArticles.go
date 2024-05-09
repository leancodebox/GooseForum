package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "fix:articles",
		Short: "runFixArticles",
		Run:   runFixArticles,
	})
}

func runFixArticles(_ *cobra.Command, _ []string) {

	for _, entity := range articles.All() {
		entity.Title = strings.Trim(entity.Title, `'`)
		fmt.Println(entity.Title)
		articles.Save(entity)
	}

}
