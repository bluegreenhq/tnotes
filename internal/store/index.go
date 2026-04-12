package store

import (
	"time"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/note"
)

const (
	// IndexFile はインデックスファイル名。
	IndexFile  = "index.json"
	timeFormat = "2006-01-02T15:04:05Z07:00"
)

// indexEntry はindex.jsonの各エントリを表す。
type indexEntry struct {
	Title     string `json:"title"`
	Preview   string `json:"preview,omitempty"`
	Pinned    bool   `json:"pinned,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Path      string `json:"path"`
}

// indexData はindex.jsonの全体構造を表す。
type indexData struct {
	Notes map[string]indexEntry `json:"notes"`
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(timeFormat, s)
	if err != nil {
		return time.Time{}, errors.WithStack(err)
	}

	return t, nil
}

func metadataFromEntry(id string, entry indexEntry) note.Metadata {
	nid := note.NoteID(id)
	createdAt, _ := parseTime(entry.CreatedAt)
	updatedAt, _ := parseTime(entry.UpdatedAt)

	return note.Metadata{
		ID:        nid,
		Title:     entry.Title,
		Preview:   entry.Preview,
		Pinned:    entry.Pinned,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Path:      entry.Path,
	}
}
