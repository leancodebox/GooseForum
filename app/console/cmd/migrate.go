package cmd

import (
	"fmt"

	"github.com/leancodebox/GooseForum/app/migration"
	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Run database schema and data migrations",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := migration.M(); err != nil {
				return fmt.Errorf("run migrations: %w", err)
			}
			fmt.Println("Database migrations completed.")
			return nil
		},
	})
}
