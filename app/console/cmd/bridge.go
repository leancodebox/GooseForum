package cmd

import (
	"github.com/spf13/cobra"
)

var commands = make([]*cobra.Command, 0)

func GetCommands() []*cobra.Command {
	return commands
}
func appendCommand(handle *cobra.Command) {
	commands = append(commands, handle)
}
