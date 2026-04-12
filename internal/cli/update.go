package cli

import (
	"fmt"
	"io"
	"time"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

const minArgsForUpdate = 3

func runUpdate(args []string, a *app.App, r io.Reader, w io.Writer) error {
	if len(args) < minArgsForUpdate {
		_, _ = fmt.Fprintf(w, "Usage: %s update <id> [file]\n", cmdName)

		return ErrMissingNoteID
	}

	id := note.NoteID(args[2])

	// ノートの存在確認
	_, err := a.GetNote(id)
	if err != nil {
		return err
	}

	body, err := readBody(args[3:], r)
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return ErrEmptyInput
	}

	_, err = a.SaveNote(id, string(body), time.Now())
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(w, id)

	return nil
}
