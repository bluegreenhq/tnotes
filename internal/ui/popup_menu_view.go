package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	menuItemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	menuItemHoverStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
	menuItemDisabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	menuBorderStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// View はポップアップメニューを枠付きで描画する。
// 各行は独立した文字列のスライスとして返す。
func (m *PopupMenu) View() []string {
	if len(m.items) == 0 {
		return nil
	}

	innerWidth := m.Width() - menuBorderLines // 左右の枠を除く

	lines := make([]string, 0, m.Height())

	// 上枠
	lines = append(lines, menuBorderStyle.Render("┌"+strings.Repeat("─", innerWidth)+"┐"))

	// 項目
	for i, item := range m.items {
		if i > 0 {
			// 項目間の空行（枠付き）
			emptyLine := menuBorderStyle.Render("│") +
				menuBorderStyle.Render(strings.Repeat(" ", innerWidth)) +
				menuBorderStyle.Render("│")
			lines = append(lines, emptyLine)
		}

		label := fmt.Sprintf(" %-*s", innerWidth-1, item.Label)

		var styled string

		switch {
		case item.Disabled:
			styled = menuItemDisabledStyle.Render(label)
		case i == m.hover:
			styled = menuItemHoverStyle.Render(label)
		default:
			styled = menuItemStyle.Render(label)
		}

		line := menuBorderStyle.Render("│") + styled + menuBorderStyle.Render("│")
		lines = append(lines, line)
	}

	// 下枠
	lines = append(lines, menuBorderStyle.Render("└"+strings.Repeat("─", innerWidth)+"┘"))

	return lines
}
