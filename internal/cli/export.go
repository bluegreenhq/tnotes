package cli

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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

	dataDir := a.DataDir()

	err := writeZip(outPath, dataDir)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Exported to %s\n", outPath)

	return nil
}

func writeZip(outPath string, dataDir string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	err = filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".tmp-") {
			return nil
		}

		return addFileToZip(zw, dataDir, path)
	})
	if err != nil {
		_ = zw.Close()
		_ = f.Close()
		_ = os.Remove(outPath)

		return errors.WithStack(err)
	}

	return nil
}

func addFileToZip(zw *zip.Writer, baseDir string, path string) error {
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return errors.WithStack(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}

	zf, err := zw.Create(relPath)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = zf.Write(data)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
