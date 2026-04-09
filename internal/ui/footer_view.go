package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	dirtyMarkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	infoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

const footerLineCount = 3

// View はフッター行を描画する。infoMsgがあればボタンの右にシアン色で表示する。
// errMsgがある場合はエラーが優先される。
// 戻り値は描画文字列と行数。
func (f *Footer) View(errMsg, infoMsg string, width int) (string, int) {
	if errMsg != "" {
		f.menuOpen = false

		return renderErrorLines(errMsg, width), errorLineCount(errMsg, width)
	}

	btns, dirtyMark := f.collectBoxButtons()
	topLine := renderFooterTopLine(btns)
	midLine := renderFooterMidLine(btns, dirtyMark, infoMsg)
	botLine := renderFooterBotLine(btns)

	return topLine + "\n" + midLine + "\n" + botLine, footerLineCount
}

func (f *Footer) collectBoxButtons() ([]BoxButton, string) {
	var btns []BoxButton

	var dirtyMark string

	for _, btn := range f.buttons {
		if btn.Target == HoverNone && btn.Disabled {
			dirtyMark = btn.Label

			continue
		}

		bb := NewBoxButton(btn.Label)
		bb.SetHovered(f.hover == btn.Target)
		btns = append(btns, bb)
	}

	return btns, dirtyMark
}

func renderFooterTopLine(btns []BoxButton) string {
	var buf strings.Builder

	buf.WriteString(" ")

	for i := range btns {
		if i > 0 {
			buf.WriteString("  ")
		}

		buf.WriteString(btns[i].ViewTop())
	}

	return buf.String()
}

func renderFooterMidLine(btns []BoxButton, dirtyMark, infoMsg string) string {
	var buf strings.Builder

	buf.WriteString(" ")

	for i := range btns {
		if i > 0 {
			buf.WriteString("  ")
		}

		buf.WriteString(btns[i].ViewMiddle())
	}

	if dirtyMark != "" {
		buf.WriteString("  ")
		buf.WriteString(dirtyMarkStyle.Render(dirtyMark))
	}

	if infoMsg != "" {
		buf.WriteString("  ")
		buf.WriteString(infoStyle.Render(infoMsg))
	}

	return buf.String()
}

func renderFooterBotLine(btns []BoxButton) string {
	var buf strings.Builder

	buf.WriteString(" ")

	for i := range btns {
		if i > 0 {
			buf.WriteString("  ")
		}

		buf.WriteString(btns[i].ViewBottom())
	}

	return buf.String()
}

func errorLineCount(msg string, width int) int {
	contentWidth := width - 1 // 先頭スペース分
	if contentWidth <= 0 {
		return 1
	}

	runes := []rune(msg)
	lines := (len(runes) + contentWidth - 1) / contentWidth
	lines = max(lines, 1)

	return lines
}

func renderErrorLines(msg string, width int) string {
	contentWidth := width - 1
	if contentWidth <= 0 || len([]rune(msg)) <= contentWidth {
		return " " + errorStyle.Render(msg)
	}

	runes := []rune(msg)

	var buf strings.Builder

	for i := 0; i < len(runes); i += contentWidth {
		end := min(i+contentWidth, len(runes))
		line := " " + errorStyle.Render(string(runes[i:end]))

		if i > 0 {
			buf.WriteString("\n")
		}

		buf.WriteString(line)
	}

	return buf.String()
}
