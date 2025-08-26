package cmd

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "repairArticleIndexData",
		Short: "检查和修复数据",
		Run:   repairArticleIndexData,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	appendCommand(cmd)
}

func repairArticleIndexData(cmd *cobra.Command, args []string) {

	fmt.Println("=== 开始构建 Meilisearch 索引 ===")
	result, err := searchservice.BuildMeilisearchIndex()
	if err != nil {
		fmt.Printf("构建索引失败: %v\n", err)
	} else {
		fmt.Printf("索引构建成功: 处理了 %d 篇文章\n", result.ProcessedCount)
	}
	fmt.Println("=== Meilisearch 索引构建完成 ===")

}
