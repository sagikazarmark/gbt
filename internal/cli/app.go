package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sagikazarmark/gbt/internal/cli/command"
)

// BuildInfo contains version and other build related information generated during the binary compilation.
type BuildInfo struct {
	Version    string
	CommitHash string
	BuildDate  string
}

// NewApplication initializes the CLI application.
func NewApplication(buildInfo BuildInfo) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "gbt [command]",
		Short:   "Go build tool",
		Version: buildInfo.Version,
	}

	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"gbt version %s (%s) built on %s\n",
		buildInfo.Version,
		buildInfo.CommitHash,
		buildInfo.BuildDate,
	))

	command.AddCommands(rootCmd)

	return rootCmd
}
