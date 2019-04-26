package command

import (
	"github.com/spf13/cobra"

	"github.com/sagikazarmark/gbt/internal/gbt"
)

// AddCommands adds all the commands from cli/command to the root command
func AddCommands(cmd *cobra.Command, config *gbt.Config) {
	cmd.AddCommand(
		NewBuildCommand(config),
		NewRunCommand(config),
		NewCheckCommand(),
		NewTestCommand(),
		NewLintCommand(),
	)
}
