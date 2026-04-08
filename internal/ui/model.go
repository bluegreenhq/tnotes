package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/app"
)

// FocusArea はフォーカス対象を表す。
type FocusArea int

const (
	// FocusSidebar はサイドバーにフォーカスしている状態。
	FocusSidebar FocusArea = iota
	// FocusEditor はエディタにフォーカスしている状態。
	FocusEditor
)

const (
	minWidth        = 80
	sidebarWidthPx  = 32
	defaultHeight   = 24
	infoMsgDuration = 3 * time.Second
)

// clearInfoMsg は一定時間後に情報メッセージを消すためのメッセージ。
type clearInfoMsg struct {
	id int
}

// Model はUIの状態を表す。
type Model struct {
	App       *app.App
	Sidebar   Sidebar
	Editor    Editor
	Footer    Footer
	Focus     FocusArea
	width     int
	height    int
	errMsg    string
	infoMsg   string
	infoMsgID int
}

var _ tea.Model = (*Model)(nil)

// InitialModel は初期状態の Model を生成する。
func InitialModel(a *app.App) *Model {
	m := &Model{
		App:       a,
		Sidebar:   NewSidebar(a.Notes, sidebarWidthPx, defaultHeight),
		Editor:    NewEditor(minWidth-sidebarWidthPx, defaultHeight),
		Footer:    Footer{hover: HoverNone, buttons: nil},
		Focus:     FocusSidebar,
		width:     0,
		height:    0,
		errMsg:    "",
		infoMsg:   "",
		infoMsgID: 0,
	}

	return m
}

// Init は初回のコマンドを返す。
func (m *Model) Init() tea.Cmd {
	if len(m.App.Notes) > 0 {
		m.loadSelectedNote()
	}

	return nil
}

func (m *Model) rebuildFooterButtons() {
	m.Footer.RebuildButtons(FooterState{
		TrashMode:    m.App.TrashMode,
		TrashCount:   len(m.App.TrashNotes),
		HasSelection: m.Editor.HasSelection(),
		EditorDirty:  m.Editor.Dirty(),
	})
}
