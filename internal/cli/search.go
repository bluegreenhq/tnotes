package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

var ErrMissingQuery = errors.New("missing search query")

const minArgsForSearch = 3

func runSearch(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForSearch {
		_, _ = fmt.Fprintf(w, "Usage: %s search <query> [--folder <name>] [--json]\n", cmdName)

		return ErrMissingQuery
	}

	query, folderName := parseSearchFlags(args)
	jsonOut := hasJSONFlag(args)

	notes, err := fetchSearchTarget(a, folderName)
	if err != nil {
		return err
	}

	matched := searchNotes(a, notes, query)

	if jsonOut {
		items := make([]noteJSON, 0, len(matched))
		for _, n := range matched {
			items = append(items, toNoteJSON(n))
		}

		return writeJSON(w, items)
	}

	if len(matched) == 0 {
		_, _ = fmt.Fprintln(w, "No matches")

		return nil
	}

	tw := titleWidth()
	printRow(w, "ID", "TITLE", "UPDATED", tw)

	for _, n := range matched {
		title := runewidth.Truncate(n.Title(), tw, "...")
		updated := n.UpdatedAt.Format("2006-01-02 15:04")
		printRow(w, string(n.ID), title, updated, tw)
	}

	return nil
}

func parseSearchFlags(args []string) (string, string) {
	query := ""
	folderName := ""

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case folderFlag:
			if i+1 < len(args) {
				folderName = args[i+1]
				i++
			}
		case "--json":
			// skip
		default:
			if query == "" {
				query = args[i]
			}
		}
	}

	return query, folderName
}

func fetchSearchTarget(a *app.App, folderName string) ([]note.Note, error) {
	if folderName != "" {
		err := validateFolder(a, folderName)
		if err != nil {
			return nil, err
		}

		return a.ListByFolder(folderName), nil
	}

	return a.ListNotes(), nil
}

func searchNotes(a *app.App, notes []note.Note, query string) []note.Note {
	q := strings.ToLower(query)
	matched := make([]note.Note, 0)

	for _, n := range notes {
		loaded, err := a.LoadNote(n)
		if err != nil {
			continue
		}

		if matchesQuery(loaded, q) {
			matched = append(matched, loaded)
		}
	}

	return matched
}

func matchesQuery(n note.Note, lowerQuery string) bool {
	if strings.Contains(strings.ToLower(n.Title()), lowerQuery) {
		return true
	}

	return strings.Contains(strings.ToLower(n.Body), lowerQuery)
}
