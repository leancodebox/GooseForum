package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "logbench",
		Short: "",
		Run:   runLogbench,
		// Args:  cobra.ExactArgs(1), // 只允许且必须传 1 个参数
	}
	// cmd.Flags().StringP("param", "p", "value", "--param=x | -p x")
	appendCommand(cmd)
}

func runLogbench(cmd *cobra.Command, args []string) {
	// param, _ := cmd.Flags().GetString("param")

	start := time.Now()
	maxI := 30000 * 100
	//maxI = 1
	for range maxI {
		slog.Info("asdasdasdasdasda", "time", time.Now())
	}
	fmt.Println(time.Since(start))

}
