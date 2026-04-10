package cli

import (
	"fmt"
	"io"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

var ErrMissingNoteID = errors.New("missing note ID argument")

const minArgsForGet = 3

func runGet(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForGet {
		_, _ = fmt.Fprintf(w, "Usage: %s get <id>\n", cmdName)

		return ErrMissingNoteID
	}

	id := note.NoteID(args[2])

	n, err := a.GetNote(id)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprint(w, n.Body)

	return nil
}
