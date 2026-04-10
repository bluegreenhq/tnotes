package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
	m.FolderList.UpdateCounts(len(m.App.Notes), len(m.App.TrashNotes))

	noteListView := m.NoteList.View(m.Focus == FocusNoteList, m.hoverSeparator || m.resizing, now, m.FolderList.Visible())

	var body string

	if m.FolderList.Visible() {
		folderView := m.FolderList.View(m.Focus == FocusFolderList)
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

	// エディタヘッダーメニューオーバーレイ
	if m.Editor.Header.MenuOpen() {
		menuLines := m.Editor.Header.PopupMenu.View()
		m.overlayEditorHeaderMenu(bodyLines, menuLines)
	}

	// フッターメニューオーバーレイ
	if m.Footer.MenuOpen() {
		menuLines := m.Footer.PopupMenu.View()
		m.overlayMenu(bodyLines, menuLines)
	}

	return strings.Join(bodyLines, "\n") + "\n" + footer
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
