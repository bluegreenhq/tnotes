package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/bluegreenhq/tnotes/internal/utils"
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
	} else if e.textarea.Focused() {
		raw = e.applyCursor(raw)
	}

	return editorStyle.Width(e.width).Height(e.height).Render(raw)
}

// applyCursor はカーソル位置の文字を反転表示する。
func (e *Editor) applyCursor(raw string) string {
	cursorViewRow := e.textarea.Line() - e.textarea.ScrollYOffset()
	logicalLine := e.textarea.Line()
	scrollXRuneOff := cellToRuneIndex(e.textarea.lines[logicalLine], e.textarea.scrollX)
	cursorCol := e.textarea.Column() - scrollXRuneOff

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
		logicalLine := i + scrollOffset

		if logicalLine < start.Line || logicalLine > end.Line {
			continue
		}

		runes := []rune(line)

		scrollXRuneOff := 0
		if logicalLine >= 0 && logicalLine < len(e.textarea.lines) {
			scrollXRuneOff = cellToRuneIndex(e.textarea.lines[logicalLine], e.textarea.scrollX)
		}

		var colStart, colEnd int
		if logicalLine == start.Line {
			colStart = start.Column - scrollXRuneOff
		}

		if logicalLine == end.Line {
			colEnd = end.Column - scrollXRuneOff
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
