package ui

import "github.com/mattn/go-runewidth"

// noWrapLayout は折り返しなしの1:1マッピングを提供する。
type noWrapLayout struct {
	lines [][]rune
}

func newNoWrapLayout() *noWrapLayout {
	return &noWrapLayout{
		lines: nil,
	}
}

func (l *noWrapLayout) rebuild(lines [][]rune, _ int) {
	l.lines = lines
}

func (l *noWrapLayout) totalVisualLines() int {
	return len(l.lines)
}

func (l *noWrapLayout) logicalToVisual(line, _ int) int {
	return line
}

func (l *noWrapLayout) visualToLogical(visualRow int) (int, int) {
	return visualRow, 0
}

func (l *noWrapLayout) visualLinesFor(line int) []visualLine {
	if line < 0 || line >= len(l.lines) {
		return nil
	}

	return []visualLine{
		{
			logicalLine: line,
			startRune:   0,
			length:      len(l.lines[line]),
		},
	}
}

func (l *noWrapLayout) adjustScroll(row, col, scrollY, scrollX, width, height int) (int, int) {
	if row < scrollY {
		scrollY = row
	}

	if height > 0 && row >= scrollY+height {
		scrollY = row - height + 1
	}

	if width <= 0 {
		return scrollY, scrollX
	}

	cellPos := l.cursorCellPos(row, col)
	if cellPos < scrollX {
		scrollX = cellPos
	}

	if cellPos >= scrollX+width {
		scrollX = cellPos - width + 1
	}

	return scrollY, scrollX
}

func (l *noWrapLayout) moveCursorUp(row, col int) (int, int, bool) {
	if row <= 0 {
		return row, col, false
	}

	row--

	if maxCol := len(l.lines[row]); col > maxCol {
		col = maxCol
	}

	return row, col, true
}

func (l *noWrapLayout) moveCursorDown(row, col int) (int, int, bool) {
	if row >= len(l.lines)-1 {
		return row, col, false
	}

	row++

	if maxCol := len(l.lines[row]); col > maxCol {
		col = maxCol
	}

	return row, col, true
}

func (l *noWrapLayout) renderViewLine(visualRow, scrollX, width int) string {
	if visualRow < 0 || visualRow >= len(l.lines) {
		return ""
	}

	line := l.lines[visualRow]

	return truncateLineWithScroll(line, scrollX, width)
}

func (l *noWrapLayout) viewLineStartRune(visualRow, scrollX int) (int, int) {
	line := visualRow
	if line < 0 || line >= len(l.lines) {
		return 0, 0
	}

	startRuneOff := cellToRuneIndex(l.lines[line], scrollX)

	return line, startRuneOff
}

func (l *noWrapLayout) viewCellToLogical(visualRow, cellCol int) (int, int) {
	line := visualRow
	if line < 0 {
		return 0, 0
	}

	if line >= len(l.lines) {
		line = len(l.lines) - 1
	}

	col := cellToRuneIndex(l.lines[line], cellCol)

	return line, col
}

func (l *noWrapLayout) cursorCellPos(row, col int) int {
	if row < 0 || row >= len(l.lines) {
		return 0
	}

	line := l.lines[row]
	pos := 0

	for i := 0; i < col && i < len(line); i++ {
		pos += runewidth.RuneWidth(line[i])
	}

	return pos
}
