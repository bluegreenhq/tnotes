package ui_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestGroupNotesBySection(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	today := time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)
	yesterday := time.Date(2026, 4, 3, 12, 0, 0, 0, time.Local)
	threeDaysAgo := time.Date(2026, 4, 1, 9, 0, 0, 0, time.Local)
	tenDaysAgo := time.Date(2026, 3, 25, 8, 0, 0, 0, time.Local)
	fiftyDaysAgo := time.Date(2026, 2, 13, 8, 0, 0, 0, time.Local)

	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", UpdatedAt: today}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", UpdatedAt: yesterday}, Body: "B"},
		{Metadata: note.Metadata{ID: "3", UpdatedAt: threeDaysAgo}, Body: "C"},
		{Metadata: note.Metadata{ID: "4", UpdatedAt: tenDaysAgo}, Body: "D"},
		{Metadata: note.Metadata{ID: "5", UpdatedAt: fiftyDaysAgo}, Body: "E"},
	}

	sections := ui.GroupNotesBySection(notes, now)

	assert.Len(t, sections, 4)

	assert.Equal(t, "Today", sections[0].Label)
	assert.Len(t, sections[0].Notes, 1)
	assert.Equal(t, note.NoteID("1"), sections[0].Notes[0].ID)

	assert.Equal(t, "Yesterday", sections[1].Label)
	assert.Len(t, sections[1].Notes, 1)
	assert.Equal(t, note.NoteID("2"), sections[1].Notes[0].ID)

	assert.Equal(t, "Previous 7 Days", sections[2].Label)
	assert.Len(t, sections[2].Notes, 1)
	assert.Equal(t, note.NoteID("3"), sections[2].Notes[0].ID)

	assert.Equal(t, "Previous 30 Days", sections[3].Label)
	assert.Len(t, sections[3].Notes, 2)
	assert.Equal(t, note.NoteID("4"), sections[3].Notes[0].ID)
	assert.Equal(t, note.NoteID("5"), sections[3].Notes[1].ID)
}

func TestGroupNotesBySectionEmpty(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sections := ui.GroupNotesBySection(nil, now)
	assert.Empty(t, sections)
}

func TestGroupNotesBySectionOnlyToday(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", UpdatedAt: time.Date(2026, 4, 4, 8, 0, 0, 0, time.Local)}, Body: "B"},
	}

	sections := ui.GroupNotesBySection(notes, now)
	assert.Len(t, sections, 1)
	assert.Equal(t, "Today", sections[0].Label)
	assert.Len(t, sections[0].Notes, 2)
}
