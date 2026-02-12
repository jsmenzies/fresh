package main

import (
	"flag"
	"fmt"
	"fresh/internal/cli"
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

func main() {
	formatUsageOutput()
	action, cfg, err := cli.ParseFlags(os.Args[1:])

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	switch action {
	case cli.ActionVersion:
		printVersion()
		os.Exit(0)
	case cli.ActionRun:
		runApp(cfg)
		os.Exit(0)
	case cli.ActionHelp:
		flag.Usage()
		os.Exit(0)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func runApp(cfg *cli.Config) {
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
