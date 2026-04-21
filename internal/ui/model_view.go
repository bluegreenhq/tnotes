package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"
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

	noteSepActive := m.hoverSeparator || m.dragTarget == dragNoteSeparator
	noteListView := m.NoteList.View(m.Focus == FocusNoteList, noteSepActive, now, m.FolderList.Visible())

	var body string

	if m.FolderList.Visible() {
		folderSepActive := m.hoverFolderSep || m.dragTarget == dragFolderSeparator
		folderView := m.FolderList.View(m.Focus == FocusFolderList, folderSepActive)
		body = lipgloss.JoinHorizontal(lipgloss.Top, folderView, noteListView, m.Editor.View())
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, noteListView, m.Editor.View())
	}

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

func (m *Model) applyOverlays(bodyLines []string) {
	m.popup.RenderOverlays(bodyLines)

	if m.Footer.MenuOpen() {
		m.overlayFooterMenu(bodyLines, m.Footer.PopupMenu.View())
	}

	if m.FolderList.ConfirmDialogVisible() {
		rendered := m.FolderList.ConfirmDialogView()
		g := tui.CalcOverlayGeometry(rendered, m.layout.width, m.layout.BodyHeight(), 0, 0, 0)
		tui.OverlayLines(bodyLines, strings.Split(rendered, "\n"), g.StartX, g.StartY)
	}

	if m.helpOverlay != nil {
		g := m.helpOverlay.Geometry()
		tui.OverlayLines(bodyLines, strings.Split(m.helpOverlay.View(), "\n"), g.StartX, g.StartY)
	}
}

// overlayFooterMenu はフッターメニューを bodyLines の下端にオーバーレイする。
func (m *Model) overlayFooterMenu(bodyLines []string, menuLines []string) {
	if len(menuLines) == 0 {
		return
	}

	startY := max(len(bodyLines)-len(menuLines), 0)

	for i, menuLine := range menuLines {
		y := startY + i
		if y >= len(bodyLines) {
			break
		}

		bodyLines[y] = " " + menuLine
	}
}
