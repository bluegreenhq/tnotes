package note

import (
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

const (
	maxTitleLen   = 50
	maxPreviewLen = 80
	idByteLen     = 8
)

// NoteID はノートの一意識別子を表す。
type NoteID string

// Metadata はノートのメタデータを表す。サイドバー表示・インデックス用。
type Metadata struct {
	ID        NoteID
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Path      string // データディレクトリからの相対パス
}

// Note はメモ1件を表す。
type Note struct {
	Metadata

	Body string
}

// ZeroNote はゼロ値の Note を返す。
func ZeroNote() Note {
	return Note{Metadata: ZeroMetadata(), Body: ""}
}

// ZeroMetadata はゼロ値の Metadata を返す。
func ZeroMetadata() Metadata {
	return Metadata{ID: "", Title: "", CreatedAt: time.Time{}, UpdatedAt: time.Time{}, Path: ""}
}

// New は新しい空のメモを生成する。
func New(now time.Time) (Note, error) {
	id, err := generateID()
	if err != nil {
		return Note{}, err
	}

	return Note{
		Metadata: Metadata{
			ID:        id,
			Title:     "",
			CreatedAt: now,
			UpdatedAt: now,
			Path:      "",
		},
		Body: "",
	}, nil
}

// FromMetadata は Metadata から Body が空の Note を生成する。
func FromMetadata(m Metadata) Note {
	return Note{Metadata: m, Body: ""}
}

// Title はBodyの1行目を返す。Bodyが空ならMetadata.Titleを返す。それも空なら "New Note"。
func (n Note) Title() string {
	body := strings.TrimSpace(n.Body)
	if body == "" {
		if n.Metadata.Title != "" {
			return n.Metadata.Title
		}

		return "New Note"
	}

	line, _, _ := strings.Cut(body, "\n")

	line = strings.TrimSpace(line)

	if len([]rune(line)) > maxTitleLen {
		return string([]rune(line)[:maxTitleLen]) + "…"
	}

	return line
}

// Preview はBodyの2行目以降をプレビューとして返す。
func (n Note) Preview() string {
	body := strings.TrimSpace(n.Body)

	_, after, found := strings.Cut(body, "\n")
	if !found {
		return ""
	}

	preview := strings.TrimSpace(after)

	preview = strings.ReplaceAll(preview, "\n", " ")

	if len([]rune(preview)) > maxPreviewLen {
		return string([]rune(preview)[:maxPreviewLen]) + "…"
	}

	return preview
}

// SortByUpdatedDesc はメモを更新日時の降順でソートする。
func SortByUpdatedDesc(notes []Note) {
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].UpdatedAt.After(notes[j].UpdatedAt)
	})
}

func generateID() (NoteID, error) {
	b := make([]byte, idByteLen)

	_, err := rand.Read(b)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return NoteID(hex.EncodeToString(b)), nil
}
