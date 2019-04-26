package gbt

type Config struct {
	Build BuildConfig
}

type BuildConfig struct {
	Targets []BuildTargetConfig
}

type BuildTargetConfig struct {
	Name string
	Main string
}
