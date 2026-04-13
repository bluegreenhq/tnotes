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

	searchField := h.renderSearchField()

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	searchLen := lipgloss.Width(searchField)
	gap := h.width - leftLen - rightLen - searchLen

	gap = max(gap, 0)

	buttonLine := left + strings.Repeat(" ", gap) + right + searchField
	separator := strings.Repeat("─", max(h.width, 0))

	return buttonLine + "\n" + separator
}

const (
	searchFieldPadding = 4 // 左右スペース×2
	searchIconWidth    = 2 // "⚲" + space
)

func (h *EditorHeader) renderSearchField() string {
	icon := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("⚲")

	if h.searchFocused {
		content := h.searchInput.ViewWithWidth(searchFieldWidth-searchFieldPadding-searchIconWidth, h.searchBlink.Visible())
		inner := " " + icon + " " + content + " "

		return padOrTruncate(inner, searchFieldWidth)
	}

	if h.searchInput.Value() != "" {
		content := h.searchInput.ViewWithWidth(searchFieldWidth-searchFieldPadding-searchIconWidth, false)
		inner := " " + icon + " " + content + " "

		return padOrTruncate(inner, searchFieldWidth)
	}

	// プレースホルダー
	placeholder := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Search")
	inner := " " + icon + " " + placeholder + " "

	return padOrTruncate(inner, searchFieldWidth)
}

func padOrTruncate(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}

	return s + strings.Repeat(" ", width-w)
}
