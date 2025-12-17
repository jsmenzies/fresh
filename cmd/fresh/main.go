package main

import (
	"flag"
	"fmt"
	"fresh/internal/git"
	"fresh/internal/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
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

func main() {
	formatUsageOutput()
	action, cfg, err := parseCliFlags()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	switch action {
	case ActionVersion:
		printVersion()
		os.Exit(0)
	case ActionRun:
		runApp(cfg)
		os.Exit(0)
	case ActionHelp:
		flag.Usage()
		os.Exit(0)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func parseCliFlags() (Action, *Config, error) {
	var showVersion bool
	var showHelp bool
	var dirPath string
	var noIcons bool

	var defaultDir, _ = os.Getwd()

	flag.BoolVar(&showVersion, "version", false, "Print version information")
	flag.BoolVar(&showVersion, "v", false, "Print version information (shorthand)")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showHelp, "h", false, "Show help message (shorthand)")
	flag.StringVar(&dirPath, "dir", defaultDir, "Specify the directory to scan (shorthand: -d)")
	flag.StringVar(&dirPath, "d", defaultDir, "Specify the directory to scan (shorthand for --dir)")
	flag.BoolVar(&noIcons, "no-icons", false, "Disable icon display")

	flag.Parse()

	if len(flag.Args()) > 0 {
		return ActionRun, nil, fmt.Errorf("unexpected arguments: %v\nUse --dir to specify a directory", flag.Args())
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

func runApp(cfg *Config) {
	if git.IsGitInstalled() == false {
		fmt.Println("Git is not installed or not found in PATH.")
		os.Exit(1)
	}

	m := ui.New(cfg.ScanDir)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func validateScanDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	return nil
}

func formatUsageOutput() {
	flag.Usage = func() {
		fmt.Println("")
		fmt.Println("Usage: fresh [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  --help -h           	Show this help message")
		fmt.Println("  --dir -d <path>     	Specify the directory to scan for git repositories")
		fmt.Println("  --no-icons		Disable icon usage in the UI")
		fmt.Println("  --version -v   	Print version information")
		fmt.Println("\nExample:")
		fmt.Printf("  fresh --dir ~/projects \n\n")
	}
}

func printVersion() {
	fmt.Printf("fresh %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  built:  %s\n", date)
	fmt.Printf("  by:     %s\n", builtBy)
}
