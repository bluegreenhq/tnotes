package ui

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/utils"
)

const (
	itemHeight  = 3 // title + date&preview + separator
	itemPadding = 4 // 左パディング "   " + 余白
)

var (
	selectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15")).
				Bold(true)
	normalItemStyle      = lipgloss.NewStyle().Bold(true)
	dateStyle            = lipgloss.NewStyle()
	previewStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedDateStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))
	selectedPreviewStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Background(lipgloss.Color("4"))
	sectionHeaderStyle   = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Bold(true)
	dirtyDotStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	selectedDirtyDotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Background(lipgloss.Color("4"))
)

// noteListRow はノート一覧の1行分のデータ（セクションヘッダーまたはノートアイテム）。
type noteListRow struct {
	isHeader  bool
	label     string
	note      note.Note
	noteIndex int
}

func newHeaderRow(label string) noteListRow {
	return noteListRow{isHeader: true, label: label, note: note.ZeroNote(), noteIndex: -1}
}

func newNoteRow(n note.Note, idx int) noteListRow {
	return noteListRow{isHeader: false, label: "", note: n, noteIndex: idx}
}

// renderItem はノートアイテム1件の描画文字列を返す。
func renderItem(n note.Note, selected, dirty bool, width int, now time.Time) string {
	title := utils.Truncate(n.Title(), width-itemPadding)
	dateStr := utils.FormatDate(n.UpdatedAt, now)

	datePart := "   " + dateStr
	previewMaxWidth := width - runewidth.StringWidth(datePart) - 1 // -1 for space

	preview := ""
	if previewMaxWidth > 0 {
		preview = utils.Truncate(n.Preview(), previewMaxWidth)
	}

	firstLine := renderItemFirstLine(title, selected, dirty)
	secondLine := renderItemSecondLine(datePart, preview, selected)

	var b strings.Builder

	if selected {
		b.WriteString(lipgloss.NewStyle().Width(width).Background(lipgloss.Color("4")).Render(firstLine))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Width(width).Background(lipgloss.Color("4")).Render(secondLine))
	} else {
		b.WriteString(lipgloss.NewStyle().Width(width).Render(firstLine))
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Width(width).Render(secondLine))
	}

	b.WriteString("\n")
	b.WriteString("\n")

	return b.String()
}

func renderItemFirstLine(title string, selected, dirty bool) string {
	if selected {
		var dot string
		if dirty {
			dot = selectedDirtyDotStyle.Render(" ●") + selectedItemStyle.Render(" ")
		} else {
			dot = selectedItemStyle.Render("   ")
		}

		return dot + selectedItemStyle.Render(title)
	}

	var dot string
	if dirty {
		dot = dirtyDotStyle.Render(" ●") + " "
	} else {
		dot = "   "
	}

	return dot + normalItemStyle.Render(title)
}

func renderItemSecondLine(datePart, preview string, selected bool) string {
	if selected {
		line := selectedDateStyle.Render(datePart)
		if preview != "" {
			line += selectedPreviewStyle.Render(" " + preview)
		}

		return line
	}

	line := dateStyle.Render(datePart)
	if preview != "" {
		line += previewStyle.Render(" " + preview)
	}

	return line
}
