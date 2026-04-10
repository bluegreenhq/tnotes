package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var folderListStyle = lipgloss.NewStyle().
	BorderRight(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("8"))

// View はフォルダ一覧の描画内容を返す。
func (fl *FolderList) View(focused bool) string {
	contentWidth := fl.width - folderListBorderWidth

	var b strings.Builder

	// ヘッダー: 閉じるボタン + タイトル
	headerStr := " ✕ Folders"
	b.WriteString(lipgloss.NewStyle().Bold(true).Width(contentWidth).Render(headerStr))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", contentWidth))
	b.WriteString("\n")

	// フォルダ一覧
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	for i, folder := range fl.folders {
		name := " " + folder.Name
		count := fmt.Sprintf("%d ", folder.Count)

		padding := max(contentWidth-lipgloss.Width(name)-lipgloss.Width(count), 0)

		if i == fl.selected {
			baseStyle := lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Bold(true)
			nameStr := baseStyle.Foreground(lipgloss.Color("15")).Render(name)
			pad := baseStyle.Render(strings.Repeat(" ", padding))
			countStr := baseStyle.Foreground(lipgloss.Color("7")).Render(count)
			b.WriteString(nameStr + pad + countStr)
		} else {
			nameStr := lipgloss.NewStyle().Render(name)
			pad := strings.Repeat(" ", padding)
			countStr := countStyle.Render(count)
			b.WriteString(nameStr + pad + countStr)
		}

		b.WriteString("\n\n")
	}

	// 残りの高さを埋める (各フォルダの下に空行1行)
	usedLines := folderListHeaderLines + len(fl.folders)*folderListItemHeight
	for i := usedLines; i < fl.height; i++ {
		b.WriteString("\n")
	}

	style := folderListStyle

	if focused {
		style = style.BorderForeground(lipgloss.Color("4"))
	}

	return style.Width(fl.width).Height(fl.height).Render(b.String())
}
