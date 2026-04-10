package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	headerButtonStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	headerButtonHoverStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
)

// View はヘッダー行を描画する。
func (h *EditorHeader) View() string {
	showNew := !h.trashMode
	showMore := h.hasNote

	var left, right string

	if showNew {
		style := headerButtonStyle
		if h.hoverNew {
			style = headerButtonHoverStyle
		}

		left = " " + style.Render("+")
	}

	if showMore {
		style := headerButtonStyle
		if h.hoverMore {
			style = headerButtonHoverStyle
		}

		right = style.Render("⋯") + " "
	}

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	gap := h.width - leftLen - rightLen

	gap = max(gap, 0)

	buttonLine := left + strings.Repeat(" ", gap) + right
	separator := strings.Repeat("─", h.width)

	return buttonLine + "\n" + separator
}
