package main

import (
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

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		printVersion()
		os.Exit(0)
	}

	scanDir, err := parseDir()
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}

	if git.IsGitInstalled() == false {
		fmt.Println("Git is not installed or not found in PATH.")
		os.Exit(1)
	}

	m := ui.New(scanDir)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func parseDir() (string, error) {
	var scanDir string
	var err error
	if len(os.Args) > 1 {
		scanDir = os.Args[1]
	} else {
		scanDir, err = os.Getwd()
	}

	if err != nil {
		return "", err
	}

	if _, err := os.Stat(scanDir); os.IsNotExist(err) {
		return "", err
	}

	return scanDir, nil
}

func printVersion() {
	fmt.Printf("fresh %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  built:  %s\n", date)
	fmt.Printf("  by:     %s\n", builtBy)
}
