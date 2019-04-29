package command

import (
	"github.com/spf13/cobra"

	"github.com/sagikazarmark/gbt/internal/gbt"
)

type checkOptions struct {
	packages []string
	suites   []string
	// parallel bool
	verbose bool
}

// NewCheckCommand returns a cobra command for running tests and linters.
func NewCheckCommand(config *gbt.Config) *cobra.Command {
	var options checkOptions

	cmd := &cobra.Command{
		Use:   "check [flags] [package1...packageN]",
		Short: "Run tests and linters",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			options.packages = args

			return runCheck(options, config)
		},
	}

	flags := cmd.Flags()

	flags.StringSliceVarP(&options.suites, "test-suite", "s", []string{}, "One or more test suites to run")
	// flags.BoolVarP(&options.parallel, "parallel", "p", false, "Run test suites in parallel")
	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func runCheck(options checkOptions, config *gbt.Config) error {
	testOptions := testOptions(options)

	if err := runTest(testOptions); err != nil {
		return err
	}

	lintOptions := lintOptions{
		packages: options.packages,
		verbose:  options.verbose,
	}

	if err := runLint(lintOptions, config); err != nil {
		return err
	}

	return nil
}
