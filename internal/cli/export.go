package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
)

var (
	ErrFileExists       = errors.New("file already exists")
	ErrMissingExportArg = errors.New("missing output file argument")
)

const minArgsForExport = 3

func runExport(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForExport {
		_, _ = fmt.Fprintf(w, "Usage: %s export <output.zip>\n", cmdName)

		return ErrMissingExportArg
	}

	outPath := args[2]

	_, statErr := os.Stat(outPath)
	if statErr == nil {
		return errors.WithDetail(ErrFileExists, outPath)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	err = a.Export(f)
	if err != nil {
		_ = f.Close()
		_ = os.Remove(outPath)

		return err
	}

	_, _ = fmt.Fprintf(w, "Exported to %s\n", outPath)

	return nil
}
