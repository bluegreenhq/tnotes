package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	buttonStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	buttonHoverStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Underline(true)
	buttonDisabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	dirtyMarkStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	errorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	infoStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

// View はフッター行を描画する。infoMsgがあればボタンの右にシアン色で表示する。
// errMsgがある場合はエラーが優先される。
// 戻り値は描画文字列と行数。
func (f *Footer) View(errMsg, infoMsg string, width int) (string, int) {
	if errMsg != "" {
		return renderErrorLines(errMsg, width), errorLineCount(errMsg, width)
	}

	var buf strings.Builder
	buf.WriteString(" ")

	for i, btn := range f.buttons {
		if i > 0 {
			buf.WriteString("  ")
		}

		switch {
		case btn.Target == HoverNone && btn.Disabled:
			buf.WriteString(dirtyMarkStyle.Render(btn.Label))
		case btn.Disabled:
			buf.WriteString(renderDisabledButton(btn.Label))
		default:
			buf.WriteString(renderButton(btn.Label, f.hover == btn.Target))
		}
	}

	if infoMsg != "" {
		buf.WriteString("  ")
		buf.WriteString(infoStyle.Render(infoMsg))
	}

	return buf.String(), 1
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

func renderButton(label string, hovered bool) string {
	if hovered {
		return buttonHoverStyle.Render(label)
	}

	return buttonStyle.Render(label)
}

func renderDisabledButton(label string) string {
	return buttonDisabledStyle.Render(label)
}
