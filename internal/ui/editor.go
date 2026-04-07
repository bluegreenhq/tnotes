package ui

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/note"
)

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
	textarea  textarea.Model
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
	ta := textarea.New()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetWidth(width - editorPadding)
	ta.SetHeight(height)
	ta.CharLimit = 0

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

// LoadNote はノートをエディタに読み込む。
func (e *Editor) LoadNote(n note.Note) {
	e.noteID = n.ID
	e.original = n.Body
	e.textarea.SetValue(n.Body)
	e.ClearSelection()
	e.UndoMgr.Clear()
}

// NoteID は現在編集中のノートIDを返す。
func (e *Editor) NoteID() note.NoteID { return e.noteID }

// Value はテキストエリアの現在の値を返す。
func (e *Editor) Value() string { return e.textarea.Value() }

// SetValue はテキストエリアの値を設定する。
func (e *Editor) SetValue(s string) { e.textarea.SetValue(s) }

// Dirty は未保存の変更があるかを返す。
func (e *Editor) Dirty() bool { return e.textarea.Value() != e.original }

// MarkClean は現在の値を基準値として記録する。
func (e *Editor) MarkClean() { e.original = e.textarea.Value() }

// Focus はエディタにフォーカスを当てる。
func (e *Editor) Focus() tea.Cmd { return e.textarea.Focus() }

// Blur はエディタのフォーカスを外す。
func (e *Editor) Blur() { e.textarea.Blur() }

// Focused はフォーカス状態を返す。
func (e *Editor) Focused() bool { return e.textarea.Focused() }

// SetSize はサイズを更新する。
func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height
	e.textarea.SetWidth(width - editorPadding)
	e.textarea.SetHeight(height)
}

// ScrollUp はカーソルを n 行上に移動する。フォーカス・readOnly に依存しない。
func (e *Editor) ScrollUp(n int) {
	for range n {
		e.textarea.CursorUp()
	}
}

// ScrollDown はカーソルを n 行下に移動する。フォーカス・readOnly に依存しない。
func (e *Editor) ScrollDown(n int) {
	for range n {
		e.textarea.CursorDown()
	}
}

// SetReadOnly は読み取り専用モードを設定する。
func (e *Editor) SetReadOnly(v bool) { e.readOnly = v }

// ReadOnly は読み取り専用モードかを返す。
func (e *Editor) ReadOnly() bool { return e.readOnly }

// Clear はエディタをクリアする。
func (e *Editor) Clear() {
	e.noteID = ""
	e.original = ""
	e.textarea.SetValue("")
	e.ClearSelection()
}
