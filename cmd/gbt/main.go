package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sagikazarmark/gbt/internal/cli"
)

// nolint: gochecknoinits
func init() {
	cobra.EnableCommandSorting = false
}

func main() {
	buildInfo := cli.BuildInfo{Version: version, CommitHash: commitHash, BuildDate: buildDate}
	app := cli.NewApplication(buildInfo)

	if err := app.Execute(); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}
