package main

import (
	"fmt"
	"fresh/internal/git"
	"fresh/internal/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	scanDir, err := parseDir()
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}

	if git.IsGitInstalled() == false {
		fmt.Println("Git is not installed or not found in PATH.")
		os.Exit(1)
	}

	m := ui.NewModel(scanDir)

	if _, err := tea.NewProgram(m).Run(); err != nil {
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
