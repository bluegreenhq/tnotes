package cli

import (
	"fmt"
	"io"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

var ErrMissingDestFolder = errors.New("missing destination folder")

const (
	minArgsForMoveID     = 3
	minArgsForMoveFolder = 4
)

func runMove(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForMoveID {
		_, _ = fmt.Fprintln(w, "Usage: tnotes move <id> <folder>")

		return ErrMissingNoteID
	}

	if len(args) < minArgsForMoveFolder {
		_, _ = fmt.Fprintln(w, "Usage: tnotes move <id> <folder>")

		return ErrMissingDestFolder
	}

	id := note.NoteID(args[2])
	folder := args[3]

	err := validateFolder(a, folder)
	if err != nil {
		return err
	}

	err = a.MoveNoteToFolder(id, folder)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Moved to %s\n", folder)

	return nil
}
