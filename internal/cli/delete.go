package cli

import (
	"fmt"
	"io"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

const minArgsForDelete = 3

func runDelete(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForDelete {
		_, _ = fmt.Fprintf(w, "Usage: %s delete <id>\n", cmdName)

		return ErrMissingNoteID
	}

	id := note.NoteID(args[2])

	_, err := a.TrashNote(id)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Deleted %s\n", id)

	return nil
}
