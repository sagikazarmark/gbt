package gbt

import (
	lint "github.com/golangci/golangci-lint/pkg/config"
)

type Config struct {
	Build BuildConfig
	Lint  lint.Config
}

func NewConfig() *Config {
	lintConfig := lint.NewDefault()

	return &Config{
		Lint: *lintConfig,
	}
}

type BuildConfig struct {
	Targets []BuildTargetConfig
}

type BuildTargetConfig struct {
	Name string
	Main string
}
