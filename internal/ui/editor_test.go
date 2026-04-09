package ui_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestEditorLoadNote(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	assert.Equal(t, note.NoteID("1"), ed.NoteID())
	assert.Equal(t, "Hello\nWorld", ed.Value())
}

func TestEditorEmpty(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	assert.Empty(t, ed.NoteID())
}

func TestEditorDirty(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "original"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	assert.False(t, ed.Dirty())

	ed.SetValue("modified")
	assert.True(t, ed.Dirty())

	ed.MarkClean()
	assert.False(t, ed.Dirty())
}

func TestEditorReadOnly(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "read only content"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SetReadOnly(true)
	assert.True(t, ed.ReadOnly())

	ed2, _ := ed.Update(tea.KeyPressMsg{Code: 'x', Text: "x"}, now)
	assert.Equal(t, "read only content", ed2.Value())

	ed2.SetReadOnly(false)
	assert.False(t, ed2.ReadOnly())
}

func TestEditorSelectionBasic(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld\nFoo"}
	ed.LoadNote(n)

	assert.False(t, ed.HasSelection())

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 1}, ui.SelectionAnchor{Line: 0, Column: 4})
	assert.True(t, ed.HasSelection())

	start, end := ed.NormalizedSelection()
	assert.Equal(t, ui.SelectionAnchor{Line: 0, Column: 1}, start)
	assert.Equal(t, ui.SelectionAnchor{Line: 0, Column: 4}, end)
}

func TestEditorSelectionNormalize(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld"}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 1, Column: 3}, ui.SelectionAnchor{Line: 0, Column: 1})
	start, end := ed.NormalizedSelection()
	assert.Equal(t, ui.SelectionAnchor{Line: 0, Column: 1}, start)
	assert.Equal(t, ui.SelectionAnchor{Line: 1, Column: 3}, end)
}

func TestEditorClearSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello"}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 0}, ui.SelectionAnchor{Line: 0, Column: 3})
	assert.True(t, ed.HasSelection())

	ed.ClearSelection()
	assert.False(t, ed.HasSelection())
}

func TestEditorSelectedText(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld\nFoo"}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 1}, ui.SelectionAnchor{Line: 0, Column: 4})
	assert.Equal(t, "ell", ed.SelectedText())

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 3}, ui.SelectionAnchor{Line: 1, Column: 2})
	assert.Equal(t, "lo\nWo", ed.SelectedText())

	ed.SetSelection(ui.SelectionAnchor{Line: 1, Column: 2}, ui.SelectionAnchor{Line: 0, Column: 3})
	assert.Equal(t, "lo\nWo", ed.SelectedText())
}

func TestEditorSelectedTextNoSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello"}
	ed.LoadNote(n)

	assert.Empty(t, ed.SelectedText())
}

func TestEditorDeleteSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld\nFoo"}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 1}, ui.SelectionAnchor{Line: 0, Column: 4})
	ed.DeleteSelection()
	assert.Equal(t, "Ho\nWorld\nFoo", ed.Value())
	assert.False(t, ed.HasSelection())
}

func TestEditorDeleteSelectionMultiLine(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld\nFoo"}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 3}, ui.SelectionAnchor{Line: 2, Column: 1})
	ed.DeleteSelection()
	assert.Equal(t, "Heloo", ed.Value())
	assert.False(t, ed.HasSelection())
}

func TestEditorCopySelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld"}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 0}, ui.SelectionAnchor{Line: 0, Column: 5})
	err := ed.CopySelection()
	require.NoError(t, err)
	assert.Equal(t, "Hello\nWorld", ed.Value())
	assert.False(t, ed.HasSelection())
}

func TestEditorCutSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld"}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 0}, ui.SelectionAnchor{Line: 0, Column: 5})
	err := ed.CutSelection()
	require.NoError(t, err)
	assert.Equal(t, "\nWorld", ed.Value())
	assert.False(t, ed.HasSelection())
}

func TestEditorShiftArrowSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld"}
	ed.LoadNote(n)
	ed.Focus()

	ed2, _ := ed.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}, now)
	assert.True(t, ed2.HasSelection())
	assert.Equal(t, "H", ed2.SelectedText())

	ed3, _ := ed2.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}, now)
	assert.Equal(t, "He", ed3.SelectedText())
}

func TestEditorShiftArrowThenPlainArrowClearsSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello"}
	ed.LoadNote(n)
	ed.Focus()

	ed2, _ := ed.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}, now)
	assert.True(t, ed2.HasSelection())

	ed3, _ := ed2.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: 0}, now)
	assert.False(t, ed3.HasSelection())
}

func TestEditorViewHasHighlight(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\nWorld"}
	ed.LoadNote(n)
	ed.Focus()

	viewNoSel := ed.View()

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 1}, ui.SelectionAnchor{Line: 0, Column: 4})
	viewWithSel := ed.View()

	assert.NotEqual(t, viewNoSel, viewWithSel)
}

func TestEditorBlinkReset(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	ed.Focus()

	assert.True(t, ed.BlinkVisible())
}

func TestEditorBlinkStopsOnBlur(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	ed.Focus()

	ed.Blur()
	assert.True(t, ed.BlinkVisible())
}

func TestEditorBlinkResetsOnKeyPress(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	ed.Focus()

	ed, _ = ed.Update(tea.KeyPressMsg{Code: 'a', Text: "a"}, now)
	assert.True(t, ed.BlinkVisible())
}
