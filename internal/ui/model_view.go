package ui

import (
	"fmt"
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

	// フォルダリストメニューオーバーレイ
	if m.FolderList.MenuOpen() {
		menuLines := m.FolderList.PopupMenu.View()
		m.overlayFolderListMenu(bodyLines, menuLines)
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

	// フォルダ削除確認ダイアログオーバーレイ
	if m.confirmDeleteFolder != "" {
		m.overlayConfirmDialog(bodyLines)
	}

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

// overlayConfirmDialog はフォルダ削除確認ダイアログをオーバーレイする。
func (m *Model) overlayConfirmDialog(bodyLines []string) {
	msg := fmt.Sprintf("Delete %q?", m.confirmDeleteFolder)
	detail := fmt.Sprintf("%d note(s) will be moved to Trash.", m.confirmDeleteCount)
	buttons := "[Y]es  [N]o"

	content := msg + "\n" + detail + "\n\n" + buttons

	const (
		dialogPaddingH = 2
		dialogWidth    = 40
		centerDivisor  = 2
	)

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")).
		Padding(1, dialogPaddingH).
		Width(dialogWidth)

	rendered := style.Render(content)
	dialogLines := strings.Split(rendered, "\n")

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
