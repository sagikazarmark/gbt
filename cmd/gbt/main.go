package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sagikazarmark/gbt/internal/cli/command"
)

func init() {
	cobra.EnableCommandSorting = false
}

func main() {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:     "gbt [command]",
		Short:   "Go build tool",
		Version: version,
	}

	rootCmd.SetVersionTemplate(fmt.Sprintf("gbt version %s (%s) built on %s\n", version, commitHash, buildDate))

	command.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}
