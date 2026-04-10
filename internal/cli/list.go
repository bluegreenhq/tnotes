package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

const (
	colGap            = 2
	idWidth           = 16
	updatedWidth      = 16 // "2006-01-02 15:04"
	fixedColumnsWidth = idWidth + updatedWidth + colGap*2
	defaultTermWidth  = 80
	minTitleWidth     = 10
)

func runList(args []string, a *app.App, w io.Writer) error {
	trash, folderName := parseListFlags(args)

	notes, err := fetchNotes(a, trash, folderName)
	if err != nil {
		return err
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

func parseListFlags(args []string) (bool, string) {
	trash := false
	folderName := ""

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--trash":
			trash = true
		case "--folder":
			if i+1 < len(args) {
				folderName = args[i+1]
				i++
			}
		}
	}

	return trash, folderName
}

func fetchNotes(a *app.App, trash bool, folderName string) ([]note.Note, error) {
	if trash {
		return a.ListTrash()
	}

	notes := a.List()

	if folderName != "" {
		err := validateFolder(a, folderName)
		if err != nil {
			return nil, err
		}

		notes = filterNotesByFolder(notes, folderName)
	}

	return notes, nil
}

const pathSplitParts = 2

func filterNotesByFolder(notes []note.Note, folderName string) []note.Note {
	filtered := make([]note.Note, 0, len(notes))
	for _, n := range notes {
		parts := strings.SplitN(n.Path, string(filepath.Separator), pathSplitParts)
		if len(parts) > 0 && parts[0] == folderName {
			filtered = append(filtered, n)
		}
	}

	return filtered
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
