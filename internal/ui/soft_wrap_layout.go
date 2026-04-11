package ui

import (
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/utils"
)

// softWrapLayout は幅で論理行を文字単位で分割する。
type softWrapLayout struct {
	visual     []visualLine // 全視覚行のフラットリスト
	lineStarts []int        // lineStarts[i] = 論理行 i の最初の視覚行インデックス
	lines      [][]rune     // rebuild 時に保持する論理行の参照
	width      int
}

func newSoftWrapLayout() *softWrapLayout {
	return &softWrapLayout{
		visual:     nil,
		lineStarts: nil,
		lines:      nil,
		width:      0,
	}
}

func (l *softWrapLayout) rebuild(lines [][]rune, width int) {
	l.width = width
	l.lines = lines
	l.visual = l.visual[:0]
	l.lineStarts = make([]int, len(lines))

	for i, line := range lines {
		l.lineStarts[i] = len(l.visual)
		l.splitLine(i, line)
	}
}

func (l *softWrapLayout) totalVisualLines() int {
	return len(l.visual)
}

func (l *softWrapLayout) logicalToVisual(line, col int) int {
	if line < 0 || line >= len(l.lineStarts) {
		return 0
	}

	start := l.lineStarts[line]
	end := l.lineEnd(line)

	for vi := start; vi < end; vi++ {
		vl := l.visual[vi]
		if col < vl.startRune+vl.length {
			return vi
		}
	}

	// col が行末以降の場合は最後の視覚行
	return end - 1
}

func (l *softWrapLayout) visualToLogical(visualRow int) (int, int) {
	if visualRow < 0 || visualRow >= len(l.visual) {
		return 0, 0
	}

	vl := l.visual[visualRow]

	return vl.logicalLine, vl.startRune
}

func (l *softWrapLayout) visualLinesFor(line int) []visualLine {
	if line < 0 || line >= len(l.lineStarts) {
		return nil
	}

	start := l.lineStarts[line]
	end := l.lineEnd(line)

	return l.visual[start:end]
}

func (l *softWrapLayout) splitLine(logicalIdx int, line []rune) {
	if l.width <= 0 || len(line) == 0 {
		l.visual = append(l.visual, visualLine{
			logicalLine: logicalIdx,
			startRune:   0,
			length:      len(line),
		})

		return
	}

	startRune := 0
	cellWidth := 0

	for i, r := range line {
		rw := runewidth.RuneWidth(r)

		if cellWidth+rw > l.width {
			l.visual = append(l.visual, visualLine{
				logicalLine: logicalIdx,
				startRune:   startRune,
				length:      i - startRune,
			})

			startRune = i
			cellWidth = rw
		} else {
			cellWidth += rw
		}
	}

	// 残りの部分
	l.visual = append(l.visual, visualLine{
		logicalLine: logicalIdx,
		startRune:   startRune,
		length:      len(line) - startRune,
	})
}

// lineEnd は論理行 line の最後の視覚行インデックス+1 を返す。
func (l *softWrapLayout) lineEnd(line int) int {
	if line+1 < len(l.lineStarts) {
		return l.lineStarts[line+1]
	}

	return len(l.visual)
}

func (l *softWrapLayout) adjustScroll(row, col, scrollY, _, _, height int) (int, int) {
	visualRow := l.logicalToVisual(row, col)

	if visualRow < scrollY {
		scrollY = visualRow
	}

	if height > 0 && visualRow >= scrollY+height {
		scrollY = visualRow - height + 1
	}

	return scrollY, 0 // scrollX は常に 0
}

func (l *softWrapLayout) moveCursorUp(row, col int) (int, int, bool) {
	visualRow := l.logicalToVisual(row, col)
	if visualRow <= 0 {
		return row, col, false
	}

	targetVisual := visualRow - 1
	logLine, startRune := l.visualToLogical(targetVisual)

	vl := l.getVisualLine(targetVisual)
	_, curStart := l.visualToLogical(visualRow)
	colInVisual := max(col-curStart, 0)

	newCol := startRune + colInVisual
	newCol = min(newCol, startRune+vl.length)
	newCol = min(newCol, len(l.lines[logLine]))

	return logLine, newCol, true
}

func (l *softWrapLayout) moveCursorDown(row, col int) (int, int, bool) {
	visualRow := l.logicalToVisual(row, col)
	if visualRow >= l.totalVisualLines()-1 {
		return row, col, false
	}

	targetVisual := visualRow + 1
	logLine, startRune := l.visualToLogical(targetVisual)

	vl := l.getVisualLine(targetVisual)
	_, curStart := l.visualToLogical(visualRow)
	colInVisual := max(col-curStart, 0)

	newCol := startRune + colInVisual
	newCol = min(newCol, startRune+vl.length)
	newCol = min(newCol, len(l.lines[logLine]))

	return logLine, newCol, true
}

func (l *softWrapLayout) renderViewLine(visualRow, _, _ int) string {
	if visualRow < 0 || visualRow >= len(l.visual) {
		return ""
	}

	logLine, startRune := l.visualToLogical(visualRow)
	if logLine >= len(l.lines) {
		return ""
	}

	line := l.lines[logLine]
	vl := l.getVisualLine(visualRow)
	end := min(startRune+vl.length, len(line))

	if startRune > len(line) {
		return ""
	}

	return string(line[startRune:end])
}

func (l *softWrapLayout) viewLineStartRune(visualRow, _ int) (int, int) {
	if visualRow < 0 || visualRow >= len(l.visual) {
		return 0, 0
	}

	return l.visualToLogical(visualRow)
}

func (l *softWrapLayout) viewCellToLogical(visualRow, cellCol int) (int, int) {
	if len(l.visual) == 0 {
		return 0, 0
	}

	if visualRow < 0 {
		visualRow = 0
	}

	if visualRow >= len(l.visual) {
		visualRow = len(l.visual) - 1
	}

	logLine, startRune := l.visualToLogical(visualRow)
	vl := l.getVisualLine(visualRow)
	end := min(startRune+vl.length, len(l.lines[logLine]))
	visibleRunes := l.lines[logLine][startRune:end]

	return logLine, startRune + utils.CellToRuneIndex(visibleRunes, cellCol)
}

// getVisualLine は指定された視覚行インデックスの visualLine を返す。
func (l *softWrapLayout) getVisualLine(visualRow int) visualLine {
	if visualRow >= 0 && visualRow < len(l.visual) {
		return l.visual[visualRow]
	}

	return visualLine{logicalLine: 0, startRune: 0, length: 0}
}
