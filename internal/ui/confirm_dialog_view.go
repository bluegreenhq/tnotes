package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	confirmDialogPaddingH = 2
	confirmDialogWidth    = 40
)

// View はダイアログの描画内容を返す。
func (d *ConfirmDialog) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	b.WriteString(titleStyle.Render(d.Title))
	b.WriteString("\n\n")
	b.WriteString(d.Detail)
	b.WriteString("\n\n")

	gap := strings.Repeat(" ", confirmButtonGapCols)
	pad := strings.Repeat(" ", d.buttonPadLeft())
	b.WriteString(pad + d.yesBtn.ViewTop() + gap + d.noBtn.ViewTop())
	b.WriteString("\n")
	b.WriteString(pad + d.yesBtn.ViewMiddle() + gap + d.noBtn.ViewMiddle())
	b.WriteString("\n")
	b.WriteString(pad + d.yesBtn.ViewBottom() + gap + d.noBtn.ViewBottom())

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("9")).
		PaddingTop(1).
		PaddingBottom(0).
		PaddingLeft(confirmDialogPaddingH).
		PaddingRight(confirmDialogPaddingH).
		Width(confirmDialogWidth)

	return style.Render(b.String())
}

// DialogLines はレンダリング後の行数を返す。
func (d *ConfirmDialog) DialogLines() int {
	// padding上1 + border上1 + content + padding下1 + border下1
	return d.contentLines() + confirmBorderPad
}
