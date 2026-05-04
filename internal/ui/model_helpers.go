package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// FocusArea はフォーカス対象を表す。
type FocusArea int

const (
	// FocusFolderList はフォルダ一覧にフォーカスしている状態。
	FocusFolderList FocusArea = iota
	// FocusNoteList はノート一覧にフォーカスしている状態。
	FocusNoteList
	// FocusEditor はエディタにフォーカスしている状態。
	FocusEditor
)

const (
	minWidth         = 80
	defaultNoteListW = 32
	minNoteListWidth = 20
	minEditorWidth   = 20
	maxNoteListPct   = 80
	percentDivisor   = 100
	defaultHeight    = 24
	infoMsgDuration  = 3 * time.Second
)

// blinkOwner はアプリ固有の CursorBlink 所有者定数。
const (
	blinkOwnerEditor = iota
	blinkOwnerFolderList
	blinkOwnerSearch
)

const searchDebounceDuration = 150 * time.Millisecond

// clearInfoMsg は一定時間後に情報メッセージを消すためのメッセージ。
type clearInfoMsg struct {
	id int
}

// mouseMsg はマウス位置を持つメッセージ。
type mouseMsg interface {
	Mouse() tea.Mouse
}

// menuAnchor はポップアップメニューのアンカー位置（画面絶対座標）を表す。
type menuAnchor struct {
	x int
	y int
}
