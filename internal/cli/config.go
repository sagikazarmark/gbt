package cli

import (
	"os"
	"runtime"
	"time"

	lintConfig "github.com/golangci/golangci-lint/pkg/config"
	"github.com/golangci/golangci-lint/pkg/exitcodes"
	"github.com/spf13/viper"
)

func configure(v *viper.Viper) {
	// Lint
	v.SetDefault("lint.output.format", lintConfig.OutFormatColoredLineNumber)
	v.SetDefault("lint.output.print-issued-line", true)
	v.SetDefault("lint.output.print-linter-name", true)
	v.SetDefault("lint.output.color", "auto")
	v.SetDefault("lint.run.issues-exit-code", exitcodes.IssuesFound)
	v.SetDefault("lint.run.deadline", time.Minute)
	v.SetDefault("lint.run.tests", true)
	v.SetDefault("lint.run.concurrency", getDefaultConcurrency())
	v.SetDefault("lint.linters-settings.errcheck.ignore", "fmt:.*")
	v.SetDefault("lint.linters-settings.golint.min-confidence", 0.8)
	v.SetDefault("lint.linters-settings.gofmt.simplify", true)
	v.SetDefault("lint.linters-settings.gocyclo.min-complexity", 30)
	v.SetDefault("lint.linters-settings.dupl.threshold", 150)
	v.SetDefault("lint.linters-settings.goconst.min-len", 3)
	v.SetDefault("lint.linters-settings.goconst.min-occurrences", 3)
	v.SetDefault("lint.linters-settings.lll.tab-width", 1)
	v.SetDefault("lint.issues.exclude-use-default", true)
	v.SetDefault("lint.issues.max-issues-per-linter", 50)
	v.SetDefault("lint.issues.max-same-issues", 3)
}

func getDefaultConcurrency() int {
	if os.Getenv("HELP_RUN") == "1" {
		return 8 // to make stable concurrency for README help generating builds
	}

	return runtime.NumCPU()
}
