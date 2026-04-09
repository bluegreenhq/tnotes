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

	sidebarView := m.Sidebar.View(m.Focus == FocusSidebar, m.hoverSeparator || m.resizing, now)
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, m.Editor.View())

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

	return strings.Join(bodyLines, "\n") + "\n" + footer
}
