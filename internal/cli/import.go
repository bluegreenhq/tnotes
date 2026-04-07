package cli

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
)

var (
	ErrDataDirNotEmpty  = errors.New("data directory is not empty")
	ErrMissingImportArg = errors.New("missing input file argument")
	ErrInvalidZipPath   = errors.New("invalid path in zip")
)

const minArgsForImport = 3

func runImport(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForImport {
		_, _ = fmt.Fprintf(w, "Usage: %s import <input.zip>\n", cmdName)

		return ErrMissingImportArg
	}

	zipPath := args[2]
	dataDir := a.DataDir()

	_, statErr := os.Stat(filepath.Join(dataDir, "index.json"))
	if statErr == nil {
		return errors.WithDetail(ErrDataDirNotEmpty, dataDir)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if strings.Contains(f.Name, "..") {
			return errors.WithDetail(ErrInvalidZipPath, f.Name)
		}

		destPath := filepath.Join(dataDir, filepath.Clean(f.Name))

		mkErr := os.MkdirAll(filepath.Dir(destPath), dirPerm)
		if mkErr != nil {
			return errors.WithStack(mkErr)
		}

		writeErr := extractZipFile(f, destPath)
		if writeErr != nil {
			return errors.WithStack(writeErr)
		}
	}

	_, _ = fmt.Fprintf(w, "Imported from %s\n", zipPath)

	return nil
}

const dirPerm = 0o750

func extractZipFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return errors.WithStack(err)
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer out.Close()

	_, err = io.Copy(out, io.LimitReader(rc, int64(f.UncompressedSize64))) //nolint:gosec // zip内ファイルサイズはユーザー管理のバックアップデータ
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
