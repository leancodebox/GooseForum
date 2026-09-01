package console

import (
	"github.com/leancodebox/GooseForum/app/bundles/closer"
	"github.com/leancodebox/GooseForum/app/bundles/eventbus"
	"github.com/leancodebox/GooseForum/app/console/cmd"
	"github.com/leancodebox/GooseForum/app/service/eventhandlers"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gooseforum",
	Short: "GooseForum command line tools",
	Long:  `GooseForum`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize the event bus for commands that publish domain events.
		eventbus.Start(eventhandlers.Handlers()...)
		return nil
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
	defer closer.CloseAll() // 执行所有注册的关闭逻辑
	cobra.CheckErr(rootCmd.Execute())
}
