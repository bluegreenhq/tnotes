package ui_test

import (
	"strconv"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestSidebarSelect(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "First"},
		{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "Second"},
	}
	sb := ui.NewSidebar(notes, 30, 20)
	assert.Equal(t, 0, sb.SelectedIndex())

	sb.SelectIndex(1, now)
	assert.Equal(t, 1, sb.SelectedIndex())
}

func TestSidebarSelectedNote(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "First"},
	}
	sb := ui.NewSidebar(notes, 30, 20)
	n, ok := sb.SelectedNote()
	assert.True(t, ok)
	assert.Equal(t, note.NoteID("1"), n.ID)
}

func TestSidebarEmpty(t *testing.T) {
	t.Parallel()

	sb := ui.NewSidebar(nil, 30, 20)
	_, ok := sb.SelectedNote()
	assert.False(t, ok)
}

func TestSidebarMoveDown(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "B"},
	}
	sb := ui.NewSidebar(notes, 30, 20)
	sb.MoveDown(now)
	assert.Equal(t, 1, sb.SelectedIndex())
	sb.MoveDown(now)
	assert.Equal(t, 1, sb.SelectedIndex())
}

func TestSidebarHitTest(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "B"},
	}
	sb := ui.NewSidebar(notes, 30, 20)
	assert.Equal(t, -1, sb.HitTest(5, 2, now))
	assert.Equal(t, 0, sb.HitTest(5, 3, now))
	assert.Equal(t, 1, sb.HitTest(5, 7, now))
	assert.Equal(t, -1, sb.HitTest(5, 0, now))
	assert.Equal(t, -1, sb.HitTest(-1, 3, now))
}

func TestSidebarHitTestWithSections(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", UpdatedAt: time.Date(2026, 4, 3, 12, 0, 0, 0, time.Local)}, Body: "B"},
	}
	sb := ui.NewSidebar(notes, 30, 40)

	assert.Equal(t, -1, sb.HitTest(5, 0, fixedNow))
	assert.Equal(t, -1, sb.HitTest(5, 1, fixedNow))
	assert.Equal(t, -1, sb.HitTest(5, 2, fixedNow))
	assert.Equal(t, 0, sb.HitTest(5, 3, fixedNow))
	assert.Equal(t, 0, sb.HitTest(5, 6, fixedNow))
	assert.Equal(t, -1, sb.HitTest(5, 7, fixedNow))
	assert.Equal(t, 1, sb.HitTest(5, 8, fixedNow))
}

func TestSidebarViewWithSections(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)}, Body: "Today note\npreview"},
		{Metadata: note.Metadata{ID: "2", UpdatedAt: time.Date(2026, 4, 3, 12, 0, 0, 0, time.Local)}, Body: "Yesterday note\npreview"},
	}
	sb := ui.NewSidebar(notes, 30, 40)

	view := sb.View(true, false, fixedNow)
	assert.Contains(t, view, "Today")
	assert.Contains(t, view, "Yesterday")
	assert.Contains(t, view, "Today note")
	assert.Contains(t, view, "Yesterday note")
	assert.NotContains(t, view, "Previous 7 Days")
	assert.NotContains(t, view, "Previous 30 Days")
}

func TestSidebarUpdateMoveDown(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "B"},
	}
	sb := ui.NewSidebar(notes, 30, 20)
	sb, cmd := sb.Update(tea.KeyPressMsg{Code: 'j'}, now, false)
	assert.Equal(t, 1, sb.SelectedIndex())
	assert.NotNil(t, cmd)
}

func TestSidebarUpdateMoveUpAtTop(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
	}
	sb := ui.NewSidebar(notes, 30, 20)
	sb, cmd := sb.Update(tea.KeyPressMsg{Code: 'k'}, now, false)
	assert.Equal(t, 0, sb.SelectedIndex())
	assert.Nil(t, cmd)
}

func TestSidebarUpdateCreateMsg(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sb := ui.NewSidebar(nil, 30, 20)
	_, cmd := sb.Update(tea.KeyPressMsg{Code: 'n'}, now, false)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.SidebarCreate, msg)
}

func TestSidebarUpdateTrashMode(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
	}
	sb := ui.NewSidebar(notes, 30, 20)
	_, cmd := sb.Update(tea.KeyPressMsg{Code: 'r'}, now, true)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.SidebarRestore, msg)
}

func TestSidebarTrashModeNoSections(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)}, Body: "Trashed"},
	}
	sb := ui.NewSidebar(notes, 30, 40)
	sb.SetTitle("Trash")
	sb.SetSectioned(false)

	view := sb.View(true, false, fixedNow)
	assert.Contains(t, view, "Trash")
	assert.NotContains(t, view, "Today")
}

func makeNotes(n int, now time.Time) []note.Note {
	notes := make([]note.Note, n)
	for i := range notes {
		id := note.NoteID(strconv.Itoa(i))
		notes[i] = note.Note{
			Metadata: note.Metadata{ID: id, CreatedAt: now, UpdatedAt: now},
			Body:     "Note " + strconv.Itoa(i),
		}
	}

	return notes
}

func TestSidebarScrollDownDoesNotChangeSelection(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sb := ui.NewSidebar(makeNotes(5, now), 30, 20)
	assert.Equal(t, 0, sb.SelectedIndex())

	sb.ScrollDown(3, now)
	assert.Equal(t, 0, sb.SelectedIndex())
}

func TestSidebarScrollUpDoesNotChangeSelection(t *testing.T) {
	t.Parallel()

	now := time.Now()
	sb := ui.NewSidebar(makeNotes(5, now), 30, 20)
	sb.SelectIndex(4, now)

	sb.ScrollUp(3, now)
	assert.Equal(t, 4, sb.SelectedIndex())
}
