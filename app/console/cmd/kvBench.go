package cmd

import (
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/service/kvstore"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "kvBench",
		Short: "",
		Run:   runKvBench,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	// cmd.Flags().StringP("param", "p", "value", "--param=x | -p x")
	appendCommand(cmd)
}

func runKvBench(cmd *cobra.Command, args []string) {
	// param, _ := cmd.Flags().GetString("param")
	fmt.Println("success")
	kvstore.Set("dev", "test", time.Second)
}
