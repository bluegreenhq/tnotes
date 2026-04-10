package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
)

var (
	ErrMissingFolderName    = errors.New("missing folder name")
	ErrMissingFolderCommand = errors.New("missing folder subcommand")
	ErrMissingNewFolderName = errors.New("missing new folder name")
)

const (
	minArgsForFolderSub    = 3
	minArgsForFolderName   = 4
	minArgsForFolderFlag   = 5
	minArgsForFolderRename = 5
)

func validateFolder(a *app.App, name string) error {
	exists, err := a.FolderExists(name)
	if err != nil {
		return err
	}

	if !exists {
		return errors.WithDetail(app.ErrFolderNotFound, name)
	}

	return nil
}

func runFolder(args []string, a *app.App, r io.Reader, w io.Writer) error {
	if len(args) < minArgsForFolderSub {
		_, _ = fmt.Fprintln(w, "Usage: tnotes folder <list|create|delete|rename>")

		return ErrMissingFolderCommand
	}

	switch args[2] {
	case "list":
		return runFolderList(a, w)
	case "create":
		return runFolderCreate(args, a, w)
	case "delete":
		return runFolderDelete(args, a, r, w)
	case "rename":
		return runFolderRename(args, a, w)
	default:
		_, _ = fmt.Fprintf(w, "unknown folder command: %s\n", args[2])

		return errors.WithDetail(ErrUnknownCommand, args[2])
	}
}

func runFolderList(a *app.App, w io.Writer) error {
	folders, err := a.ListFolders()
	if err != nil {
		return err
	}

	noteCount := len(a.ListByFolder(app.DefaultFolder))
	_, _ = fmt.Fprintf(w, "%-20s %d\n", app.DefaultFolder, noteCount)

	for _, name := range folders {
		count, err := a.FolderNoteCount(name)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "%-20s %d\n", name, count)
	}

	trashNotes, err := a.ListTrash()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "%-20s %d\n", "Trash", len(trashNotes))

	return nil
}

func runFolderCreate(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForFolderName {
		_, _ = fmt.Fprintln(w, "Usage: tnotes folder create <name>")

		return ErrMissingFolderName
	}

	name := args[3]

	err := a.CreateFolder(name)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Created folder: %s\n", name)

	return nil
}

func runFolderDelete(args []string, a *app.App, r io.Reader, w io.Writer) error {
	if len(args) < minArgsForFolderName {
		_, _ = fmt.Fprintln(w, "Usage: tnotes folder delete <name> [--force]")

		return ErrMissingFolderName
	}

	name := args[3]
	force := len(args) >= minArgsForFolderFlag && args[4] == "--force"

	count, err := a.FolderNoteCount(name)
	if err != nil {
		return err
	}

	if count > 0 && !force {
		_, _ = fmt.Fprintf(w, "Folder %q contains %d note(s). Notes will be moved to Trash.\nDelete? (y/N) ", name, count)

		scanner := bufio.NewScanner(r)
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())

		if !strings.EqualFold(answer, "y") {
			_, _ = fmt.Fprintln(w, "Cancelled")

			return nil
		}
	}

	deleted, err := a.DeleteFolder(name)
	if err != nil {
		return err
	}

	if deleted > 0 {
		_, _ = fmt.Fprintf(w, "Deleted folder: %s (%d note(s) moved to Trash)\n", name, deleted)
	} else {
		_, _ = fmt.Fprintf(w, "Deleted folder: %s\n", name)
	}

	return nil
}

func runFolderRename(args []string, a *app.App, w io.Writer) error {
	if len(args) < minArgsForFolderName {
		_, _ = fmt.Fprintln(w, "Usage: tnotes folder rename <old-name> <new-name>")

		return ErrMissingFolderName
	}

	if len(args) < minArgsForFolderRename {
		_, _ = fmt.Fprintln(w, "Usage: tnotes folder rename <old-name> <new-name>")

		return ErrMissingNewFolderName
	}

	oldName := args[3]
	newName := args[4]

	err := a.RenameFolder(oldName, newName)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Renamed folder: %s → %s\n", oldName, newName)

	return nil
}
