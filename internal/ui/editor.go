package ui

import (
	"time"
	"unicode"

	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/note"
)

const editorPadding = 2 // lipgloss Padding(0,1) の左右合計

// SelectionAnchor はテキスト内の位置を表す。
type SelectionAnchor struct {
	Line   int
	Column int
}

// NewSelectionAnchor は新しい SelectionAnchor を生成する。
func NewSelectionAnchor(line, col int) SelectionAnchor {
	return SelectionAnchor{Line: line, Column: col}
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
	Header          *EditorHeader
	textarea        simpleTextArea
	noteID          note.NoteID
	original        string
	width           int
	height          int
	readOnly        bool
	selecting       bool // ドラッグ中か
	selStart        *SelectionAnchor
	selEnd          *SelectionAnchor
	UndoMgr         *EditorUndoManager
	blink           cursorBlink
	contextMenuOpen bool           // コンテキストメニュー表示中
	ContextMenu     *tui.PopupMenu // コンテキストメニュー
	searchQuery     string         // 検索クエリ
	lastClickTime   time.Time
	lastClickPos    SelectionAnchor
	clickCount      int
}

// NewEditor は新しい Editor を生成する。
func NewEditor(width, height int, noWrap bool) Editor {
	ta := newSimpleTextArea(noWrap)
	ta.SetWidth(width - editorPadding)
	ta.SetHeight(height - editorHeaderHeight)

	return Editor{
		Header:          NewEditorHeader(width),
		textarea:        ta,
		noteID:          "",
		original:        "",
		width:           width,
		height:          height,
		readOnly:        false,
		selecting:       false,
		selStart:        nil,
		selEnd:          nil,
		UndoMgr:         NewEditorUndoManager(),
		blink:           newCursorBlink(blinkOwnerEditor),
		contextMenuOpen: false,
		ContextMenu:     tui.NewPopupMenu(nil),
		searchQuery:     "",
		lastClickTime:   time.Time{},
		lastClickPos:    NewSelectionAnchor(0, 0),
		clickCount:      0,
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

// BlinkVisible はカーソルの表示状態を返す。
func (e *Editor) BlinkVisible() bool { return e.blink.Visible() }

// SetSearchQuery は検索クエリを設定する。
func (e *Editor) SetSearchQuery(q string) { e.searchQuery = q }

// runeClass はワード選択のための文字種分類を表す。
type runeClass int

const (
	classWord     runeClass = iota // ASCII英数字 + '_'
	classHiragana                  // ひらがな
	classKatakana                  // カタカナ
	classKanji                     // 漢字
	classSpace                     // 空白
	classPunct                     // その他
)

// classifyRune はルーンの文字種を返す。
func classifyRune(r rune) runeClass {
	if unicode.Is(unicode.Hiragana, r) {
		return classHiragana
	}

	if unicode.Is(unicode.Katakana, r) || r == 'ー' {
		return classKatakana
	}

	if unicode.Is(unicode.Han, r) {
		return classKanji
	}

	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
		return classWord
	}

	if unicode.IsSpace(r) {
		return classSpace
	}

	return classPunct
}
