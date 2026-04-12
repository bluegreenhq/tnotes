package app_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/app"
)

func TestNoteUndoManagerPushAndUndo(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	m := &a.NoteUndo

	action := &app.CreateAction{NoteID: "test"}
	m.Push(action)

	assert.True(t, m.CanUndo())
	assert.False(t, m.CanRedo())

	got := m.PopUndo()
	assert.Equal(t, action, got)
	assert.False(t, m.CanUndo())
}

func TestNoteUndoManagerPushAndRedo(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	m := &a.NoteUndo

	action := &app.CreateAction{NoteID: "test"}
	m.Push(action)
	m.PopUndo()
	m.PushRedo(action)

	assert.True(t, m.CanRedo())

	got := m.PopRedo()
	assert.Equal(t, action, got)
	assert.False(t, m.CanRedo())
}

func TestNoteUndoManagerPushClearsRedoStack(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	m := &a.NoteUndo

	action1 := &app.CreateAction{NoteID: "test1"}
	action2 := &app.CreateAction{NoteID: "test2"}

	m.Push(action1)
	m.PopUndo()
	m.PushRedo(action1)
	assert.True(t, m.CanRedo())

	m.Push(action2)
	assert.False(t, m.CanRedo())
}

func TestNoteUndoManagerPopUndoEmpty(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	assert.Nil(t, a.NoteUndo.PopUndo())
}

func TestNoteUndoManagerPopRedoEmpty(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	assert.Nil(t, a.NoteUndo.PopRedo())
}

func TestAppNoteUndoManager(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	assert.False(t, a.NoteUndo.CanUndo())
}

func TestCreateActionUndo(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	now := time.Now()
	result, _ := a.CreateNote(now, "")

	assert.Len(t, a.ListNotes(), 1)

	action := &app.CreateAction{NoteID: result.Note.ID}
	err := action.Undo(a)
	require.NoError(t, err)
	assert.Empty(t, a.ListNotes())
	assert.Len(t, a.ListTrashNotes(), 1)
}

func TestCreateActionRedo(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	now := time.Now()
	result, _ := a.CreateNote(now, "")

	action := &app.CreateAction{NoteID: result.Note.ID}
	_ = action.Undo(a)
	assert.Empty(t, a.ListNotes())

	err := action.Redo(a)
	require.NoError(t, err)
	assert.Len(t, a.ListNotes(), 1)
}

func TestTrashActionUndo(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	now := time.Now()
	result, _ := a.CreateNote(now, "")

	action := &app.TrashAction{NoteID: result.Note.ID, OriginalIndex: 0, OriginalFolder: app.DefaultFolder}
	_, _ = a.TrashNote(a.ListNotes(), 0)
	assert.Empty(t, a.ListNotes())

	err := action.Undo(a)
	require.NoError(t, err)
	assert.Len(t, a.ListNotes(), 1)
}

func TestTrashActionRedo(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	now := time.Now()
	_, _ = a.CreateNote(now, "")

	action := &app.TrashAction{NoteID: a.ListNotes()[0].ID, OriginalIndex: 0, OriginalFolder: app.DefaultFolder}
	_, _ = a.TrashNote(a.ListNotes(), 0)
	_ = action.Undo(a)
	assert.Len(t, a.ListNotes(), 1)

	err := action.Redo(a)
	require.NoError(t, err)
	assert.Empty(t, a.ListNotes())
}

func TestMoveNoteFromTrash(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.TrashNote(a.ListNotes(), 0)
	assert.Empty(t, a.ListNotes())
	assert.Len(t, a.ListTrashNotes(), 1)

	// MoveNoteToFolder で Trash から Notes に移動
	err := a.MoveNoteToFolder(result.Note.ID, app.DefaultFolder)
	require.NoError(t, err)
	assert.Len(t, a.ListNotes(), 1)
	assert.Empty(t, a.ListTrashNotes())
}

func TestTrashActionUndoRestoresToOriginalFolder(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	now := time.Now()
	result, _ := a.CreateNote(now, "")

	action := &app.TrashAction{NoteID: result.Note.ID, OriginalIndex: 0, OriginalFolder: app.DefaultFolder}
	_, _ = a.TrashNote(a.ListNotes(), 0)
	assert.Empty(t, a.ListNotes())

	// Undo すると元のフォルダに戻る
	err := action.Undo(a)
	require.NoError(t, err)
	assert.Len(t, a.ListNotes(), 1)
	assert.Empty(t, a.ListTrashNotes())
}

func TestAppUndoNote(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	now := time.Now()
	_, _ = a.CreateNote(now, "")
	assert.Len(t, a.ListNotes(), 1)

	_, _ = a.TrashNote(a.ListNotes(), 0)
	assert.Empty(t, a.ListNotes())

	result, err := a.UndoNote()
	require.NoError(t, err)
	assert.Len(t, a.ListNotes(), 1)
	assert.Equal(t, "Redo: Ctrl+Shift+Z", result.InfoHint)
}

func TestAppRedoNote(t *testing.T) {
	t.Parallel()

	a, _ := app.New(nil)
	now := time.Now()
	_, _ = a.CreateNote(now, "")
	_, _ = a.TrashNote(a.ListNotes(), 0)
	_, _ = a.UndoNote()

	result, err := a.RedoNote()
	require.NoError(t, err)
	assert.Empty(t, a.ListNotes())
	assert.Equal(t, "Undo: Ctrl+Z", result.InfoHint)
}
