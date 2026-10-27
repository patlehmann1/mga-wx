package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/patlehmann1/mga-wx/config"
	"github.com/patlehmann1/mga-wx/ui"
)

// Set by goreleaser via -ldflags "-X main.version=...", so it must stay a var.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("mga-wx %s\n", version)
		return
	}

	path, err := config.Path()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %s\n", path, err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(ui.New(cfg), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
