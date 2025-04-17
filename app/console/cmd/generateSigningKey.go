package cmd

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/algorithm"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "generate:SigningKey",
		Short: "生成 signingKey",
		Run:   runSigningKey,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	})
}
func runSigningKey(_ *cobra.Command, _ []string) {

	fmt.Println("Generated signingKey:", algorithm.SafeGenerateSigningKey(32))
}
