package cmd

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/migration"
	"github.com/leancodebox/GooseForum/app/service/topicrankservice"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "rebuild-topic-ranks",
		Short: "Recalculate all topic ranking scores in batches",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := migration.M(); err != nil {
				return err
			}
			count, err := topicrankservice.Rebuild(cmd.Context(), func(n int64) { cmd.Printf("Recalculated %d topics\n", n) })
			if err != nil {
				return fmt.Errorf("rebuild topic ranks after %d topics: %w", count, err)
			}
			cmd.Printf("Topic ranking rebuild completed: %d topics.\n", count)
			return nil
		},
	})
}
