package app

import (
	"github.com/bluegreenhq/tnotes/internal/note"
)

// UndoableAction はundo/redo可能なノート操作を表す。
type UndoableAction interface {
	Undo(a *App) error
	Redo(a *App) error
	TargetNoteID() note.NoteID
}

// NoteUndoManager はノート操作のundo/redoスタックを管理する。
type NoteUndoManager struct {
	undoStack []UndoableAction
	redoStack []UndoableAction
}

// Push はアクションをundoスタックに追加し、redoスタックをクリアする。
func (m *NoteUndoManager) Push(action UndoableAction) {
	m.undoStack = append(m.undoStack, action)
	m.redoStack = nil
}

// PushUndo はアクションをundoスタックに追加する（redoスタックはクリアしない）。
func (m *NoteUndoManager) PushUndo(action UndoableAction) {
	m.undoStack = append(m.undoStack, action)
}

// PushRedo はアクションをredoスタックに追加する。
func (m *NoteUndoManager) PushRedo(action UndoableAction) {
	m.redoStack = append(m.redoStack, action)
}

// PopUndo はundoスタックから最新のアクションを取り出す。空ならnilを返す。
func (m *NoteUndoManager) PopUndo() UndoableAction { //nolint:ireturn // スタックに格納された任意のアクションを返す
	if len(m.undoStack) == 0 {
		return nil
	}

	action := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]

	return action
}

// PopRedo はredoスタックから最新のアクションを取り出す。空ならnilを返す。
func (m *NoteUndoManager) PopRedo() UndoableAction { //nolint:ireturn // スタックに格納された任意のアクションを返す
	if len(m.redoStack) == 0 {
		return nil
	}

	action := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]

	return action
}

// CanUndo はundoスタックにアクションがあるかを返す。
func (m *NoteUndoManager) CanUndo() bool { return len(m.undoStack) > 0 }

// CanRedo はredoスタックにアクションがあるかを返す。
func (m *NoteUndoManager) CanRedo() bool { return len(m.redoStack) > 0 }

// UndoNote は直前のノート操作を取り消す。undoスタックが空なら空のNoteResultを返す。
func (a *App) UndoNote() (NoteResult, error) {
	action := a.NoteUndo.PopUndo()
	if action == nil {
		return NoteResult{}, nil //nolint:exhaustruct // スタック空
	}

	err := action.Undo(a)
	if err != nil {
		return NoteResult{}, err
	}

	a.NoteUndo.PushRedo(action)

	selectIdx := a.findNoteIndex(action.TargetNoteID())
	if selectIdx < 0 && len(a.Notes) > 0 {
		selectIdx = 0
	}

	return NoteResult{Notes: a.Notes, SelectIdx: selectIdx, InfoHint: "Redo: Ctrl+Shift+Z"}, nil //nolint:exhaustruct // undo結果にNoteは不要
}

// RedoNote は直前のundo操作をやり直す。redoスタックが空なら空のNoteResultを返す。
func (a *App) RedoNote() (NoteResult, error) {
	action := a.NoteUndo.PopRedo()
	if action == nil {
		return NoteResult{}, nil //nolint:exhaustruct // スタック空
	}

	err := action.Redo(a)
	if err != nil {
		return NoteResult{}, err
	}

	a.NoteUndo.PushUndo(action)

	selectIdx := a.findNoteIndex(action.TargetNoteID())
	if selectIdx < 0 && len(a.Notes) > 0 {
		selectIdx = 0
	}

	return NoteResult{Notes: a.Notes, SelectIdx: selectIdx, InfoHint: "Undo: Ctrl+Z"}, nil //nolint:exhaustruct // redo結果にNoteは不要
}

// CreateAction はノート作成操作を表す。
type CreateAction struct {
	NoteID note.NoteID
}

// Undo はノートをゴミ箱に移動する（作成の取り消し）。
func (c *CreateAction) Undo(a *App) error {
	idx := a.findNoteIndex(c.NoteID)
	if idx < 0 {
		return nil
	}

	_, err := a.trashNoteInternal(idx)

	return err
}

// TargetNoteID は対象ノートのIDを返す。
func (c *CreateAction) TargetNoteID() note.NoteID { return c.NoteID }

// Redo はノートをゴミ箱から復元する（作成のやり直し）。
func (c *CreateAction) Redo(a *App) error {
	idx := a.findTrashNoteIndex(c.NoteID)
	if idx < 0 {
		return nil
	}

	_, _, err := a.restoreNoteInternal(idx)

	return err
}

// TrashAction はノート削除操作を表す。
type TrashAction struct {
	NoteID        note.NoteID
	OriginalIndex int
}

// Undo はノートをゴミ箱から復元する（削除の取り消し）。
func (ta *TrashAction) Undo(a *App) error {
	idx := a.findTrashNoteIndex(ta.NoteID)
	if idx < 0 {
		return nil
	}

	_, _, err := a.restoreNoteInternal(idx)

	return err
}

// TargetNoteID は対象ノートのIDを返す。
func (ta *TrashAction) TargetNoteID() note.NoteID { return ta.NoteID }

// Redo はノートをゴミ箱に移動する（削除のやり直し）。
func (ta *TrashAction) Redo(a *App) error {
	idx := a.findNoteIndex(ta.NoteID)
	if idx < 0 {
		return nil
	}

	_, err := a.trashNoteInternal(idx)

	return err
}

// RestoreAction はノート復元操作を表す。
type RestoreAction struct {
	NoteID note.NoteID
}

// Undo はノートをゴミ箱に戻す（復元の取り消し）。
func (r *RestoreAction) Undo(a *App) error {
	idx := a.findNoteIndex(r.NoteID)
	if idx < 0 {
		return nil
	}

	_, err := a.trashNoteInternal(idx)

	return err
}

// TargetNoteID は対象ノートのIDを返す。
func (r *RestoreAction) TargetNoteID() note.NoteID { return r.NoteID }

// Redo はノートをゴミ箱から復元する（復元のやり直し）。
func (r *RestoreAction) Redo(a *App) error {
	idx := a.findTrashNoteIndex(r.NoteID)
	if idx < 0 {
		return nil
	}

	_, _, err := a.restoreNoteInternal(idx)

	return err
}
