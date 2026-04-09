package main

import (
	"fmt"
	"os"
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/cli"
	"github.com/bluegreenhq/tnotes/internal/store"
	"github.com/bluegreenhq/tnotes/internal/ui"
)

// version はGoReleaserによりビルド時に -ldflags で注入される。
var version string

func hasFlag(args []string, flag string) bool {
	return slices.Contains(args[1:], flag)
}

func main() {
	cli.Version = version

	dir := os.Getenv("NOTES_DATA_DIR")
	if dir == "" {
		dir = store.DefaultDataDir()
	}

	s, err := store.NewFileStore(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	a, err := app.New(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	handled, err := cli.Run(os.Args, a, os.Stdin, os.Stdout)
	if handled {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		return
	}

	noWrap := hasFlag(os.Args, "--no-wrap")
	m := ui.InitialModel(a, noWrap)
	p := tea.NewProgram(m, tea.WithoutSignalHandler())

	_, err = p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
