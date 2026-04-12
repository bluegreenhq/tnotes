package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
)

var ErrEmptyInput = errors.New("empty input")

func runCreate(args []string, a *app.App, r io.Reader, w io.Writer) error {
	folder, fileArgs := parseFolderFlag(args[2:])

	if folder != "" {
		err := validateFolder(a, folder)
		if err != nil {
			return err
		}
	}

	body, err := readBody(fileArgs, r)
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return ErrEmptyInput
	}

	now := time.Now()

	result, err := a.CreateNote(now, folder)
	if err != nil {
		return err
	}

	_, err = a.SaveNote(result.Note.ID, string(body), now)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(w, result.Note.ID)

	return nil
}

func parseFolderFlag(fileArgs []string) (string, []string) {
	for i := range fileArgs {
		if fileArgs[i] == folderFlag && i+1 < len(fileArgs) {
			folder := fileArgs[i+1]

			return folder, append(fileArgs[:i], fileArgs[i+2:]...)
		}
	}

	return "", fileArgs
}

func readBody(fileArgs []string, r io.Reader) ([]byte, error) {
	if len(fileArgs) > 0 {
		body, err := os.ReadFile(fileArgs[0])
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return body, nil
	}

	body, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return body, nil
}
