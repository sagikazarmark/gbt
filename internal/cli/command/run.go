package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
)

type runOptions struct {
	targets []string
	verbose bool
}

// NewRunCommand returns a cobra command for building and running one or more targets.
func NewRunCommand() *cobra.Command {
	var options runOptions

	cmd := &cobra.Command{
		Use:   "run [target1...targetN]",
		Short: "Run all or specific targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			options.targets = args

			return runRun(options)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func runRun(options runOptions) error {
	buildOptions := buildOptions(options)

	err := runBuild(buildOptions)
	if err != nil {
		return err
	}

	targets, err := getTargets(options.targets)
	if err != nil {
		return err
	}

	var group run.Group

	for _, target := range targets {
		target := target
		ctx, cancel := context.WithCancel(context.Background())

		group.Add(
			func() error {
				cmd := exec.CommandContext(ctx, fmt.Sprintf("build/%s", target))
				cmd.Env = os.Environ()
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					return err
				}

				<-ctx.Done()

				return nil
			},
			func(e error) {
				cancel()
			},
		)
	}

	// Setup signal handler
	{
		var (
			cancelInterrupt = make(chan struct{})
			ch              = make(chan os.Signal, 2)
		)
		defer close(ch)

		group.Add(
			func() error {
				signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

				select {
				case <-ch:
				case <-cancelInterrupt:
				}

				return nil
			},
			func(e error) {
				close(cancelInterrupt)
				signal.Stop(ch)
			},
		)
	}

	return group.Run()
}
