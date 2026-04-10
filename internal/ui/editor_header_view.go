package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// View はヘッダー行を描画する。
func (h *EditorHeader) View() string {
	showNew := !h.trashMode
	showMore := h.hasNote

	var left, right string

	if showNew {
		style := buttonStyle
		if h.hoverNew {
			style = buttonHoverStyle
		}

		left = " " + style.Render("+")
	}

	if showMore {
		style := buttonStyle
		if h.hoverMore {
			style = buttonHoverStyle
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
