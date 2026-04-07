package cli

import (
	"fmt"
	"io"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

var (
	ErrMissingNoteID = errors.New("missing note ID argument")
	ErrNoteNotFound  = errors.New("note not found")
)

const minArgsForGet = 3

func runGet(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForGet {
		_, _ = fmt.Fprintf(w, "Usage: %s get <id>\n", cmdName)

		return ErrMissingNoteID
	}

	id := note.NoteID(args[2])
	notes := a.Notes

	for _, n := range notes {
		if n.ID == id {
			loaded, err := a.LoadNote(n)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprint(w, loaded.Body)

			return nil
		}
	}

	return errors.WithDetail(ErrNoteNotFound, string(id))
}
