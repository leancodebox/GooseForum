package cmd

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/service/searchservice"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "rebuild-search-index",
		Short: "Rebuild the Meilisearch topic index",
		RunE:  runRebuildSearchIndex,
	}
	appendCommand(cmd)
}

func runRebuildSearchIndex(_ *cobra.Command, _ []string) error {
	fmt.Println("Rebuilding Meilisearch topic index...")
	result, err := searchservice.BuildMeilisearchIndex()
	if err != nil {
		return fmt.Errorf("rebuild Meilisearch topic index: %w", err)
	}
	fmt.Printf("Meilisearch topic index rebuilt: processed %d topics.\n", result.ProcessedCount)
	return nil
}
