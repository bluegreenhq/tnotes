package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

// helpCloseBtnRow は✕ボタンを配置する行（0始まり）。border上(0) の1つ下 = paddingTop行。
const helpCloseBtnRow = 1

const (
	helpOverlayWidth    = 42
	helpOverlayPaddingH = 2
	helpKeyMinWidth     = 15
)

// View はオーバーレイの描画内容を返す。
func (h *HelpOverlay) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	b.WriteString(titleStyle.Render("Shortcuts"))
	b.WriteString("\n")

	sectionTitleStyle := lipgloss.NewStyle().Bold(true)
	hintStyle := lipgloss.NewStyle().Faint(true)

	for i, section := range h.sections {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(sectionTitleStyle.Render(section.Title))
		b.WriteString("\n")

		keyWidth := h.keyWidth(section)

		for _, item := range section.Items {
			padded := item.Key + strings.Repeat(" ", keyWidth-lipgloss.Width(item.Key))
			b.WriteString(padded + item.Description + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("Press Esc/?/Ctrl+Shift+/ to close"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(helpOverlayPaddingH).
		PaddingRight(helpOverlayPaddingH).
		Width(helpOverlayWidth)

	rendered := boxStyle.Render(b.String())

	return h.overlayCloseButton(rendered)
}

// overlayCloseButton はレンダリング済みオーバーレイの右上（paddingTop行、border右の直前）に✕を重ねる。
func (h *HelpOverlay) overlayCloseButton(rendered string) string {
	lines := strings.Split(rendered, "\n")
	if len(lines) <= helpCloseBtnRow {
		return rendered
	}

	style := buttonStyle
	if h.closeHover {
		style = buttonHoverStyle
	}

	closeStr := style.Render("✕")
	lineWidth := lipgloss.Width(lines[helpCloseBtnRow])
	// border右(1文字) の直前に✕を配置
	btnX := lineWidth - 3 //nolint:mnd // border右(1) + padding右(1) の内側

	lines[helpCloseBtnRow] = tui.ComposeLine(lines[helpCloseBtnRow], closeStr, btnX, 1)

	return strings.Join(lines, "\n")
}

// keyWidth はセクション内のキー列の表示幅を返す。
func (h *HelpOverlay) keyWidth(section HelpSection) int {
	maxLen := 0

	for _, item := range section.Items {
		w := lipgloss.Width(item.Key)
		if w > maxLen {
			maxLen = w
		}
	}

	const keyGap = 2

	return max(maxLen+keyGap, helpKeyMinWidth)
}
