package store //nolint:testpackage // 内部関数 marshalNoteFile/parseNoteFile のテスト

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/note"
)

func TestMarshalFrontmatter(t *testing.T) {
	t.Parallel()

	n := note.Note{
		Metadata: note.Metadata{
			ID:        "abc123",
			CreatedAt: time.Date(2026, 4, 4, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 4, 4, 11, 0, 0, 0, time.UTC),
		},
		Body: "Hello\nWorld",
	}
	got := marshalNoteFile(n)
	assert.Contains(t, got, "---\n")
	assert.Contains(t, got, "id: abc123\n")
	assert.Contains(t, got, "created_at: ")
	assert.Contains(t, got, "updated_at: ")
	assert.Contains(t, got, "Hello\nWorld")
}

func TestUnmarshalFrontmatter(t *testing.T) {
	t.Parallel()

	input := "---\nid: abc123\ncreated_at: 2026-04-04T10:30:00Z\nupdated_at: 2026-04-04T11:00:00Z\n---\nHello\nWorld"
	n, err := parseNoteFile(input)
	require.NoError(t, err)
	assert.Equal(t, note.NoteID("abc123"), n.ID)
	assert.Equal(t, "Hello\nWorld", n.Body)
	assert.Equal(t, 2026, n.CreatedAt.Year())
	assert.Equal(t, 2026, n.UpdatedAt.Year())
}

func TestUnmarshalFrontmatter_EmptyBody(t *testing.T) {
	t.Parallel()

	input := "---\nid: abc123\ncreated_at: 2026-04-04T10:30:00Z\nupdated_at: 2026-04-04T11:00:00Z\n---\n"
	n, err := parseNoteFile(input)
	require.NoError(t, err)
	assert.Equal(t, note.NoteID("abc123"), n.ID)
	assert.Empty(t, n.Body)
}

func TestUnmarshalFrontmatter_InvalidFormat(t *testing.T) {
	t.Parallel()

	input := "no frontmatter here"
	_, err := parseNoteFile(input)
	assert.Error(t, err)
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	original := note.Note{
		Metadata: note.Metadata{
			ID:        "roundtrip1",
			CreatedAt: time.Date(2026, 4, 4, 10, 30, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 4, 4, 11, 0, 0, 0, time.UTC),
		},
		Body: "Line1\nLine2\nLine3",
	}
	data := marshalNoteFile(original)
	parsed, err := parseNoteFile(data)
	require.NoError(t, err)
	assert.Equal(t, original.ID, parsed.ID)
	assert.Equal(t, original.Body, parsed.Body)
	assert.True(t, original.CreatedAt.Equal(parsed.CreatedAt))
	assert.True(t, original.UpdatedAt.Equal(parsed.UpdatedAt))
}
