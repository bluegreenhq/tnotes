package ui

import "time"

const snapshotInterval = 500 * time.Millisecond

// EditorSnapshot はエディタの状態スナップショットを表す。
type EditorSnapshot struct {
	Text       string
	CursorLine int
	CursorCol  int
}

// EditorUndoManager はエディタのundo/redoスタックを管理する。
type EditorUndoManager struct {
	undoStack []EditorSnapshot
	redoStack []EditorSnapshot
	lastSave  time.Time
}

// NewEditorUndoManager は新しい EditorUndoManager を生成する。
func NewEditorUndoManager() *EditorUndoManager {
	return &EditorUndoManager{undoStack: nil, redoStack: nil, lastSave: time.Time{}}
}

// Save はスナップショットをundoスタックに追加し、redoスタックをクリアする。
func (m *EditorUndoManager) Save(snap EditorSnapshot) {
	m.undoStack = append(m.undoStack, snap)
	m.redoStack = nil
}

// MaybeSave は前回保存から snapshotInterval 以上経過していればスナップショットを保存する。
func (m *EditorUndoManager) MaybeSave(snap EditorSnapshot, now time.Time) {
	if !m.lastSave.IsZero() && now.Sub(m.lastSave) < snapshotInterval {
		return
	}

	m.Save(snap)
	m.lastSave = now
}

// ForceSave は間隔に関係なくスナップショットを保存し、タイマーをリセットする。
func (m *EditorUndoManager) ForceSave(snap EditorSnapshot, now time.Time) {
	m.Save(snap)
	m.lastSave = now
}

// PushUndo はスナップショットをundoスタックに追加する（redoスタックはクリアしない）。
func (m *EditorUndoManager) PushUndo(snap EditorSnapshot) {
	m.undoStack = append(m.undoStack, snap)
}

// PushRedo はスナップショットをredoスタックに追加する。
func (m *EditorUndoManager) PushRedo(snap EditorSnapshot) {
	m.redoStack = append(m.redoStack, snap)
}

// PopUndo はundoスタックから最新のスナップショットを取り出す。空ならnilを返す。
func (m *EditorUndoManager) PopUndo() *EditorSnapshot {
	if len(m.undoStack) == 0 {
		return nil
	}

	snap := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]

	return &snap
}

// PopRedo はredoスタックから最新のスナップショットを取り出す。空ならnilを返す。
func (m *EditorUndoManager) PopRedo() *EditorSnapshot {
	if len(m.redoStack) == 0 {
		return nil
	}

	snap := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]

	return &snap
}

// CanUndo はundoスタックにスナップショットがあるかを返す。
func (m *EditorUndoManager) CanUndo() bool { return len(m.undoStack) > 0 }

// CanRedo はredoスタックにスナップショットがあるかを返す。
func (m *EditorUndoManager) CanRedo() bool { return len(m.redoStack) > 0 }

// Clear はすべてのスタックをクリアする。
func (m *EditorUndoManager) Clear() {
	m.undoStack = nil
	m.redoStack = nil
	m.lastSave = time.Time{}
}
