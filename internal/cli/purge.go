package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/bluegreenhq/tnotes/internal/app"
)

const minArgsForPurgeFlag = 3

func runPurge(args []string, a *app.App, r io.Reader, w io.Writer) error {
	force := len(args) >= minArgsForPurgeFlag && args[2] == "--force"

	// ゴミ箱の件数を確認
	err := a.EnterTrashMode()
	if err != nil {
		return err
	}

	trashCount := len(a.TrashNotes)

	a.ExitTrashMode()

	if trashCount == 0 {
		_, _ = fmt.Fprintln(w, "Trash is empty")

		return nil
	}

	// 確認プロンプト
	if !force {
		_, _ = fmt.Fprintf(w, "Permanently delete %d note(s)? [y/N] ", trashCount)

		scanner := bufio.NewScanner(r)
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())

		if !strings.EqualFold(answer, "y") {
			_, _ = fmt.Fprintln(w, "Cancelled")

			return nil
		}
	}

	count, err := a.PurgeTrash()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Purged %d note(s)\n", count)

	return nil
}
