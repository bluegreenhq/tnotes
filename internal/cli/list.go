package cli

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
)

const (
	tabPadding       = 2
	minArgsForListFlag = 3
)

func runList(args []string, a *app.App, w io.Writer) error {
	trash := len(args) >= minArgsForListFlag && args[2] == "--trash"

	if trash {
		err := a.EnterTrashMode()
		if err != nil {
			return err
		}

		defer a.ExitTrashMode()
	}

	notes := a.Notes
	if trash {
		notes = a.TrashNotes
	}

	if len(notes) == 0 {
		_, _ = fmt.Fprintln(w, "No notes")

		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, tabPadding, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tTITLE\tUPDATED")

	for _, n := range notes {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", n.ID, n.Title(), n.UpdatedAt.Format("2006-01-02 15:04"))
	}

	err := tw.Flush()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
