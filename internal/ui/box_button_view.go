package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	boxBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	boxLabelStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	boxLabelHover  = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
)

// ViewTop は上罫線行（┌────┐）を返す。
func (b *BoxButton) ViewTop() string {
	inner := b.innerLabel()

	return boxBorderStyle.Render("┌" + strings.Repeat("─", len(inner)) + "┐")
}

// ViewMiddle はラベル行（│ label │）を返す。
func (b *BoxButton) ViewMiddle() string {
	inner := b.innerLabel()

	var labelStyled string
	if b.hovered {
		labelStyled = boxLabelHover.Render(inner)
	} else {
		labelStyled = boxLabelStyle.Render(inner)
	}

	return boxBorderStyle.Render("│") + labelStyled + boxBorderStyle.Render("│")
}

// ViewBottom は下罫線行（└────┘）を返す。
func (b *BoxButton) ViewBottom() string {
	inner := b.innerLabel()

	return boxBorderStyle.Render("└" + strings.Repeat("─", len(inner)) + "┘")
}
