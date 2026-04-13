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
	minEditorWidth   = 20
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
	layout              Layout
	resizingFolder      bool
	hoverFolderSep      bool
	resizing            bool
	hoverSeparator      bool
	errMsg              string
	infoMsg             string
	infoMsgID           int
	indexModTime        time.Time
	confirmDialog       *ConfirmDialog // 削除確認ダイアログ（nil = 非表示）
	confirmDeleteFolder string         // 削除確認中のフォルダ名
	popup               PopupCoordinator
	helpOverlay         *HelpOverlay // ショートカットヘルプ（nil = 非表示）
	searchDebounceID    int          // デバウンスタイマーの世代ID
}

var _ tea.Model = (*Model)(nil)

// InitialModel は初期状態の Model を生成する。
func InitialModel(a *app.App, noWrap bool) *Model {
	m := &Model{
		App:                 a,
		NoteList:            NewNoteList(a.ListByFolder(app.DefaultFolder), defaultNoteListW, defaultHeight),
		Editor:              NewEditor(minWidth-defaultNoteListW, defaultHeight, noWrap),
		Footer:              NewFooter(),
		Focus:               FocusNoteList,
		FolderList:          NewFolderList(defaultFolderListW, defaultHeight),
		layout:              NewLayout(defaultFolderListW, defaultNoteListW),
		resizingFolder:      false,
		hoverFolderSep:      false,
		resizing:            false,
		hoverSeparator:      false,
		errMsg:              "",
		infoMsg:             "",
		infoMsgID:           0,
		indexModTime:        time.Time{},
		confirmDialog:       nil,
		confirmDeleteFolder: "",
		popup:               PopupCoordinator{anchor: nil, lastAnchor: nil, entries: nil, layout: nil},
		helpOverlay:         nil,
		searchDebounceID:    0,
	}

	m.popup = m.newPopupCoordinator()

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

	if len(m.App.ListByFolder(app.DefaultFolder)) > 0 {
		m.loadSelectedNote()
	}

	return nil
}

// NoteListWidth は現在のノート一覧幅を返す。
func (m *Model) NoteListWidth() int { return m.layout.noteListWidth }

// HelpVisible はヘルプオーバーレイが表示中かを返す。
func (m *Model) HelpVisible() bool { return m.helpOverlay != nil }

func (m *Model) newPopupCoordinator() PopupCoordinator { //nolint:funlen // メニュー登録の一覧性を優先
	return NewPopupCoordinator(&m.layout, []popupEntry{
		{
			kind:   menuKindEditorContext,
			menu:   func() *PopupMenu { return m.Editor.ContextMenu },
			isOpen: func() bool { return m.Editor.IsContextMenuOpen() },
			close:  func() { m.Editor.CloseContextMenu() },
			execute: func(_ int, _ time.Time) tea.Cmd {
				return nil
			},
			handleAnchorClick: func(relX, relY int) tea.Cmd {
				m.Editor.HandleContextMenuClick(relX, relY)

				return nil
			},
		},
		{
			kind:   menuKindMoveMenu,
			menu:   func() *PopupMenu { return m.Editor.Header.MoveMenu },
			isOpen: func() bool { return m.Editor.Header.MoveMenuOpen() },
			close:  func() { m.Editor.Header.CloseMoveMenu() },
			execute: func(idx int, _ time.Time) tea.Cmd {
				return m.Editor.Header.ExecuteMoveMenuAction(idx)
			},
			handleAnchorClick: func(relX, relY int) tea.Cmd {
				return m.Editor.Header.HandleMoveMenuClick(relX, relY)
			},
		},
		{
			kind:   menuKindEditorHeader,
			menu:   func() *PopupMenu { return m.Editor.Header.PopupMenu },
			isOpen: func() bool { return m.Editor.Header.MenuOpen() },
			close:  func() { m.Editor.Header.CloseMenu() },
			execute: func(idx int, _ time.Time) tea.Cmd {
				return m.Editor.Header.ExecuteMenuAction(idx)
			},
			handleAnchorClick: func(relX, relY int) tea.Cmd {
				return m.Editor.Header.HandleMenuClick(relX, relY)
			},
		},
		{
			kind:   menuKindFolderList,
			menu:   func() *PopupMenu { return m.FolderList.PopupMenu },
			isOpen: func() bool { return m.FolderList.MenuOpen() },
			close:  func() { m.FolderList.CloseMenu() },
			execute: func(idx int, _ time.Time) tea.Cmd {
				return m.handleFolderMenuAction(idx)
			},
			handleAnchorClick: func(relX, relY int) tea.Cmd {
				idx, hit := m.FolderList.PopupMenu.HandleClick(relX, relY)
				m.FolderList.CloseMenu()

				if hit {
					return m.handleFolderMenuAction(idx)
				}

				return nil
			},
		},
		{
			kind:   menuKindFooter,
			menu:   func() *PopupMenu { return m.Footer.PopupMenu },
			isOpen: func() bool { return m.Footer.MenuOpen() },
			close:  func() { m.Footer.CloseMenu() },
			execute: func(idx int, now time.Time) tea.Cmd {
				return m.processFooterMenuAction(idx, now)
			},
			handleAnchorClick: nil,
		},
	})
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

	bodyHeight := m.layout.BodyHeight()
	dialogWidth := lipgloss.Width(dialogLines[0])

	sy := max((bodyHeight-len(dialogLines))/centerDivisor, 0) + borderPaddingLines
	sx := max((m.layout.width-dialogWidth)/centerDivisor, 0) + borderPaddingCols

	return sx, sy
}

func (m *Model) rebuildFooterButtons() {
	m.Footer.RebuildButtons()
}

// isTrashFolder は現在 Trash フォルダを表示しているかを返す。
func (m *Model) isTrashFolder() bool {
	return m.FolderList.SelectedKind() == FolderTrash
}
