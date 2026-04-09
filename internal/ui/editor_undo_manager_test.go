package ui_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestEditorUndoManagerSaveAndUndo(t *testing.T) {
	t.Parallel()

	m := ui.NewEditorUndoManager()
	snap := ui.EditorSnapshot{Text: "hello", CursorLine: 0, CursorCol: 5}
	m.Save(snap)

	assert.True(t, m.CanUndo())

	got := m.PopUndo()
	assert.NotNil(t, got)
	assert.Equal(t, "hello", got.Text)
}

func TestEditorUndoManagerSaveAndRedo(t *testing.T) {
	t.Parallel()

	m := ui.NewEditorUndoManager()
	snap := ui.EditorSnapshot{Text: "hello", CursorLine: 0, CursorCol: 5}
	m.Save(snap)
	m.PopUndo()
	m.PushRedo(ui.EditorSnapshot{Text: "current", CursorLine: 0, CursorCol: 7})

	assert.True(t, m.CanRedo())

	got := m.PopRedo()
	assert.NotNil(t, got)
	assert.Equal(t, "current", got.Text)
}

func TestEditorUndoManagerSaveClearsRedoStack(t *testing.T) {
	t.Parallel()

	m := ui.NewEditorUndoManager()
	m.Save(ui.EditorSnapshot{Text: "a", CursorLine: 0, CursorCol: 1})
	m.PopUndo()
	m.PushRedo(ui.EditorSnapshot{Text: "b", CursorLine: 0, CursorCol: 1})
	assert.True(t, m.CanRedo())

	m.Save(ui.EditorSnapshot{Text: "c", CursorLine: 0, CursorCol: 1})
	assert.False(t, m.CanRedo())
}

func TestEditorUndoManagerClear(t *testing.T) {
	t.Parallel()

	m := ui.NewEditorUndoManager()
	m.Save(ui.EditorSnapshot{Text: "a", CursorLine: 0, CursorCol: 1})
	assert.True(t, m.CanUndo())

	m.Clear()
	assert.False(t, m.CanUndo())
	assert.False(t, m.CanRedo())
}

func TestEditorUndoManagerDebounce(t *testing.T) {
	t.Parallel()

	m := ui.NewEditorUndoManager()
	now := time.Now()

	m.MaybeSave(ui.EditorSnapshot{Text: "a", CursorLine: 0, CursorCol: 1}, now)
	assert.True(t, m.CanUndo())

	// 500ms未満では保存されない
	m.MaybeSave(ui.EditorSnapshot{Text: "ab", CursorLine: 0, CursorCol: 2}, now.Add(100*time.Millisecond))
	got := m.PopUndo()
	assert.Equal(t, "a", got.Text)
	assert.False(t, m.CanUndo())
}

func TestEditorUndoManagerDebounceAfterInterval(t *testing.T) {
	t.Parallel()

	m := ui.NewEditorUndoManager()
	now := time.Now()

	m.MaybeSave(ui.EditorSnapshot{Text: "a", CursorLine: 0, CursorCol: 1}, now)
	m.MaybeSave(ui.EditorSnapshot{Text: "ab", CursorLine: 0, CursorCol: 2}, now.Add(600*time.Millisecond))

	got := m.PopUndo()
	assert.Equal(t, "ab", got.Text)
	assert.True(t, m.CanUndo())
}

func TestEditorUndoManagerPopEmpty(t *testing.T) {
	t.Parallel()

	m := ui.NewEditorUndoManager()
	assert.Nil(t, m.PopUndo())
	assert.Nil(t, m.PopRedo())
}

func TestEditorUndoRestore(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello"}
	ed.LoadNote(n)

	// 変更前の状態をスナップショットとして保存
	ed.SaveSnapshot(now)
	ed.SetValue("Hello World")

	// Undo: "Hello" に戻る
	ed.Undo()
	assert.Equal(t, "Hello", ed.Value())

	// Redo: "Hello World" に戻る
	ed.Redo()
	assert.Equal(t, "Hello World", ed.Value())
}

func TestEditorAutoSnapshotOnTextChange(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: ""}
	ed.LoadNote(n)
	ed.Focus()

	ed2, _ := ed.Update(tea.KeyPressMsg{Code: 'a', Text: "a"}, now)
	assert.True(t, ed2.UndoMgr.CanUndo())
}

func TestEditorAutoSnapshotOnNewline(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello"}
	ed.LoadNote(n)
	ed.Focus()

	ed2, _ := ed.Update(tea.KeyPressMsg{Code: tea.KeyEnter}, now)
	assert.True(t, ed2.UndoMgr.CanUndo())
}

func TestEditorUndoClearedOnNoteSwitch(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n1 := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Note 1"}
	n2 := note.Note{Metadata: note.Metadata{ID: "2", CreatedAt: now, UpdatedAt: now}, Body: "Note 2"}

	ed.LoadNote(n1)
	ed.SaveSnapshot(now)
	ed.SetValue("Modified")

	ed.LoadNote(n2)
	assert.False(t, ed.UndoMgr.CanUndo())
}
