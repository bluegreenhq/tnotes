package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/bluegreenhq/tnotes/internal/utils"
)

const (
	ansiBoldOn  = "\x1b[1m"
	ansiBoldOff = "\x1b[m"
	ansiReset   = "\x1b[m"
)

var (
	editorStyle       = lipgloss.NewStyle().Padding(0, 1)
	editorCursorStyle = lipgloss.NewStyle().Reverse(true)
)

// View はエディタの描画内容を返す。
func (e *Editor) View() string {
	if e.noteID == "" {
		placeholder := "Press 'n' to create a note"
		if e.readOnly {
			placeholder = ""
		}

		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Width(e.width).
			Height(e.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(placeholder)
	}

	raw := e.textarea.View()

	if e.HasSelection() {
		raw = e.applySelectionHighlight(raw)
	} else if e.textarea.Focused() && e.blinkVisible {
		raw = e.applyCursor(raw)
	}

	raw = e.applyTitleBold(raw)

	return editorStyle.Width(e.width).Height(e.height).Render(raw)
}

// applyTitleBold は先頭のタイトル行を太字にする。
// カーソルや選択のANSIエスケープが含まれていても正しく動作する。
func (e *Editor) applyTitleBold(raw string) string {
	scrollOffset := e.textarea.ScrollYOffset()

	// タイトル行（論理行0）が表示する視覚行数を取得
	titleVisualLines := len(e.textarea.layout.visualLinesFor(0))
	if titleVisualLines == 0 {
		titleVisualLines = 1
	}

	// スクロールでタイトル行が完全に画面外なら何もしない
	if scrollOffset >= titleVisualLines {
		return raw
	}

	viewLines := strings.Split(raw, "\n")
	if len(viewLines) == 0 {
		return raw
	}

	// 画面に表示されているタイトルの視覚行数
	boldCount := min(titleVisualLines-scrollOffset, len(viewLines))

	for i := range boldCount {
		line := viewLines[i]
		// 内部のリセットシーケンス後に太字を再適用する
		line = strings.ReplaceAll(line, ansiReset, ansiReset+ansiBoldOn)
		viewLines[i] = ansiBoldOn + line + ansiBoldOff
	}

	return strings.Join(viewLines, "\n")
}

// applyCursor はカーソル位置の文字を反転表示する。
func (e *Editor) applyCursor(raw string) string {
	visualRow := e.textarea.layout.logicalToVisual(e.textarea.Line(), e.textarea.Column())
	cursorViewRow := visualRow - e.textarea.ScrollYOffset()

	_, startRuneOff := e.textarea.layout.viewLineStartRune(visualRow, e.textarea.scrollX)
	cursorCol := e.textarea.Column() - startRuneOff

	viewLines := strings.Split(raw, "\n")
	if cursorViewRow < 0 || cursorViewRow >= len(viewLines) {
		return raw
	}

	line := viewLines[cursorViewRow]
	runes := []rune(line)

	if cursorCol < 0 || cursorCol > len(runes) {
		return raw
	}

	if cursorCol < len(runes) {
		before := string(runes[:cursorCol])
		cursor := editorCursorStyle.Render(string(runes[cursorCol]))
		after := string(runes[cursorCol+1:])
		viewLines[cursorViewRow] = before + cursor + after
	} else {
		viewLines[cursorViewRow] = line + editorCursorStyle.Render(" ")
	}

	return strings.Join(viewLines, "\n")
}

func (e *Editor) applySelectionHighlight(raw string) string {
	selectionStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15"))

	if !e.HasSelection() {
		return raw
	}

	start, end := e.NormalizedSelection()
	scrollOffset := e.textarea.ScrollYOffset()

	viewLines := strings.Split(raw, "\n")

	for i, line := range viewLines {
		visualRow := i + scrollOffset
		logLine, startRuneOff := e.textarea.layout.viewLineStartRune(visualRow, e.textarea.scrollX)

		if logLine < start.Line || logLine > end.Line {
			continue
		}

		runes := []rune(line)

		var colStart, colEnd int
		if logLine == start.Line {
			colStart = start.Column - startRuneOff
		}

		if logLine == end.Line {
			colEnd = end.Column - startRuneOff
		} else {
			colEnd = len(runes)
		}

		colStart = utils.ClampInt(colStart, 0, len(runes))
		colEnd = utils.ClampInt(colEnd, 0, len(runes))

		if colStart >= colEnd {
			continue
		}

		before := string(runes[:colStart])
		middle := string(runes[colStart:colEnd])
		after := string(runes[colEnd:])

		viewLines[i] = before + selectionStyle.Render(middle) + after
	}

	return strings.Join(viewLines, "\n")
}
