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
func (fl *FolderList) View(focused bool, hoverSeparator bool) string {
	contentWidth := max(fl.width-folderListBorderWidth, 0)

	var b strings.Builder

	// ヘッダー: 閉じるボタン + タイトル + ボタン
	fl.renderHeader(&b, contentWidth)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", contentWidth))
	b.WriteString("\n")

	// フォルダ一覧
	for i, folder := range fl.folders {
		fl.renderFolder(&b, folder, i == fl.selected, contentWidth)
		b.WriteString("\n\n")
	}

	// インライン入力
	if fl.inputMode {
		inputStr := " " + fl.inputValue + "█"
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(inputStr))
		b.WriteString("\n\n")
	}

	// 残りの高さを埋める
	usedLines := folderListHeaderLines + len(fl.folders)*folderListItemHeight
	if fl.inputMode {
		usedLines += folderListItemHeight
	}

	for i := usedLines; i < fl.height; i++ {
		b.WriteString("\n")
	}

	style := folderListStyle

	if focused {
		style = style.BorderForeground(lipgloss.Color("4"))
	}

	if hoverSeparator {
		style = style.BorderStyle(lipgloss.ThickBorder())
	}

	return style.Width(fl.width).Height(fl.height).Render(b.String())
}

var folderCountStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

func (fl *FolderList) renderFolder(b *strings.Builder, folder Folder, selected bool, contentWidth int) {
	name := " " + folder.Name
	count := fmt.Sprintf("%d ", folder.Count)
	padding := max(contentWidth-lipgloss.Width(name)-lipgloss.Width(count), 0)

	if selected {
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
		countStr := folderCountStyle.Render(count)
		b.WriteString(nameStr + pad + countStr)
	}
}

func (fl *FolderList) renderHeader(b *strings.Builder, contentWidth int) {
	closeBtn := " "
	if fl.hoverClose {
		closeBtn += buttonHoverStyle.Render("✕")
	} else {
		closeBtn += buttonStyle.Render("✕")
	}

	title := " Folders"

	addBtn := " "
	if fl.hoverAdd {
		addBtn += buttonHoverStyle.Render("+")
	} else {
		addBtn += buttonStyle.Render("+")
	}

	moreBtn := ""
	if fl.IsUserFolder() {
		moreBtn = " "
		if fl.hoverMore {
			moreBtn += buttonHoverStyle.Render("⋯")
		} else {
			moreBtn += buttonStyle.Render("⋯")
		}
	}

	headerLeft := closeBtn + title
	headerRight := addBtn + moreBtn + " "
	headerLeftWidth := lipgloss.Width(headerLeft)
	headerRightWidth := lipgloss.Width(headerRight)
	padding := max(contentWidth-headerLeftWidth-headerRightWidth, 0)

	headerStr := headerLeft + strings.Repeat(" ", padding) + headerRight
	b.WriteString(lipgloss.NewStyle().Bold(true).Width(contentWidth).Render(headerStr))
}
