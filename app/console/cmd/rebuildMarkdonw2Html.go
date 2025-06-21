package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "rebuildMarkdonw2Html",
		Short: "",
		Run:   runRebuildMarkdonw2Html,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	// cmd.Flags().StringP("param", "p", "value", "--param=x | -p x")
	appendCommand(cmd)
}

func runRebuildMarkdonw2Html(cmd *cobra.Command, args []string) {
	var articleStartId uint64 = 0
	limit := 100
	for {
		articleList := articles.QueryById(articleStartId, limit)
		for _, article := range articleList {
			if articleStartId < article.Id {
				articleStartId = article.Id
			}
			mdInfo := markdown2html.MarkdownToHTML(article.Content)
			article.RenderedHTML = mdInfo
			article.RenderedVersion = markdown2html.GetVersion()
			articles.SaveNoUpdate(article)
			fmt.Println(article.Id)
		}

		if len(articleList) < limit {
			break
		}
	}
}
