package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/spf13/cobra"

	"github.com/sagikazarmark/gbt/internal/gbt"
)

type runOptions struct {
	targets []string
	verbose bool
}

// NewRunCommand returns a cobra command for building and running one or more targets.
func NewRunCommand(config *gbt.Config) *cobra.Command {
	var options runOptions

	cmd := &cobra.Command{
		Use:   "run [target1...targetN]",
		Short: "Run all or specific targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			options.targets = args

			return runRun(options, config)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func runRun(options runOptions, config *gbt.Config) error {
	buildOptions := buildOptions(options)

	err := runBuild(buildOptions, config)
	if err != nil {
		return err
	}

	targets, err := getTargets(options.targets, config)
	if err != nil {
		return err
	}

	var group run.Group
	var wg sync.WaitGroup

	for _, target := range targets {
		target := target
		ctx, cancel := context.WithCancel(context.Background())
		wg.Add(1)

		cmd := exec.CommandContext(ctx, fmt.Sprintf("build/%s", target))
		cmd.Env = os.Environ()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		group.Add(
			func() error {
				err := cmd.Run()
				wg.Done()
				if err != nil {
					return err
				}

				<-ctx.Done()

				return nil
			},
			func(e error) {
				if s, ok := e.(*signalError); ok {
					_ = cmd.Process.Signal(s.signal)
					for i := 0; i < 5; i++ {
						if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
							break
						}

						time.Sleep(time.Second)
					}
				}

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
				case sig := <-ch:
					return &signalError{sig}
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

	// If every process exists successfully, stop the application
	group.Add(
		func() error {
			wg.Wait()

			return nil
		},
		func(e error) {
		},
	)

	return group.Run()
}

type signalError struct {
	signal os.Signal
}

func (s *signalError) Error() string {
	return "signal received: " + s.signal.String()
}
