package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/connect/meiliconnect"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "meilisearchtaskquery",
		Short: "",
		Run:   runMeilisearchtaskquery,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	// cmd.Flags().StringP("param", "p", "value", "--param=x | -p x")
	appendCommand(cmd)
}

func runMeilisearchtaskquery(cmd *cobra.Command, args []string) {
	// param, _ := cmd.Flags().GetString("param")
	fmt.Println("success")
	client := meiliconnect.GetClient()
	indexName := "articles"
	index := client.Index(indexName)

	for i := 1; i < 395; i++ {
		task, err := index.GetTask(cast.ToInt64(i))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(task.Status, task.Error)
		}
	}

}
