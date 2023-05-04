package console

import (
	"github.com/leancodebox/GooseForum/app/console/cmd"
	"github.com/leancodebox/GooseForum/app/migration"
	"github.com/leancodebox/GooseForum/bundles/app"
	"github.com/leancodebox/goose/fileopt"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "GooseForum",
	Short: "A brief description of your application",
	Long:  `GooseForum`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if !fileopt.IsExist("config.toml") {
			err := fileopt.Put([]byte(app.GetDefaulConfig()), "./config.toml")
			if err != nil {
				panic(err)
			}
		}
		fileopt.SetBasePath("storage/")
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
	cobra.CheckErr(rootCmd.Execute())
}
