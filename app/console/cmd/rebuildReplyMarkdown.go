package cmd

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/service/datamigration"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "rebuildReplyMarkdown",
		Short: "Rebuild rendered HTML for replies",
		Run:   runRebuildReplyMarkdown,
	}
	appendCommand(cmd)
}

func runRebuildReplyMarkdown(cmd *cobra.Command, args []string) {
	result := datamigration.RebuildReplyMarkdown()
	fmt.Printf("reply markdown rebuild done, processed=%d skipped=%d failed=%d\n", result.Processed, result.Skipped, result.Failed)
}
