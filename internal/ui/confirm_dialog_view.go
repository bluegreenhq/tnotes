package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	confirmDialogPaddingH = 2
	confirmDialogWidth    = 40
	confirmButtonGap      = "  "
)

// View はダイアログの描画内容を返す。
func (d *ConfirmDialog) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	b.WriteString(titleStyle.Render(d.Title))
	b.WriteString("\n")
	b.WriteString(d.Detail)
	b.WriteString("\n\n")

	yesBtn := d.renderButton(d.confirmBtn, d.hoverYes)
	noBtn := d.renderButton(d.cancelBtn, d.hoverNo)
	b.WriteString(yesBtn + confirmButtonGap + noBtn)

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")).
		Padding(1, confirmDialogPaddingH).
		Width(confirmDialogWidth)

	return style.Render(b.String())
}

// DialogLines はレンダリング後の行数を返す。
func (d *ConfirmDialog) DialogLines() int {
	// padding上1 + border上1 + content + padding下1 + border下1
	return d.contentLines() + confirmBorderPad
}

func (d *ConfirmDialog) renderButton(label string, hovered bool) string {
	if hovered {
		return buttonHoverStyle.Render(label)
	}

	return buttonStyle.Render(label)
}
