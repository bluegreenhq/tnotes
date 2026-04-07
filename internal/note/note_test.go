package note_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/note"
)

func TestNew(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	n, err := note.New(now)
	require.NoError(t, err)
	assert.NotEmpty(t, string(n.ID))
	assert.Empty(t, n.Body)
	assert.Equal(t, now, n.CreatedAt)
	assert.Equal(t, now, n.UpdatedAt)
}

func TestTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want string
	}{
		{"empty", "", "New Note"},
		{"single line", "Hello World", "Hello World"},
		{"multi line", "First Line\nSecond Line", "First Line"},
		{"long", "ABCDEFGHIJ ABCDEFGHIJ ABCDEFGHIJ ABCDEFGHIJ ABCDEFGHIJ 123456", "ABCDEFGHIJ ABCDEFGHIJ ABCDEFGHIJ ABCDEFGHIJ ABCDEF…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			n, _ := note.New(time.Now())
			n.Body = tt.body
			assert.Equal(t, tt.want, n.Title())
		})
	}
}

func TestPreview(t *testing.T) {
	t.Parallel()

	n, _ := note.New(time.Now())
	n.Body = "Title\nThis is the preview text"
	assert.Equal(t, "This is the preview text", n.Preview())
}

func TestPreviewEmpty(t *testing.T) {
	t.Parallel()

	n, _ := note.New(time.Now())
	n.Body = "Only title"
	assert.Empty(t, n.Preview())
}

func TestSortByUpdatedDesc(t *testing.T) {
	t.Parallel()

	now := time.Now()
	older := note.Note{Metadata: note.Metadata{ID: "a", UpdatedAt: now.Add(-time.Hour)}}
	newer := note.Note{Metadata: note.Metadata{ID: "b", UpdatedAt: now}}

	notes := []note.Note{older, newer}
	note.SortByUpdatedDesc(notes)

	assert.Equal(t, note.NoteID("b"), notes[0].ID)
	assert.Equal(t, note.NoteID("a"), notes[1].ID)
}
