package cli

import (
	"encoding/json"
	"io"
	"slices"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/note"
)

type noteJSON struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Folder    string `json:"folder"`
	Pinned    bool   `json:"pinned"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type noteDetailJSON struct {
	noteJSON

	Body string `json:"body"`
}

func toNoteJSON(n note.Note) noteJSON {
	return noteJSON{
		ID:        string(n.ID),
		Title:     n.Title(),
		Folder:    n.Folder(),
		Pinned:    n.Pinned,
		CreatedAt: formatTime(n.CreatedAt),
		UpdatedAt: formatTime(n.UpdatedAt),
	}
}

func toNoteDetailJSON(n note.Note) noteDetailJSON {
	return noteDetailJSON{
		noteJSON: toNoteJSON(n),
		Body:     n.Body,
	}
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	err := enc.Encode(v)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

const (
	jsonFlag   = "--json"
	folderFlag = "--folder"
)

func hasJSONFlag(args []string) bool {
	return slices.Contains(args, jsonFlag)
}
