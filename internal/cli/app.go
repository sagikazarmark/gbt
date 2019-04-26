package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sagikazarmark/gbt/internal/cli/command"
	"github.com/sagikazarmark/gbt/internal/gbt"
)

// BuildInfo contains version and other build related information generated during the binary compilation.
type BuildInfo struct {
	Version    string
	CommitHash string
	BuildDate  string
}

// NewApplication initializes the CLI application.
func NewApplication(buildInfo BuildInfo) *cobra.Command {
	var configFile string

	v := viper.New()
	v.SetEnvPrefix("gbt")
	v.AllowEmptyEnv(true)
	v.SetConfigName("gbt")
	v.AddConfigPath(".")
	// v.AddConfigPath("$GBT_CONFIG_DIR/")

	config := new(gbt.Config)

	rootCmd := &cobra.Command{
		Use:     "gbt [command]",
		Short:   "Go build tool",
		Version: buildInfo.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if configFile != "" {
				v.SetConfigFile(configFile)
			} else if configFile := os.Getenv("GBTCONFIG"); configFile != "" {
				v.SetConfigFile(configFile)
			}

			if err := v.ReadInConfig(); err != nil {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true

				return err
			}

			if err := v.Unmarshal(config); err != nil {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true

				return err
			}

			return nil
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringVar(&configFile, "config", "", "Custom configuration file")

	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"gbt version %s (%s) built on %s\n",
		buildInfo.Version,
		buildInfo.CommitHash,
		buildInfo.BuildDate,
	))

	command.AddCommands(rootCmd, config)

	return rootCmd
}
