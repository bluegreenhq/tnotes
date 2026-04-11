package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

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
	if m.width < minWidth {
		return "Terminal too small — please resize to at least 80 columns"
	}

	// フォルダの件数を更新
	m.updateFolderCounts()

	noteListView := m.NoteList.View(m.Focus == FocusNoteList, m.hoverSeparator || m.resizing, now, m.FolderList.Visible())

	var body string

	if m.FolderList.Visible() {
		folderView := m.FolderList.View(m.Focus == FocusFolderList, m.hoverFolderSep || m.resizingFolder)
		body = lipgloss.JoinHorizontal(lipgloss.Top, folderView, noteListView, m.Editor.View())
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, noteListView, m.Editor.View())
	}

	m.rebuildFooterButtons()
	footer, footerLines := m.Footer.View(m.errMsg, m.infoMsg, m.width)

	// bodyを正確に height-footerLines 行に切り詰め/パディング
	bodyLines := strings.Split(body, "\n")
	targetBodyLines := m.height - footerLines

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
			m.FolderList.folders[i].Count = len(m.App.TrashNotes)
		case FolderUser:
			count, err := m.App.FolderNoteCount(m.FolderList.folders[i].Name)
			if err == nil {
				m.FolderList.folders[i].Count = count
			}
		}
	}
}

func (m *Model) applyOverlays(bodyLines []string) {
	if m.FolderList.MenuOpen() {
		menuLines := m.FolderList.PopupMenu.View()
		if m.menuAnchor != nil {
			m.overlayAtAnchor(bodyLines, menuLines, m.menuAnchor)
		} else {
			m.overlayFolderListMenu(bodyLines, menuLines)
		}
	}

	if m.Editor.Header.MenuOpen() {
		menuLines := m.Editor.Header.PopupMenu.View()
		if m.menuAnchor != nil {
			m.overlayAtAnchor(bodyLines, menuLines, m.menuAnchor)
		} else {
			m.overlayEditorHeaderMenu(bodyLines, menuLines)
		}
	}

	if m.Editor.Header.MoveMenuOpen() {
		menuLines := m.Editor.Header.MoveMenu.View()
		m.overlayMoveMenu(bodyLines, menuLines)
	}

	if m.Editor.IsContextMenuOpen() && m.menuAnchor != nil {
		menuLines := m.Editor.ContextMenu.View()
		m.overlayAtAnchor(bodyLines, menuLines, m.menuAnchor)
	}

	if m.Footer.MenuOpen() {
		menuLines := m.Footer.PopupMenu.View()
		m.overlayMenu(bodyLines, menuLines)
	}

	if m.confirmDialog != nil {
		m.overlayConfirmDialog(bodyLines)
	}
}

// overlayAtAnchor はメニューを指定座標にオーバーレイする。
// 画面端でメニューがはみ出す場合は左方向・上方向にフォールバックする。
func (m *Model) overlayAtAnchor(bodyLines []string, menuLines []string, anchor *menuAnchor) {
	if len(menuLines) == 0 {
		return
	}

	menuWidth := lipgloss.Width(menuLines[0])
	x, y := m.clampAnchor(anchor, menuWidth, len(menuLines))

	menuRight := x + menuWidth

	for i, menuLine := range menuLines {
		row := y + i
		if row < 0 || row >= len(bodyLines) {
			continue
		}

		truncated := ansi.Truncate(bodyLines[row], x, "")
		w := lipgloss.Width(truncated)
		padded := truncated + strings.Repeat(" ", x-w)
		rest := ansi.TruncateLeft(bodyLines[row], menuRight, "")
		bodyLines[row] = padded + menuLine + rest
	}
}

// overlayFolderListMenu はフォルダリストのmoreメニューをオーバーレイする。
// メニューはヘッダー直下、フォルダリスト領域の右端寄せで表示する。
func (m *Model) overlayFolderListMenu(bodyLines []string, menuLines []string) {
	if len(menuLines) == 0 {
		return
	}

	menuWidth := m.FolderList.PopupMenu.Width()
	menuX := m.FolderList.Width() - folderListBorderWidth - menuWidth

	menuX = max(menuX, 0)

	startY := folderListHeaderLines // ヘッダー直下

	menuRight := menuX + menuWidth

	for i, menuLine := range menuLines {
		y := startY + i
		if y >= len(bodyLines) {
			break
		}

		truncated := ansi.Truncate(bodyLines[y], menuX, "")
		w := lipgloss.Width(truncated)
		padded := truncated + strings.Repeat(" ", menuX-w)
		rest := ansi.TruncateLeft(bodyLines[y], menuRight, "")
		bodyLines[y] = padded + menuLine + rest
	}
}

// overlayEditorHeaderMenu はエディタヘッダーメニューをオーバーレイする。
// メニューはヘッダーの直下、エディタ領域の右端寄せで表示する。
func (m *Model) overlayEditorHeaderMenu(bodyLines []string, menuLines []string) {
	if len(menuLines) == 0 {
		return
	}

	editorStartX := m.noteListOffset() + m.noteListWidth
	menuWidth := m.Editor.Header.PopupMenu.Width()
	menuX := editorStartX + m.Editor.Header.Width() - menuWidth

	menuX = max(menuX, editorStartX)

	startY := editorHeaderMenuTopY // セパレーター行に重ねる

	for i, menuLine := range menuLines {
		y := startY + i
		if y >= len(bodyLines) {
			break
		}

		// ANSI エスケープシーケンスを考慮して視覚幅で切り詰め
		truncated := ansi.Truncate(bodyLines[y], menuX, "")
		w := lipgloss.Width(truncated)
		padded := truncated + strings.Repeat(" ", menuX-w)
		bodyLines[y] = padded + menuLine
	}
}

// overlayMoveMenu は移動先メニューをオーバーレイする。
// メニュー右端を ⋯ ボタンの右端（headerWidth - moreButtonOffset + 1）に揃える。
func (m *Model) overlayMoveMenu(bodyLines []string, menuLines []string) {
	if len(menuLines) == 0 {
		return
	}

	editorStartX := m.noteListOffset() + m.noteListWidth
	menuWidth := m.Editor.Header.MoveMenu.Width()
	menuX := editorStartX + m.Editor.Header.Width() - moreButtonOffset + 1 - menuWidth

	menuX = max(menuX, editorStartX)

	startY := editorHeaderMenuTopY

	for i, menuLine := range menuLines {
		y := startY + i
		if y >= len(bodyLines) {
			break
		}

		truncated := ansi.Truncate(bodyLines[y], menuX, "")
		w := lipgloss.Width(truncated)
		padded := truncated + strings.Repeat(" ", menuX-w)
		bodyLines[y] = padded + menuLine
	}
}

// overlayConfirmDialog はフォルダ削除確認ダイアログをオーバーレイする。
func (m *Model) overlayConfirmDialog(bodyLines []string) {
	rendered := m.confirmDialog.View()
	dialogLines := strings.Split(rendered, "\n")

	const centerDivisor = 2

	// 画面中央に配置
	startY := max((len(bodyLines)-len(dialogLines))/centerDivisor, 0)
	startX := max((m.width-lipgloss.Width(dialogLines[0]))/centerDivisor, 0)

	for i, dLine := range dialogLines {
		y := startY + i
		if y >= len(bodyLines) {
			break
		}

		truncated := ansi.Truncate(bodyLines[y], startX, "")
		w := lipgloss.Width(truncated)
		padded := truncated + strings.Repeat(" ", startX-w)
		bodyLines[y] = padded + dLine
	}
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
