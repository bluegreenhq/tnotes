package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

	noteListView := m.NoteList.View(m.Focus == FocusNoteList, m.hoverSeparator || m.resizing, now)
	body := lipgloss.JoinHorizontal(lipgloss.Top, noteListView, m.Editor.View())

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

	// メニューオーバーレイ
	if m.Footer.MenuOpen() {
		menuLines := m.Footer.PopupMenuRef().View()
		m.overlayMenu(bodyLines, menuLines)
	}

	return strings.Join(bodyLines, "\n") + "\n" + footer
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
