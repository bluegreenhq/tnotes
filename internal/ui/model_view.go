package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/app"
)

// View はターミナルに描画する内容を返す。
func (m *Model) View() tea.View {
	now := time.Now()
	v := tea.NewView(m.renderView(now))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	v.ReportFocus = true

	return v
}

func (m *Model) renderView(now time.Time) string {
	if m.layout.width < minWidth {
		return "Terminal too small — please resize to at least 80 columns"
	}

	// フォルダの件数を更新
	m.updateFolderCounts()

	// エディタの未保存状態をノートリストに反映
	if m.Editor.Dirty() {
		m.NoteList.SetDirtyNoteID(m.Editor.NoteID())
	} else {
		m.NoteList.SetDirtyNoteID("")
	}

	noteListView := m.NoteList.View(m.Focus == FocusNoteList, m.hoverSeparator || m.resizing, now, m.FolderList.Visible())

	var body string

	if m.FolderList.Visible() {
		folderView := m.FolderList.View(m.Focus == FocusFolderList, m.hoverFolderSep || m.resizingFolder)
		body = lipgloss.JoinHorizontal(lipgloss.Top, folderView, noteListView, m.Editor.View())
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, noteListView, m.Editor.View())
	}

	m.rebuildFooterButtons()
	footer, footerLines := m.Footer.View(m.errMsg, m.infoMsg, m.layout.width)

	// bodyを正確に height-footerLines 行に切り詰め/パディング
	bodyLines := strings.Split(body, "\n")
	targetBodyLines := m.layout.height - footerLines

	targetBodyLines = max(targetBodyLines, 1)
	if len(bodyLines) > targetBodyLines {
		bodyLines = bodyLines[:targetBodyLines]
	}

	for len(bodyLines) < targetBodyLines {
		bodyLines = append(bodyLines, "")
	}

	m.applyOverlays(bodyLines)

	return strings.Join(bodyLines, "\n") + "\n" + footer
}

func (m *Model) updateFolderCounts() {
	notesCount := len(m.App.ListByFolder(app.DefaultFolder))

	for i := range m.FolderList.folders {
		switch m.FolderList.folders[i].Kind {
		case FolderNotes:
			m.FolderList.folders[i].Count = notesCount
		case FolderTrash:
			m.FolderList.folders[i].Count = len(m.App.ListTrashNotes())
		case FolderUser:
			count, err := m.App.FolderNoteCount(m.FolderList.folders[i].Name)
			if err == nil {
				m.FolderList.folders[i].Count = count
			}
		}
	}
}

func (m *Model) applyOverlays(bodyLines []string) { //nolint:cyclop // overlay dispatch
	if m.FolderList.MenuOpen() {
		menuLines := m.FolderList.PopupMenu.View()
		if m.popup.Anchor() != nil {
			m.overlayAtAnchor(bodyLines, menuLines, m.popup.Anchor())
		} else {
			m.overlayFolderListMenu(bodyLines, menuLines)
		}
	}

	if m.Editor.Header.MenuOpen() {
		menuLines := m.Editor.Header.PopupMenu.View()
		if m.popup.Anchor() != nil {
			m.overlayAtAnchor(bodyLines, menuLines, m.popup.Anchor())
		} else {
			m.overlayEditorHeaderMenu(bodyLines, menuLines)
		}
	}

	if m.Editor.Header.MoveMenuOpen() {
		menuLines := m.Editor.Header.MoveMenu.View()
		if m.popup.Anchor() != nil {
			m.overlayAtAnchor(bodyLines, menuLines, m.popup.Anchor())
		} else {
			m.overlayMoveMenu(bodyLines, menuLines)
		}
	}

	if m.Editor.IsContextMenuOpen() && m.popup.Anchor() != nil {
		menuLines := m.Editor.ContextMenu.View()
		m.overlayAtAnchor(bodyLines, menuLines, m.popup.Anchor())
	}

	if m.Footer.MenuOpen() {
		menuLines := m.Footer.PopupMenu.View()
		m.overlayMenu(bodyLines, menuLines)
	}

	if m.confirmDialog != nil {
		m.overlayConfirmDialog(bodyLines)
	}

	if m.helpOverlay != nil {
		m.overlayHelpOverlay(bodyLines)
	}
}

// overlayHelpOverlay はショートカットヘルプをオーバーレイする。
func (m *Model) overlayHelpOverlay(bodyLines []string) {
	m.helpOverlay.SetScreenSize(m.layout.width, m.layout.BodyHeight())

	g := m.helpOverlay.Geometry()
	dialogLines := strings.Split(m.helpOverlay.View(), "\n")

	tui.OverlayLines(bodyLines, dialogLines, g.StartX, g.StartY)
}

// overlayAtAnchor はメニューを指定座標にオーバーレイする。
// 画面端でメニューがはみ出す場合は左方向・上方向にフォールバックする。
func (m *Model) overlayAtAnchor(bodyLines []string, menuLines []string, anchor *menuAnchor) {
	if len(menuLines) == 0 {
		return
	}

	menuWidth := lipgloss.Width(menuLines[0])
	x, y := tui.ClampMenuOrigin(menuWidth, len(menuLines), anchor.x, anchor.y, m.layout.width, m.layout.BodyHeight())

	tui.OverlayLines(bodyLines, menuLines, x, y)
}

// overlayFolderListMenu はフォルダリストのmoreメニューをオーバーレイする。
// メニューはヘッダー直下、フォルダリスト領域の右端寄せで表示する。
func (m *Model) overlayFolderListMenu(bodyLines []string, menuLines []string) {
	if len(menuLines) == 0 {
		return
	}

	menuX := max(m.FolderList.MenuLeftX(), 0)

	tui.OverlayLines(bodyLines, menuLines, menuX, folderListHeaderLines)
}

// overlayEditorHeaderMenu はエディタヘッダーメニューをオーバーレイする。
// メニューはヘッダーの直下、エディタ領域の右端寄せで表示する。
func (m *Model) overlayEditorHeaderMenu(bodyLines []string, menuLines []string) {
	if len(menuLines) == 0 {
		return
	}

	editorStartX := m.layout.EditorStartX()
	menuX := editorStartX + m.Editor.Header.MenuLeftX()

	menuX = max(menuX, editorStartX)

	tui.OverlayLines(bodyLines, menuLines, menuX, editorHeaderMenuTopY)
}

// overlayMoveMenu は移動先メニューをオーバーレイする。
// メニュー右端を ⋯ ボタンの右端（headerWidth - moreButtonOffset + 1）に揃える。
func (m *Model) overlayMoveMenu(bodyLines []string, menuLines []string) {
	if len(menuLines) == 0 {
		return
	}

	editorStartX := m.layout.EditorStartX()
	menuX := max(editorStartX+m.Editor.Header.MoveMenuLeftX(), editorStartX)

	tui.OverlayLines(bodyLines, menuLines, menuX, editorHeaderMenuTopY)
}

// overlayConfirmDialog はフォルダ削除確認ダイアログをオーバーレイする。
func (m *Model) overlayConfirmDialog(bodyLines []string) {
	m.confirmDialog.SetScreenSize(m.layout.width, m.layout.BodyHeight())

	rendered := m.confirmDialog.View()
	g := tui.CalcOverlayGeometry(rendered, m.layout.width, m.layout.BodyHeight(), 0, 0, 0)

	tui.OverlayLines(bodyLines, strings.Split(rendered, "\n"), g.StartX, g.StartY)
}

// overlayMenu はノート一覧領域にメニューをオーバーレイする。
// bodyLines の下端（フッターの直上）にメニューを重ねる。
func (m *Model) overlayMenu(bodyLines []string, menuLines []string) {
	if len(menuLines) == 0 {
		return
	}

	// メニューをbodyの下端に配置
	startY := max(len(bodyLines)-len(menuLines), 0)

	for i, menuLine := range menuLines {
		y := startY + i
		if y >= len(bodyLines) {
			break
		}

		// メニュー行の前にスペース1つを付加
		bodyLines[y] = " " + menuLine
	}
}
