package store

import (
	"time"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/note"
)

const (
	indexFile  = "index.json"
	timeFormat = "2006-01-02T15:04:05Z07:00"
)

// indexEntry はindex.jsonの各エントリを表す。
type indexEntry struct {
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Path      string `json:"path"`
}

// indexData はindex.jsonの全体構造を表す。
type indexData struct {
	Notes map[string]indexEntry `json:"notes"`
}

// trashIndexEntry はゴミ箱index.jsonの各エントリを表す。
type trashIndexEntry struct {
	Title        string `json:"title"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Path         string `json:"path"`
	OriginalPath string `json:"original_path"`
}

// trashIndexData はゴミ箱index.jsonの全体構造を表す。
type trashIndexData struct {
	Notes map[string]trashIndexEntry `json:"notes"`
}

// trashMetadata はゴミ箱ノートのメタデータ。
type trashMetadata struct {
	note.Metadata

	OriginalPath string
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(timeFormat, s)
	if err != nil {
		return time.Time{}, errors.WithStack(err)
	}

	return t, nil
}
