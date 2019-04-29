package command

import (
	"context"
	"fmt"
	"os"

	lintConfig "github.com/golangci/golangci-lint/pkg/config"
	"github.com/golangci/golangci-lint/pkg/exitcodes"
	"github.com/golangci/golangci-lint/pkg/fsutils"
	"github.com/golangci/golangci-lint/pkg/goutil"
	"github.com/golangci/golangci-lint/pkg/lint"
	"github.com/golangci/golangci-lint/pkg/lint/lintersdb"
	"github.com/golangci/golangci-lint/pkg/logutils"
	"github.com/golangci/golangci-lint/pkg/printers"
	"github.com/golangci/golangci-lint/pkg/report"
	"github.com/golangci/golangci-lint/pkg/result"
	"github.com/golangci/golangci-lint/pkg/result/processors"
	"github.com/spf13/cobra"

	"github.com/sagikazarmark/gbt/internal/gbt"
)

type lintOptions struct {
	packages []string
	verbose  bool
}

// NewLintCommand returns a cobra command for running linters.
func NewLintCommand(config *gbt.Config) *cobra.Command {
	var options lintOptions

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Run linters",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			options.packages = args

			return runLint(options, config)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output")

	return cmd
}

func runLint(options lintOptions, config *gbt.Config) error {
	var reportData report.Data

	logger := report.NewLogWrapper(logutils.NewStderrLog(""), &reportData)

	// cfg := config.NewDefault()
	// setDefaults(cfg)
	c := config.Lint
	cfg := &c
	cfg.Run.IsVerbose = options.verbose

	conf := &c

	r := lintConfig.NewFileReader(conf, cfg, logger.Child("config_reader"))
	if err := r.Read(); err != nil {
		return err
	}

	dbmanager := lintersdb.NewManager(conf)

	conf.LintersSettings.Gocritic.InferEnabledChecks(logger)
	if err := conf.LintersSettings.Gocritic.Validate(logger); err != nil {
		return fmt.Errorf("invalid gocritic settings: %s", err)
	}

	enabledLintersSet := lintersdb.NewEnabledSet(
		dbmanager,
		lintersdb.NewValidator(dbmanager),
		logger.Child("lintersdb"),
		conf,
	)
	goenv := goutil.NewEnv(logger.Child("goenv"))
	contextLoader := lint.NewContextLoader(conf, logger.Child("loader"), goenv)
	fileCache := fsutils.NewFileCache()
	lineCache := fsutils.NewLineCache(fileCache)

	ctx, cancel := context.WithTimeout(context.Background(), conf.Run.Deadline)
	defer cancel()

	if err := goenv.Discover(ctx); err != nil {
		logger.Warnf("failed to discover go env: %s", err)
	}

	conf.Run.Args = options.packages

	enabledLinters, err := enabledLintersSet.Get(true)
	if err != nil {
		return err
	}

	for _, lc := range dbmanager.GetAllSupportedLinterConfigs() {
		isEnabled := false
		for _, enabledLC := range enabledLinters {
			if enabledLC.Name() == lc.Name() {
				isEnabled = true
				break
			}
		}
		reportData.AddLinter(lc.Name(), isEnabled, lc.EnabledByDefault)
	}

	lintCtx, err := contextLoader.Load(ctx, enabledLinters)
	if err != nil {
		return fmt.Errorf("context loading failed: %s", err)
	}
	lintCtx.Log = logger.Child("linters context")

	runner, err := lint.NewRunner(
		lintCtx.ASTCache,
		conf,
		logger.Child("runner"),
		goenv,
		lineCache,
		dbmanager,
	)
	if err != nil {
		return err
	}

	issuesCh := runner.Run(ctx, enabledLinters, lintCtx)
	fixer := processors.NewFixer(conf, logger, fileCache)

	issues := fixer.Process(issuesCh)

	p, err := createPrinter(conf, &reportData, logger)
	if err != nil {
		return err
	}

	var exitCode int
	resCh := make(chan result.Issue, 1024)

	go func() {
		issuesFound := false
		for i := range issues {
			issuesFound = true
			resCh <- i
		}

		if issuesFound {
			exitCode = conf.Run.ExitCodeIfIssuesFound
		}

		close(resCh)
	}()

	if err = p.Print(ctx, resCh); err != nil {
		return fmt.Errorf("can't print %d issues: %s", len(issues), err)
	}

	fileCache.PrintStats(logger)

	if ctx.Err() != nil {
		exitCode = exitcodes.Timeout
		logger.Errorf("Deadline exceeded: try increase it by passing --deadline option")
	}

	if exitCode == exitcodes.Success &&
		(os.Getenv("GL_TEST_RUN") == "1" || os.Getenv("FAIL_ON_WARNINGS") == "1") &&
		len(reportData.Warnings) != 0 {

		exitCode = exitcodes.WarningInTest
	}

	os.Exit(exitCode)

	return nil
}

func createPrinter(conf *lintConfig.Config, reportData *report.Data, logger logutils.Log) (printers.Printer, error) {
	var p printers.Printer
	format := conf.Output.Format
	switch format {
	case lintConfig.OutFormatJSON:
		p = printers.NewJSON(reportData)
	case lintConfig.OutFormatColoredLineNumber, lintConfig.OutFormatLineNumber:
		p = printers.NewText(conf.Output.PrintIssuedLine,
			format == lintConfig.OutFormatColoredLineNumber, conf.Output.PrintLinterName,
			logger.Child("text_printer"))
	case lintConfig.OutFormatTab:
		p = printers.NewTab(conf.Output.PrintLinterName, logger.Child("tab_printer"))
	case lintConfig.OutFormatCheckstyle:
		p = printers.NewCheckstyle()
	case lintConfig.OutFormatCodeClimate:
		p = printers.NewCodeClimate()
	default:
		return nil, fmt.Errorf("unknown output format %s", format)
	}

	return p, nil
}
