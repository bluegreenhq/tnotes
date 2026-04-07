package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/cli"
	"github.com/bluegreenhq/tnotes/internal/store"
	"github.com/bluegreenhq/tnotes/internal/ui"
)

func main() {
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

	m := ui.InitialModel(a)
	p := tea.NewProgram(m)

	_, err = p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
