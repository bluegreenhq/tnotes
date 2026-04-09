package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

var noteListStyle = lipgloss.NewStyle().
	BorderRight(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("8"))

// View はノート一覧の描画内容を返す。
func (s *NoteList) View(focused bool, hoverSeparator bool, now time.Time, folderVisible bool) string {
	contentWidth := s.width - noteListBorderWidth

	var b strings.Builder

	var titleStr string
	if folderVisible {
		titleStr = fmt.Sprintf(" %s (%d)", s.title, len(s.notes))
	} else {
		titleStr = fmt.Sprintf(" ≡ %s (%d)", s.title, len(s.notes))
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Width(contentWidth).Render(titleStr))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", contentWidth))
	b.WriteString("\n")

	rows := s.buildRows(now)
	visEnd := visibleEndRow(rows, s.offset, s.visibleLines())

	usedLines := 2

	for i := s.offset; i < visEnd; i++ {
		row := rows[i]
		if row.isHeader {
			b.WriteString(sectionHeaderStyle.Width(contentWidth).Render(" " + row.label))
			b.WriteString("\n")

			line := " " + strings.Repeat("─", contentWidth-sectionLinePadding)
			b.WriteString(sectionHeaderStyle.Width(contentWidth).Render(line))
			b.WriteString("\n")

			usedLines += sectionHeaderHeight

			continue
		}

		b.WriteString(renderItem(row.note, row.noteIndex == s.selected, contentWidth, now))

		usedLines += itemHeight
	}

	for i := usedLines; i < s.height; i++ {
		b.WriteString("\n")
	}

	style := noteListStyle

	if focused {
		style = style.BorderForeground(lipgloss.Color("4"))
	}

	if hoverSeparator {
		style = style.BorderStyle(lipgloss.ThickBorder())
	}

	return style.Width(s.width).Height(s.height).Render(b.String())
}
