package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
)

var ErrMissingImportArg = errors.New("missing input file argument")

const minArgsForImport = 3

func runImport(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForImport {
		_, _ = fmt.Fprintf(w, "Usage: %s import <input.zip>\n", cmdName)

		return ErrMissingImportArg
	}

	zipPath := args[2]

	if a.HasData() {
		return errors.WithDetail(
			errors.New("data directory is not empty"),
			a.DataDir(),
		)
	}

	f, err := os.Open(zipPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	err = a.Import(f)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Imported from %s\n", zipPath)

	return nil
}
