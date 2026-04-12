package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

var ErrMissingQuery = errors.New("missing search query")

const (
	minArgsForSearch   = 3
	contextFlag        = "--context"
	defaultContextSize = 40
)

type searchResult struct {
	note     note.Note
	snippets []string
}

func runSearch(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForSearch {
		_, _ = fmt.Fprintf(w, "Usage: %s search <query> [--folder <name>] [--context <n>] [--json]\n", cmdName)

		return ErrMissingQuery
	}

	opts := parseSearchFlags(args)
	jsonOut := hasJSONFlag(args)

	notes, err := fetchSearchTarget(a, opts.folder)
	if err != nil {
		return err
	}

	results := searchNotes(a, notes, opts.query, opts.context)

	if jsonOut {
		return writeSearchJSON(w, results)
	}

	return writeSearchText(w, results)
}

type searchOpts struct {
	query   string
	folder  string
	context int
}

func parseSearchFlags(args []string) searchOpts {
	opts := searchOpts{query: "", folder: "", context: defaultContextSize}

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case folderFlag:
			if i+1 < len(args) {
				opts.folder = args[i+1]
				i++
			}
		case contextFlag:
			i = parseContextFlag(args, i, &opts)
		case jsonFlag:
			// skip
		default:
			if opts.query == "" {
				opts.query = args[i]
			}
		}
	}

	return opts
}

func parseContextFlag(args []string, i int, opts *searchOpts) int {
	if i+1 >= len(args) {
		return i
	}

	n, err := strconv.Atoi(args[i+1])
	if err == nil && n > 0 {
		opts.context = n
	}

	return i + 1
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

func searchNotes(a *app.App, notes []note.Note, query string, contextSize int) []searchResult {
	q := strings.ToLower(query)
	results := make([]searchResult, 0)

	for _, n := range notes {
		loaded, err := a.LoadNote(n)
		if err != nil {
			continue
		}

		snippets := extractSnippets(loaded.Body, q, contextSize)
		if len(snippets) > 0 {
			results = append(results, searchResult{note: loaded, snippets: snippets})
		}
	}

	return results
}

func extractSnippets(body string, lowerQuery string, contextSize int) []string {
	lower := strings.ToLower(body)
	runes := []rune(body)
	lowerRunes := []rune(lower)
	queryRunes := []rune(lowerQuery)
	queryLen := len(queryRunes)

	var snippets []string

	for i := 0; i <= len(lowerRunes)-queryLen; i++ {
		if string(lowerRunes[i:i+queryLen]) != lowerQuery {
			continue
		}

		start := max(i-contextSize, 0)
		end := min(i+queryLen+contextSize, len(runes))

		var sb strings.Builder

		if start > 0 {
			sb.WriteString("...")
		}

		sb.WriteString(string(runes[start:end]))

		if end < len(runes) {
			sb.WriteString("...")
		}

		snippets = append(snippets, collapseWhitespace(sb.String()))

		// マッチ位置の直後へスキップして重複を避ける
		i += queryLen - 1
	}

	return snippets
}

func collapseWhitespace(s string) string {
	var sb strings.Builder

	prevSpace := false

	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			if !prevSpace {
				sb.WriteByte(' ')
			}

			prevSpace = true

			continue
		}

		sb.WriteRune(r)

		prevSpace = false
	}

	return sb.String()
}

type searchResultJSON struct {
	noteJSON

	Snippets []string `json:"snippets"`
}

func writeSearchJSON(w io.Writer, results []searchResult) error {
	items := make([]searchResultJSON, 0, len(results))
	for _, r := range results {
		items = append(items, searchResultJSON{
			noteJSON: toNoteJSON(r.note),
			Snippets: r.snippets,
		})
	}

	return writeJSON(w, items)
}

func writeSearchText(w io.Writer, results []searchResult) error {
	if len(results) == 0 {
		_, _ = fmt.Fprintln(w, "No matches")

		return nil
	}

	tw := titleWidth()
	printRow(w, "ID", "TITLE", "UPDATED", tw)

	for _, r := range results {
		title := runewidth.Truncate(r.note.Title(), tw, "...")
		updated := r.note.UpdatedAt.Format("2006-01-02 15:04")
		printRow(w, string(r.note.ID), title, updated, tw)

		for _, s := range r.snippets {
			_, _ = fmt.Fprintf(w, "    %s\n", s)
		}
	}

	return nil
}
