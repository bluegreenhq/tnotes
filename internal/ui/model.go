package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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
	App                 *app.App
	NoteList            NoteList
	Editor              Editor
	Footer              Footer
	Focus               FocusArea
	FolderList          FolderList
	folderListWidth     int
	resizingFolder      bool
	hoverFolderSep      bool
	noteListWidth       int
	width               int
	height              int
	resizing            bool
	hoverSeparator      bool
	errMsg              string
	infoMsg             string
	infoMsgID           int
	indexModTime        time.Time
	confirmDialog       *ConfirmDialog // 削除確認ダイアログ（nil = 非表示）
	confirmDeleteFolder string         // 削除確認中のフォルダ名
}

var _ tea.Model = (*Model)(nil)

// InitialModel は初期状態の Model を生成する。
func InitialModel(a *app.App, noWrap bool) *Model {
	m := &Model{
		App:                 a,
		NoteList:            NewNoteList(a.Notes, defaultNoteListW, defaultHeight),
		Editor:              NewEditor(minWidth-defaultNoteListW, defaultHeight, noWrap),
		Footer:              NewFooter(),
		Focus:               FocusNoteList,
		FolderList:          NewFolderList(defaultFolderListW, defaultHeight),
		folderListWidth:     defaultFolderListW,
		resizingFolder:      false,
		hoverFolderSep:      false,
		noteListWidth:       defaultNoteListW,
		width:               0,
		height:              0,
		resizing:            false,
		hoverSeparator:      false,
		errMsg:              "",
		infoMsg:             "",
		infoMsgID:           0,
		indexModTime:        time.Time{},
		confirmDialog:       nil,
		confirmDeleteFolder: "",
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

// confirmDialogOrigin はダイアログの画面上のコンテンツ左上座標を返す。
// overlayConfirmDialog と同じ起点計算を行い、border + padding 分を加算する。
func (m *Model) confirmDialogOrigin() (int, int) {
	if m.confirmDialog == nil {
		return 0, 0
	}

	rendered := m.confirmDialog.View()
	dialogLines := strings.Split(rendered, "\n")

	const (
		centerDivisor      = 2
		borderPaddingLines = 2 // border上 + padding上
		borderPaddingCols  = 3 // border左1 + padding左2
	)

	bodyHeight := m.height - footerLineCount
	dialogWidth := lipgloss.Width(dialogLines[0])

	sy := max((bodyHeight-len(dialogLines))/centerDivisor, 0) + borderPaddingLines
	sx := max((m.width-dialogWidth)/centerDivisor, 0) + borderPaddingCols

	return sx, sy
}

func (m *Model) rebuildFooterButtons() {
	m.Footer.RebuildButtons(FooterState{
		TrashMode:    m.App.TrashMode,
		TrashCount:   len(m.App.TrashNotes),
		HasSelection: m.Editor.HasSelection(),
		EditorDirty:  m.Editor.Dirty(),
	})
}
