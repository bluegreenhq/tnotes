package ui

import "github.com/bluegreenhq/tnotes/internal/note"

const editorPadding = 2 // lipgloss Padding(0,1) の左右合計

// SelectionAnchor はテキスト内の位置を表す。
type SelectionAnchor struct {
	Line   int
	Column int
}

// selBefore は a が b より前にあるかを返す。
func selBefore(a, b SelectionAnchor) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}

	return a.Column < b.Column
}

// Editor はテキスト編集ペインの状態を表す。
type Editor struct {
	textarea  simpleTextArea
	noteID    note.NoteID
	original  string
	width     int
	height    int
	readOnly  bool
	selecting bool // ドラッグ中か
	selStart  *SelectionAnchor
	selEnd    *SelectionAnchor
	UndoMgr   *EditorUndoManager
}

// NewEditor は新しい Editor を生成する。
func NewEditor(width, height int) Editor {
	ta := newSimpleTextArea()
	ta.SetWidth(width - editorPadding)
	ta.SetHeight(height)

	return Editor{
		textarea:  ta,
		noteID:    "",
		original:  "",
		width:     width,
		height:    height,
		readOnly:  false,
		selecting: false,
		selStart:  nil,
		selEnd:    nil,
		UndoMgr:   NewEditorUndoManager(),
	}
}

// NoteID は現在編集中のノートIDを返す。
func (e *Editor) NoteID() note.NoteID { return e.noteID }

// Value はテキストエリアの現在の値を返す。
func (e *Editor) Value() string { return e.textarea.Value() }

// Dirty は未保存の変更があるかを返す。
func (e *Editor) Dirty() bool { return e.textarea.Value() != e.original }

// Focused はフォーカス状態を返す。
func (e *Editor) Focused() bool { return e.textarea.Focused() }

// ReadOnly は読み取り専用モードかを返す。
func (e *Editor) ReadOnly() bool { return e.readOnly }

// HasSelection は選択範囲があるかを返す。
func (e *Editor) HasSelection() bool {
	return e.selStart != nil && e.selEnd != nil && *e.selStart != *e.selEnd
}

// Selecting はドラッグ中かを返す。
func (e *Editor) Selecting() bool { return e.selecting }
