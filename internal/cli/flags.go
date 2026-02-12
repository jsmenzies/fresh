package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type Config struct {
	ScanDir string
	NoIcons bool
}

type Action int

const (
	ActionRun Action = iota
	ActionVersion
	ActionHelp
)

func ParseFlags(args []string) (Action, *Config, error) {
	var showVersion bool
	var showHelp bool
	var noIcons bool

	defaultDir, _ := os.Getwd()
	dirPath := defaultDir

	fs := flag.NewFlagSet("fresh", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.BoolVar(&showVersion, "version", false, "Print version information")
	fs.BoolVar(&showVersion, "v", false, "Print version information (shorthand)")
	fs.BoolVar(&showHelp, "help", false, "Show help message")
	fs.BoolVar(&showHelp, "h", false, "Show help message (shorthand)")
	fs.StringVar(&dirPath, "dir", defaultDir, "Specify the directory to scan (shorthand: -d)")
	fs.StringVar(&dirPath, "d", defaultDir, "Specify the directory to scan (shorthand for --dir)")
	fs.BoolVar(&noIcons, "no-icons", false, "Disable icon display")

	if err := fs.Parse(args); err != nil {
		return ActionRun, nil, err
	}

	if fs.NArg() > 0 {
		return ActionRun, nil, fmt.Errorf("unexpected arguments: %v\nUse --dir to specify a directory", fs.Args())
	}

	switch {
	case showVersion:
		return ActionVersion, nil, nil
	case showHelp:
		return ActionHelp, nil, nil
	default:
		cfg, err := buildConfig(dirPath, noIcons)
		return ActionRun, cfg, err
	}
}

func buildConfig(dirPath string, noIcons bool) (*Config, error) {
	if err := validateScanDir(dirPath); err != nil {
		return nil, err
	}

	cfg := &Config{
		ScanDir: dirPath,
		NoIcons: noIcons,
	}
	return cfg, nil
}

func validateScanDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", dir)
		}
		return fmt.Errorf("cannot access directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dir)
	}
	return nil
}
