package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/app"
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
	maxNoteListPct   = 80
	percentDivisor   = 100
	defaultHeight    = 24
	infoMsgDuration  = 3 * time.Second
)

// clearInfoMsg は一定時間後に情報メッセージを消すためのメッセージ。
type clearInfoMsg struct {
	id int
}

// Model はUIの状態を表す。
type Model struct {
	App             *app.App
	NoteList        NoteList
	Editor          Editor
	Footer          Footer
	Focus           FocusArea
	FolderList      FolderList
	folderListWidth int
	resizingFolder  bool
	hoverFolderSep  bool
	noteListWidth   int
	width           int
	height          int
	resizing        bool
	hoverSeparator  bool
	errMsg          string
	infoMsg         string
	infoMsgID       int
	indexModTime    time.Time
}

var _ tea.Model = (*Model)(nil)

// InitialModel は初期状態の Model を生成する。
func InitialModel(a *app.App, noWrap bool) *Model {
	m := &Model{
		App:             a,
		NoteList:        NewNoteList(a.Notes, defaultNoteListW, defaultHeight),
		Editor:          NewEditor(minWidth-defaultNoteListW, defaultHeight, noWrap),
		Footer:          NewFooter(),
		Focus:           FocusNoteList,
		FolderList:      NewFolderList(defaultFolderListW, defaultHeight),
		folderListWidth: defaultFolderListW,
		resizingFolder:  false,
		hoverFolderSep:  false,
		noteListWidth:   defaultNoteListW,
		width:           0,
		height:          0,
		resizing:        false,
		hoverSeparator:  false,
		errMsg:          "",
		infoMsg:         "",
		infoMsgID:       0,
		indexModTime:    time.Time{},
	}

	return m
}

// Init は初回のコマンドを返す。
func (m *Model) Init() tea.Cmd {
	mt, err := m.App.IndexModTime()
	if err != nil {
		m.errMsg = err.Error()
	} else {
		m.indexModTime = mt
	}

	if len(m.App.Notes) > 0 {
		m.loadSelectedNote()
	}

	return nil
}

// NoteListWidth は現在のノート一覧幅を返す。
func (m *Model) NoteListWidth() int { return m.noteListWidth }

func (m *Model) maxNoteListWidth() int {
	return m.width * maxNoteListPct / percentDivisor
}

func (m *Model) rebuildFooterButtons() {
	m.Footer.RebuildButtons(FooterState{
		TrashMode:    m.App.TrashMode,
		TrashCount:   len(m.App.TrashNotes),
		HasSelection: m.Editor.HasSelection(),
		EditorDirty:  m.Editor.Dirty(),
	})
}
