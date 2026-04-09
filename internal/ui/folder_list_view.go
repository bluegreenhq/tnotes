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
	for i, folder := range fl.folders {
		label := fmt.Sprintf(" %s (%d)", folder.Name, folder.Count)

		if i == fl.selected {
			style := lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15")).
				Bold(true).
				Width(contentWidth)
			b.WriteString(style.Render(label))
		} else {
			b.WriteString(lipgloss.NewStyle().Width(contentWidth).Render(label))
		}

		b.WriteString("\n")
	}

	// 残りの高さを埋める
	usedLines := folderListHeaderLines + len(fl.folders)
	for i := usedLines; i < fl.height; i++ {
		b.WriteString("\n")
	}

	style := folderListStyle

	if focused {
		style = style.BorderForeground(lipgloss.Color("4"))
	}

	return style.Width(fl.width).Height(fl.height).Render(b.String())
}
