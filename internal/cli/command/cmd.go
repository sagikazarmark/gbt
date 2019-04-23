package command

import (
	"github.com/spf13/cobra"
)

// AddCommands adds all the commands from cli/command to the root command
func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		NewBuildCommand(),
		NewRunCommand(),
		NewCheckCommand(),
		NewTestCommand(),
		NewLintCommand(),
	)
}
