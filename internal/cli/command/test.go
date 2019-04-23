package command

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"gotest.tools/gotestsum/testjson"
)

type testOptions struct {
	packages []string
	suites   []string
	// parallel bool
	verbose bool
}

// NewTestCommand returns a cobra command for running tests.
func NewTestCommand() *cobra.Command {
	var options testOptions

	cmd := &cobra.Command{
		Use:   "test [flags] [package1...packageN]",
		Short: "Run tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			options.packages = args

			return runTest(options)
		},
	}

	flags := cmd.Flags()

	flags.StringSliceVarP(&options.suites, "suite", "s", []string{}, "One or more test suites to run")
	// flags.BoolVarP(&options.parallel, "parallel", "p", false, "Run test suites in parallel")
	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func runTest(options testOptions) error {
	packages := options.packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	args := []string{"test", "-json"}

	suites := options.suites
	if len(suites) == 0 {
		suites = []string{""}
	}

	for _, suite := range suites {
		args := args

		if suite != "" {
			args = append(args, "-run", fmt.Sprintf("^Test%s$", strings.Title(suite)))
		}

		args = append(args, packages...)

		cmd := exec.Command("go", args...)
		cmd.Env = os.Environ()

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}

		err = cmd.Start()
		if err != nil {
			return nil
		}

		format := "short"
		if options.verbose {
			format = "short-verbose"
		}

		out := os.Stdout
		handler, err := newEventHandler(format, out, os.Stderr)
		if err != nil {
			return err
		}

		e, err := testjson.ScanTestOutput(testjson.ScanConfig{
			Stdout:  stdout,
			Stderr:  stderr,
			Handler: handler,
		})
		if err != nil {
			return err
		}

		if err := testjson.PrintSummary(out, e, testjson.SummarizeSkipped); err != nil {
			return err
		}

		err = cmd.Wait()
		if err != nil {
			return err
		}
	}

	return nil
}

type eventHandler struct {
	formatter testjson.EventFormatter
	out       io.Writer
	err       io.Writer
}

func (h *eventHandler) Err(text string) error {
	_, err := h.err.Write([]byte(text + "\n"))
	return err
}

func (h *eventHandler) Event(event testjson.TestEvent, execution *testjson.Execution) error {
	line, err := h.formatter(event, execution)
	if err != nil {
		return err
	}

	_, err = h.out.Write([]byte(line))
	return err
}

func (h *eventHandler) Close() error {
	return nil
}

var _ testjson.EventHandler = &eventHandler{}

func newEventHandler(format string, wout io.Writer, werr io.Writer) (*eventHandler, error) {
	formatter := testjson.NewEventFormatter(format)
	if formatter == nil {
		return nil, fmt.Errorf("unknown format %s", format)
	}

	handler := &eventHandler{
		formatter: formatter,
		out:       wout,
		err:       werr,
	}

	return handler, nil
}
