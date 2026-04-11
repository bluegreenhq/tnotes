package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const ansiReset = "\x1b[m" // lipgloss Render が出力するリセットの検出用

var (
	buttonStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	buttonHoverStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4"))

	editorStyle = lipgloss.NewStyle().Padding(0, 1)

	editorBoldStyle      = lipgloss.NewStyle().Bold(true)
	editorCursorStyle    = lipgloss.NewStyle().Reverse(true)
	editorSelectionStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Foreground(lipgloss.Color("15"))

		// タイトル太字.
	editorBoldOn  = extractANSIOn(editorBoldStyle)
	editorBoldOff = extractANSIOff(editorBoldStyle)

	// カーソル: reverse のみをトグルし、他の属性を維持する。
	editorCursorOn  = extractANSIOn(editorCursorStyle)
	editorCursorOff = extractANSIOff(editorCursorStyle)

	// 選択: fg/bg を設定し、解除時は fg/bg のみリセットする。
	editorSelectionOn  = extractANSIOn(editorSelectionStyle)
	editorSelectionOff = extractANSIOff(editorSelectionStyle)
)

// extractANSIOn は lipgloss Style から開始 ANSI シーケンスを抽出する。
func extractANSIOn(s lipgloss.Style) string {
	rendered := s.Render("\x00")
	before, _, _ := strings.Cut(rendered, "\x00")

	return before
}

// extractANSIOff は lipgloss Style で設定された属性に対応する
// 個別リセットシーケンスを生成する。全属性リセット (\x1b[m) を避け、
// 周囲のスタイルを維持する。
func extractANSIOff(s lipgloss.Style) string {
	var b strings.Builder

	if s.GetBold() {
		b.WriteString("\x1b[22m")
	}

	if s.GetItalic() {
		b.WriteString("\x1b[23m")
	}

	if s.GetUnderline() {
		b.WriteString("\x1b[24m")
	}

	if s.GetReverse() {
		b.WriteString("\x1b[27m")
	}

	if s.GetStrikethrough() {
		b.WriteString("\x1b[29m")
	}

	if _, ok := s.GetForeground().(lipgloss.NoColor); !ok {
		b.WriteString("\x1b[39m")
	}

	if _, ok := s.GetBackground().(lipgloss.NoColor); !ok {
		b.WriteString("\x1b[49m")
	}

	return b.String()
}
