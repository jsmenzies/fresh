package config

import "time"

type TimeoutConfig struct {
	Default time.Duration
	Pull    time.Duration
	Fetch   time.Duration
}

type Config struct {
	ProtectedBranches []string
	Timeout           TimeoutConfig
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
		Timeout: TimeoutConfig{
			Default: 30 * time.Second,
			Pull:    2 * time.Minute,
			Fetch:   1 * time.Minute,
		},
	}
}
