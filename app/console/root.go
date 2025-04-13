package console

import (
	"github.com/leancodebox/GooseForum/app/bundles/connect/db4fileconnect"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/bundles/logging"
	"github.com/leancodebox/GooseForum/app/console/cmd"
	"github.com/leancodebox/GooseForum/app/migration"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "GooseForum",
	Short: "A brief description of your application",
	Long:  `GooseForum`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		migration.M()
	},
	// Run: runWeb,
}

func init() {
	rootCmd.AddCommand(CmdServe)
	rootCmd.AddCommand(cmd.GetCommands()...)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer PostAll()
	cobra.CheckErr(rootCmd.Execute())
}

func PostAll() {
	defer logging.Shutdown()
	defer db4fileconnect.Close()
	defer dbconnect.Close()
}
