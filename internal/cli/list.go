package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

const (
	minArgsForListFlag = 3
	colGap             = 2
	idWidth            = 16
	updatedWidth       = 16 // "2006-01-02 15:04"
	fixedColumnsWidth  = idWidth + updatedWidth + colGap*2
	defaultTermWidth   = 80
	minTitleWidth      = 10
)

func runList(args []string, a *app.App, w io.Writer) error {
	trash := len(args) >= minArgsForListFlag && args[2] == "--trash"

	var notes []note.Note

	if trash {
		var err error

		notes, err = a.ListTrash()
		if err != nil {
			return err
		}
	} else {
		notes = a.List()
	}

	if len(notes) == 0 {
		_, _ = fmt.Fprintln(w, "No notes")

		return nil
	}

	tw := titleWidth()
	printRow(w, "ID", "TITLE", "UPDATED", tw)

	for _, n := range notes {
		title := runewidth.Truncate(n.Title(), tw, "...")
		updated := n.UpdatedAt.Format("2006-01-02 15:04")
		printRow(w, string(n.ID), title, updated, tw)
	}

	return nil
}

func titleWidth() int {
	w, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		w = defaultTermWidth
	}

	return max(w-fixedColumnsWidth, minTitleWidth)
}

func printRow(w io.Writer, id, title, updated string, tw int) {
	paddedID := runewidth.FillRight(id, idWidth)
	paddedTitle := runewidth.FillRight(title, tw)
	gap := strings.Repeat(" ", colGap)
	_, _ = fmt.Fprintf(w, "%s%s%s%s%s\n", paddedID, gap, paddedTitle, gap, updated)
}
