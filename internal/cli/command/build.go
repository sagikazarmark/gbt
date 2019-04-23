package command

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type buildOptions struct {
	targets []string
	verbose bool
}

// NewBuildCommand returns a cobra command for building one or more targets.
func NewBuildCommand() *cobra.Command {
	var options buildOptions

	cmd := &cobra.Command{
		Use:   "build [target1...targetN]",
		Short: "Build all or specific targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			options.targets = args

			return runBuild(options)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func runBuild(options buildOptions) error {
	targets, err := getTargets(options.targets)
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	version := "none"
	commitHash := "none"

	if isGitRepo() {
		c, err := gitClean(gitRun("show", "--format='%h'", "HEAD", "-q"))
		if err != nil {
			return err
		}

		commitHash = c

		v, err := gitClean(gitRun("describe", "--tags", "--exact-match"))
		if err != nil {
			v, err := gitClean(gitRun("symbolic-ref", "-q", "--short", "HEAD"))
			if err != nil {
				return err
			}

			version = v
		} else {
			version = v
		}
	}

	ldflags := fmt.Sprintf(
		"-ldflags=-X \"main.version=%s\" -X \"main.commitHash=%s\" -X \"main.buildDate=%s\"",
		version,
		commitHash,
		now,
	)

	for _, target := range targets {
		args := []string{"build"}

		if options.verbose {
			args = append(args, "-v")
		}

		args = append(
			args,
			ldflags,
			"-o", fmt.Sprintf("build/%s", target),
			fmt.Sprintf("./cmd/%s", target),
		)

		cmd := exec.Command("go", args...)
		cmd.Env = os.Environ()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Println("Compiling", target)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func getTargets(selectedTargets []string) ([]string, error) {
	files, err := ioutil.ReadDir("./cmd")
	if err != nil {
		return nil, err
	}

	var targets []string // nolint: prealloc
	possibleTargets := map[string]bool{}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		targets = append(targets, file.Name())
		possibleTargets[file.Name()] = true
	}

	if len(selectedTargets) > 0 {
		targets = []string{}

		for _, target := range selectedTargets {
			_, ok := possibleTargets[target]
			if !ok {
				return nil, errors.New("unknown build target: " + target)
			}

			targets = append(targets, target)
		}
	}

	return targets, nil
}

// isGitRepo returns true if current folder is a git repository.
func isGitRepo() bool {
	out, err := gitRun("rev-parse", "--is-inside-work-tree")

	return err == nil && strings.TrimSpace(out) == "true"
}

// gitRun runs a git command and returns its output or errors.
func gitRun(args ...string) (string, error) {
	var extraArgs = []string{
		"-c", "log.showSignature=false",
	}
	args = append(extraArgs, args...)

	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return "", errors.New(string(out))
	}

	return string(out), nil
}

// gitClean cleans the git output.
func gitClean(output string, err error) (string, error) {
	output = strings.Replace(strings.Split(output, "\n")[0], "'", "", -1)

	if err != nil {
		err = errors.New(strings.TrimSuffix(err.Error(), "\n"))
	}

	return output, err
}
