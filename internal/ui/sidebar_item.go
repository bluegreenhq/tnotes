package ui

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/utils"
)

const (
	itemHeight  = 4 // title + preview + date + separator
	itemPadding = 4 // 左パディング "   " + 余白
)

var (
	selectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15")).
				Bold(true)
	normalItemStyle      = lipgloss.NewStyle()
	previewStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	dateStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedPreviewStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Background(lipgloss.Color("4"))
	selectedDateStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Background(lipgloss.Color("4"))
	sectionHeaderStyle   = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Bold(true)
)

// sidebarRow はサイドバーの1行分のデータ（セクションヘッダーまたはノートアイテム）。
type sidebarRow struct {
	isHeader  bool
	label     string
	note      note.Note
	noteIndex int
}

func newHeaderRow(label string) sidebarRow {
	return sidebarRow{isHeader: true, label: label, note: note.ZeroNote(), noteIndex: -1}
}

func newNoteRow(n note.Note, idx int) sidebarRow {
	return sidebarRow{isHeader: false, label: "", note: n, noteIndex: idx}
}

// renderItem はノートアイテム1件の描画文字列を返す。
func renderItem(n note.Note, selected bool, width int, now time.Time) string {
	title := utils.Truncate(n.Title(), width-itemPadding)
	preview := utils.Truncate(n.Preview(), width-itemPadding)
	dateStr := utils.FormatDate(n.UpdatedAt, now)

	var b strings.Builder
	if selected {
		b.WriteString(selectedItemStyle.Width(width).Render("   " + title))
		b.WriteString("\n")
		b.WriteString(selectedPreviewStyle.Width(width).Render("   " + preview))
		b.WriteString("\n")
		b.WriteString(selectedDateStyle.Width(width).Render("   " + dateStr))
	} else {
		b.WriteString(normalItemStyle.Width(width).Render("   " + title))
		b.WriteString("\n")
		b.WriteString(previewStyle.Width(width).Render("   " + preview))
		b.WriteString("\n")
		b.WriteString(dateStyle.Width(width).Render("   " + dateStr))
	}

	b.WriteString("\n")
	b.WriteString("\n")

	return b.String()
}
