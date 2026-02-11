package config

type Config struct {
	ProtectedBranches []string
}

func DefaultConfig() *Config {
	return &Config{
		ProtectedBranches: []string{
			"main",
			"master",
			"develop",
			"dev",
			"production",
			"staging",
			"release",
		},
	}
}
