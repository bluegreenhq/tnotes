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

const minArgsForFile = 3

func runCreate(args []string, a *app.App, r io.Reader, w io.Writer) error {
	var body []byte

	var err error

	switch {
	case len(args) >= minArgsForFile:
		// ファイルから読み込み
		body, err = os.ReadFile(args[2])
		if err != nil {
			return errors.WithStack(err)
		}
	default:
		// 標準入力から読み込み
		body, err = io.ReadAll(r)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	if len(body) == 0 {
		return ErrEmptyInput
	}

	now := time.Now()

	result, err := a.CreateNote(now)
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
