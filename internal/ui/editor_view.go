package ui

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"github.com/bluegreenhq/tnotes/internal/utils"
)

// View はエディタの描画内容を返す。
func (e *Editor) View() string {
	headerLine := e.Header.View()

	if e.noteID == "" {
		placeholder := "Press 'n' to create a note"
		if e.readOnly {
			placeholder = ""
		}

		body := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Width(e.width).
			Height(e.height-editorHeaderHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render(placeholder)

		return headerLine + "\n" + body
	}

	raw := e.textarea.View()

	raw = e.applyTitleBold(raw)
	raw = e.applyURLStyle(raw)

	if e.HasSelection() {
		raw = e.applySelectionHighlight(raw)
	} else if e.textarea.Focused() && e.blinkVisible {
		raw = e.applyCursor(raw)
	}

	textBody := editorStyle.Width(e.width).Height(e.height - editorHeaderHeight).Render(raw)

	return headerLine + "\n" + textBody
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
		line = strings.ReplaceAll(line, ansiReset, ansiReset+editorBoldOn)
		viewLines[i] = editorBoldOn + line + editorBoldOff
	}

	return strings.Join(viewLines, "\n")
}

var (
	urlPattern   = regexp.MustCompile(`https?://[^\s]+`)
	urlGrayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// applyURLStyle はテキスト中の URL をグレー文字色にし、OSC 8 ハイパーリンクを付与する。
// 論理行の元テキストから URL 位置を検出し、視覚行上の対応範囲にスタイルを適用する。
func (e *Editor) applyURLStyle(raw string) string {
	scrollOffset := e.textarea.ScrollYOffset()
	viewLines := strings.Split(raw, "\n")

	for i, line := range viewLines {
		visualRow := i + scrollOffset
		logLine, startRuneOff := e.textarea.layout.viewLineStartRune(visualRow, e.textarea.scrollX)

		if logLine >= len(e.textarea.lines) {
			continue
		}

		logicalText := string(e.textarea.lines[logLine])
		locs := urlPattern.FindAllStringIndex(logicalText, -1)

		if len(locs) == 0 {
			continue
		}

		visLen := findVisualLineLength(e.textarea.layout.visualLinesFor(logLine), startRuneOff)
		styled := styleURLsInLine([]rune(line), logicalText, locs, startRuneOff, visLen)

		if styled != "" {
			viewLines[i] = styled
		}
	}

	return strings.Join(viewLines, "\n")
}

func findVisualLineLength(vLines []visualLine, startRuneOff int) int {
	for _, v := range vLines {
		if v.startRune == startRuneOff {
			return v.length
		}
	}

	return 0
}

func styleURLsInLine(runes []rune, logicalText string, locs [][]int, startRuneOff, visLen int) string {
	lineLen := len(runes)

	var result strings.Builder

	pos := 0

	for _, loc := range locs {
		urlStartRune := utf8.RuneCountInString(logicalText[:loc[0]])
		urlEndRune := utf8.RuneCountInString(logicalText[:loc[1]])
		urlText := logicalText[loc[0]:loc[1]]

		colStart := max(urlStartRune-startRuneOff, 0)
		colEnd := min(urlEndRune-startRuneOff, visLen)
		colEnd = min(colEnd, lineLen)

		if colStart >= colEnd {
			continue
		}

		styled := urlGrayStyle.Hyperlink(urlText).Render(string(runes[colStart:colEnd]))
		result.WriteString(string(runes[pos:colStart]))
		result.WriteString(styled)

		pos = colEnd
	}

	if pos > 0 {
		result.WriteString(string(runes[pos:]))

		return result.String()
	}

	return ""
}

// applyCursor はカーソル位置の文字を反転表示する。
// ANSI エスケープシーケンスをスキップして可視ルーン位置を計算する。
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

	byteStart, byteEnd := visibleRuneByteRange(line, cursorCol)
	if byteStart < 0 {
		// カーソルが行末の場合
		viewLines[cursorViewRow] = line + editorCursorOn + " " + editorCursorOff
	} else {
		before := line[:byteStart]
		cursor := editorCursorOn + line[byteStart:byteEnd] + editorCursorOff
		after := line[byteEnd:]
		viewLines[cursorViewRow] = before + cursor + after
	}

	return strings.Join(viewLines, "\n")
}

func (e *Editor) applySelectionHighlight(raw string) string {
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

		visibleCount := countVisibleRunes(line)

		var colStart, colEnd int
		if logLine == start.Line {
			colStart = start.Column - startRuneOff
		}

		if logLine == end.Line {
			colEnd = end.Column - startRuneOff
		} else {
			colEnd = visibleCount
		}

		colStart = utils.ClampInt(colStart, 0, visibleCount)
		colEnd = utils.ClampInt(colEnd, 0, visibleCount)

		if colStart >= colEnd {
			continue
		}

		byteStart, _ := visibleRuneByteRange(line, colStart)
		_, byteEnd := visibleRuneByteRange(line, colEnd-1)

		if byteStart < 0 || byteEnd < 0 {
			continue
		}

		before := line[:byteStart]
		middle := line[byteStart:byteEnd]
		after := line[byteEnd:]

		restore := collectANSIState(line, byteEnd)
		viewLines[i] = before + editorSelectionOn + middle + editorSelectionOff + restore + after
	}

	return strings.Join(viewLines, "\n")
}
