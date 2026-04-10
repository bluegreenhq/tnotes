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

func TestNoteListSelect(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "First"},
		{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "Second"},
	}
	nl := ui.NewNoteList(notes, 30, 20)
	assert.Equal(t, 0, nl.SelectedIndex())

	nl.SelectIndex(1, now)
	assert.Equal(t, 1, nl.SelectedIndex())
}

func TestNoteListSelectedNote(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "First"},
	}
	nl := ui.NewNoteList(notes, 30, 20)
	n, ok := nl.SelectedNote()
	assert.True(t, ok)
	assert.Equal(t, note.NoteID("1"), n.ID)
}

func TestNoteListEmpty(t *testing.T) {
	t.Parallel()

	nl := ui.NewNoteList(nil, 30, 20)
	_, ok := nl.SelectedNote()
	assert.False(t, ok)
}

func TestNoteListMoveDown(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "B"},
	}
	nl := ui.NewNoteList(notes, 30, 20)
	nl.MoveDown(now)
	assert.Equal(t, 1, nl.SelectedIndex())
	nl.MoveDown(now)
	assert.Equal(t, 1, nl.SelectedIndex())
}

func TestNoteListMoveFollowsSectionOrder(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	// notes は更新日時降順: Today(idx=0), Pinned(idx=1), Today(idx=2)
	// 画面上は Pinned(idx=1) → Today(idx=0) → Today(idx=2)
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "today1", UpdatedAt: time.Date(2026, 4, 4, 12, 0, 0, 0, time.Local)}, Body: "Today 1"},
		{Metadata: note.Metadata{ID: "pinned", UpdatedAt: time.Date(2026, 4, 4, 11, 0, 0, 0, time.Local), Pinned: true}, Body: "Pinned"},
		{Metadata: note.Metadata{ID: "today2", UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)}, Body: "Today 2"},
	}
	nl := ui.NewNoteList(notes, 30, 40)

	// 先頭（idx=0=Today1）から上に移動 → 画面上で上はPinned(idx=1)
	nl.SelectIndex(0, fixedNow)
	nl.MoveUp(fixedNow)
	assert.Equal(t, 1, nl.SelectedIndex(), "Today1の上はPinned")

	// Pinned(idx=1)から下に移動 → Today1(idx=0)
	nl.MoveDown(fixedNow)
	assert.Equal(t, 0, nl.SelectedIndex(), "Pinnedの下はToday1")

	// Today1(idx=0)から下に移動 → Today2(idx=2)
	nl.MoveDown(fixedNow)
	assert.Equal(t, 2, nl.SelectedIndex(), "Today1の下はToday2")

	// Today2(idx=2)から上に移動 → Today1(idx=0)
	nl.MoveUp(fixedNow)
	assert.Equal(t, 0, nl.SelectedIndex(), "Today2の上はToday1")
}

func TestNoteListHitTest(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "B"},
	}
	nl := ui.NewNoteList(notes, 30, 20)
	assert.Equal(t, -1, nl.HitTest(5, 2, now)) // section header
	assert.Equal(t, -1, nl.HitTest(5, 3, now)) // section header line
	assert.Equal(t, 0, nl.HitTest(5, 4, now))  // note 0
	assert.Equal(t, 1, nl.HitTest(5, 7, now))  // note 1
	assert.Equal(t, -1, nl.HitTest(5, 0, now))
	assert.Equal(t, -1, nl.HitTest(-1, 4, now))
}

func TestNoteListHitTestWithSections(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", UpdatedAt: time.Date(2026, 4, 3, 12, 0, 0, 0, time.Local)}, Body: "B"},
	}
	nl := ui.NewNoteList(notes, 30, 40)

	assert.Equal(t, -1, nl.HitTest(5, 0, fixedNow)) // header
	assert.Equal(t, -1, nl.HitTest(5, 1, fixedNow)) // separator
	assert.Equal(t, -1, nl.HitTest(5, 2, fixedNow)) // Today label
	assert.Equal(t, -1, nl.HitTest(5, 3, fixedNow)) // Today line
	assert.Equal(t, 0, nl.HitTest(5, 4, fixedNow))  // note 0
	assert.Equal(t, 0, nl.HitTest(5, 6, fixedNow))  // note 0 (last line)
	assert.Equal(t, -1, nl.HitTest(5, 7, fixedNow)) // Yesterday label
	assert.Equal(t, -1, nl.HitTest(5, 8, fixedNow)) // Yesterday line
	assert.Equal(t, 1, nl.HitTest(5, 9, fixedNow))  // note 1
}

func TestNoteListViewWithSections(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)}, Body: "Today note\npreview"},
		{Metadata: note.Metadata{ID: "2", UpdatedAt: time.Date(2026, 4, 3, 12, 0, 0, 0, time.Local)}, Body: "Yesterday note\npreview"},
	}
	nl := ui.NewNoteList(notes, 30, 40)

	view := nl.View(true, false, fixedNow, false)
	assert.Contains(t, view, "Today")
	assert.Contains(t, view, "Yesterday")
	assert.Contains(t, view, "Today note")
	assert.Contains(t, view, "Yesterday note")
	assert.NotContains(t, view, "Previous 7 Days")
	assert.NotContains(t, view, "Previous 30 Days")
}

func TestNoteListUpdateMoveDown(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
		{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "B"},
	}
	nl := ui.NewNoteList(notes, 30, 20)
	nl, cmd := nl.Update(tea.KeyPressMsg{Code: 'j'}, now, false)
	assert.Equal(t, 1, nl.SelectedIndex())
	assert.NotNil(t, cmd)
}

func TestNoteListUpdateMoveUpAtTop(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
	}
	nl := ui.NewNoteList(notes, 30, 20)
	nl, cmd := nl.Update(tea.KeyPressMsg{Code: 'k'}, now, false)
	assert.Equal(t, 0, nl.SelectedIndex())
	assert.Nil(t, cmd)
}

func TestNoteListUpdateCreateMsg(t *testing.T) {
	t.Parallel()

	now := time.Now()
	nl := ui.NewNoteList(nil, 30, 20)
	_, cmd := nl.Update(tea.KeyPressMsg{Code: 'n'}, now, false)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.NoteListCreate, msg)
}

func TestNoteListUpdateTrashMode(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
	}
	nl := ui.NewNoteList(notes, 30, 20)
	_, cmd := nl.Update(tea.KeyPressMsg{Code: 'r'}, now, true)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.NoteListRestore, msg)
}

func TestNoteListTrashModeNoSections(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2026, 4, 4, 15, 0, 0, 0, time.Local)
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.Local)}, Body: "Trashed"},
	}
	nl := ui.NewNoteList(notes, 30, 40)
	nl.SetTitle("Trash")
	nl.SetSectioned(false)

	view := nl.View(true, false, fixedNow, false)
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

func TestNoteListScrollDownDoesNotChangeSelection(t *testing.T) {
	t.Parallel()

	now := time.Now()
	nl := ui.NewNoteList(makeNotes(5, now), 30, 20)
	assert.Equal(t, 0, nl.SelectedIndex())

	nl.ScrollDown(3, now)
	assert.Equal(t, 0, nl.SelectedIndex())
}

func TestNoteListScrollUpDoesNotChangeSelection(t *testing.T) {
	t.Parallel()

	now := time.Now()
	nl := ui.NewNoteList(makeNotes(5, now), 30, 20)
	nl.SelectIndex(4, now)

	nl.ScrollUp(3, now)
	assert.Equal(t, 4, nl.SelectedIndex())
}
