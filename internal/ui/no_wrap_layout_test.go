package ui //nolint:testpackage // 内部フィールドへの直接アクセスが必要

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoWrapLayout_BasicMapping(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("Hello"),
		[]rune("World"),
		[]rune(""),
	}
	l.rebuild(lines, 80)

	assert.Equal(t, 3, l.totalVisualLines())
	assert.Equal(t, 0, l.logicalToVisual(0, 0))
	assert.Equal(t, 1, l.logicalToVisual(1, 0))
	assert.Equal(t, 2, l.logicalToVisual(2, 0))

	line, startRune := l.visualToLogical(0)
	assert.Equal(t, 0, line)
	assert.Equal(t, 0, startRune)

	line, startRune = l.visualToLogical(1)
	assert.Equal(t, 1, line)
	assert.Equal(t, 0, startRune)
}

func TestNoWrapLayout_VisualLinesFor(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("Hello"),
		[]rune("World"),
	}
	l.rebuild(lines, 80)

	vl := l.visualLinesFor(0)
	assert.Len(t, vl, 1)
	assert.Equal(t, 0, vl[0].logicalLine)
	assert.Equal(t, 0, vl[0].startRune)
	assert.Equal(t, 5, vl[0].length)

	assert.Nil(t, l.visualLinesFor(-1))
	assert.Nil(t, l.visualLinesFor(2))
}

func TestNoWrapLayout_AdjustScroll(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("line0"),
		[]rune("line1"),
		[]rune("line2"),
		[]rune("line3"),
		[]rune("line4"),
	}
	l.rebuild(lines, 80)

	// カーソルが画面外（下）→ scrollY 調整
	scrollY, scrollX := l.adjustScroll(4, 0, 0, 0, 80, 3)
	assert.Equal(t, 2, scrollY)
	assert.Equal(t, 0, scrollX)

	// カーソルが画面外（上）→ scrollY 調整
	scrollY, _ = l.adjustScroll(0, 0, 2, 0, 80, 3)
	assert.Equal(t, 0, scrollY)
}

func TestNoWrapLayout_AdjustScrollHorizontal(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("abcdefghij"),
	}
	l.rebuild(lines, 5)

	// col=8 → cellPos=8, width=5 → scrollX = 8-5+1 = 4
	_, scrollX := l.adjustScroll(0, 8, 0, 0, 5, 3)
	assert.Equal(t, 4, scrollX)
}

func TestNoWrapLayout_MoveCursorUp(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("Hi"),
		[]rune("Hello"),
	}
	l.rebuild(lines, 80)

	row, col, moved := l.moveCursorUp(1, 4)
	assert.True(t, moved)
	assert.Equal(t, 0, row)
	assert.Equal(t, 2, col, "col should be clamped to shorter line")

	_, _, moved = l.moveCursorUp(0, 0)
	assert.False(t, moved)
}

func TestNoWrapLayout_MoveCursorDown(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("Hello"),
		[]rune("Hi"),
	}
	l.rebuild(lines, 80)

	row, col, moved := l.moveCursorDown(0, 4)
	assert.True(t, moved)
	assert.Equal(t, 1, row)
	assert.Equal(t, 2, col, "col should be clamped to shorter line")

	_, _, moved = l.moveCursorDown(1, 0)
	assert.False(t, moved)
}

func TestNoWrapLayout_RenderViewLine(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("abcdefghij"),
	}
	l.rebuild(lines, 5)

	assert.Equal(t, "abcde", l.renderViewLine(0, 0, 5))
	assert.Equal(t, "defgh", l.renderViewLine(0, 3, 5))
	assert.Empty(t, l.renderViewLine(-1, 0, 5))
}

func TestNoWrapLayout_ViewLineStartRune(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("abcdefghij"),
		[]rune("xyz"),
	}
	l.rebuild(lines, 5)

	logLine, startOff := l.viewLineStartRune(0, 3)
	assert.Equal(t, 0, logLine)
	assert.Equal(t, 3, startOff) // cellToRuneIndex("abcdefghij", 3) = 3

	logLine, startOff = l.viewLineStartRune(1, 0)
	assert.Equal(t, 1, logLine)
	assert.Equal(t, 0, startOff)
}

func TestNoWrapLayout_ViewCellToLogical(t *testing.T) {
	t.Parallel()

	l := newNoWrapLayout()
	lines := [][]rune{
		[]rune("Hello"),
		[]rune("World"),
	}
	l.rebuild(lines, 80)

	logLine, col := l.viewCellToLogical(0, 3)
	assert.Equal(t, 0, logLine)
	assert.Equal(t, 3, col)

	logLine, col = l.viewCellToLogical(1, 10)
	assert.Equal(t, 1, logLine)
	assert.Equal(t, 5, col, "should clamp to line length")
}
