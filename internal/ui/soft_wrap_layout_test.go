package ui //nolint:testpackage // 内部フィールドへの直接アクセスが必要

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSoftWrapLayout_NoWrapNeeded(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("Hello"),
		[]rune("World"),
	}
	l.rebuild(lines, 10)

	assert.Equal(t, 2, l.totalVisualLines())

	vl := l.visualLinesFor(0)
	assert.Len(t, vl, 1)
	assert.Equal(t, 0, vl[0].startRune)
	assert.Equal(t, 5, vl[0].length)
}

func TestSoftWrapLayout_BasicWrap(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // width=5 → "abcde" + "fgh"
	}
	l.rebuild(lines, 5)

	assert.Equal(t, 2, l.totalVisualLines())

	vl := l.visualLinesFor(0)
	assert.Len(t, vl, 2)
	assert.Equal(t, 0, vl[0].startRune)
	assert.Equal(t, 5, vl[0].length)
	assert.Equal(t, 5, vl[1].startRune)
	assert.Equal(t, 3, vl[1].length)
}

func TestSoftWrapLayout_ExactWidth(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcde"), // width=5 → ぴったり1行
	}
	l.rebuild(lines, 5)

	assert.Equal(t, 1, l.totalVisualLines())
}

func TestSoftWrapLayout_EmptyLine(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("Hello"),
		{},
		[]rune("World"),
	}
	l.rebuild(lines, 10)

	assert.Equal(t, 3, l.totalVisualLines())

	vl := l.visualLinesFor(1)
	assert.Len(t, vl, 1)
	assert.Equal(t, 0, vl[0].length)
}

func TestSoftWrapLayout_FullWidthChars(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	// 全角文字は幅2。width=5 → "あい"(4) + "うえ"(4) + "お"(2)
	lines := [][]rune{
		[]rune("あいうえお"),
	}
	l.rebuild(lines, 5)

	assert.Equal(t, 3, l.totalVisualLines())

	vl := l.visualLinesFor(0)
	assert.Len(t, vl, 3)
	assert.Equal(t, 0, vl[0].startRune)
	assert.Equal(t, 2, vl[0].length) // "あい"
	assert.Equal(t, 2, vl[1].startRune)
	assert.Equal(t, 2, vl[1].length) // "うえ"
	assert.Equal(t, 4, vl[2].startRune)
	assert.Equal(t, 1, vl[2].length) // "お"
}

func TestSoftWrapLayout_LogicalToVisual(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → "abcde" + "fgh"
		[]rune("xyz"),
	}
	l.rebuild(lines, 5)

	// 論理行0, col=0 → 視覚行0
	assert.Equal(t, 0, l.logicalToVisual(0, 0))
	// 論理行0, col=4 → 視覚行0
	assert.Equal(t, 0, l.logicalToVisual(0, 4))
	// 論理行0, col=5 → 視覚行1
	assert.Equal(t, 1, l.logicalToVisual(0, 5))
	// 論理行0, col=7 → 視覚行1
	assert.Equal(t, 1, l.logicalToVisual(0, 7))
	// 論理行1, col=0 → 視覚行2
	assert.Equal(t, 2, l.logicalToVisual(1, 0))
}

func TestSoftWrapLayout_VisualToLogical(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → "abcde" + "fgh"
		[]rune("xyz"),
	}
	l.rebuild(lines, 5)

	line, startRune := l.visualToLogical(0)
	assert.Equal(t, 0, line)
	assert.Equal(t, 0, startRune)

	line, startRune = l.visualToLogical(1)
	assert.Equal(t, 0, line)
	assert.Equal(t, 5, startRune)

	line, startRune = l.visualToLogical(2)
	assert.Equal(t, 1, line)
	assert.Equal(t, 0, startRune)
}

func TestSoftWrapLayout_MultipleWraps(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	// 15文字, width=5 → 3視覚行
	lines := [][]rune{
		[]rune("abcdefghijklmno"),
	}
	l.rebuild(lines, 5)

	assert.Equal(t, 3, l.totalVisualLines())

	vl := l.visualLinesFor(0)
	assert.Len(t, vl, 3)
	assert.Equal(t, 0, vl[0].startRune)
	assert.Equal(t, 5, vl[0].length)
	assert.Equal(t, 5, vl[1].startRune)
	assert.Equal(t, 5, vl[1].length)
	assert.Equal(t, 10, vl[2].startRune)
	assert.Equal(t, 5, vl[2].length)
}

func TestSoftWrapLayout_MultipleLines(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → 2視覚行
		[]rune("ij"),       // → 1視覚行
		[]rune("klmnopqr"), // → 2視覚行
	}
	l.rebuild(lines, 5)

	assert.Equal(t, 5, l.totalVisualLines())
	assert.Equal(t, 0, l.logicalToVisual(0, 0))
	assert.Equal(t, 2, l.logicalToVisual(1, 0))
	assert.Equal(t, 3, l.logicalToVisual(2, 0))
}

func TestSoftWrapLayout_AdjustScroll(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → 2 visual lines
		[]rune("xyz"),      // → 1 visual line
	}
	l.rebuild(lines, 5)

	// カーソルが視覚行2（論理行1, col=0）、height=2 → scrollY=1
	scrollY, scrollX := l.adjustScroll(1, 0, 0, 0, 5, 2)
	assert.Equal(t, 1, scrollY)
	assert.Equal(t, 0, scrollX, "scrollX should always be 0 for soft wrap")
}

func TestSoftWrapLayout_MoveCursorUp(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → "abcde" + "fgh"
	}
	l.rebuild(lines, 5)

	// col=6（視覚行1）から上へ → 視覚行0内に移動
	row, col, moved := l.moveCursorUp(0, 6)
	assert.True(t, moved)
	assert.Equal(t, 0, row)
	assert.Less(t, col, 5, "cursor should be in first visual line")

	// 視覚行0の先頭からさらに上 → 動かない
	_, _, moved = l.moveCursorUp(0, 0)
	assert.False(t, moved)
}

func TestSoftWrapLayout_MoveCursorDown(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → "abcde" + "fgh"
		[]rune("xyz"),
	}
	l.rebuild(lines, 5)

	// col=2（視覚行0）から下へ → 視覚行1（同じ論理行の折り返し部分）
	row, col, moved := l.moveCursorDown(0, 2)
	assert.True(t, moved)
	assert.Equal(t, 0, row)
	assert.Equal(t, 7, col) // startRune=5 + colInVisual=2 = 7

	// 最後の視覚行からさらに下 → 動かない
	_, _, moved = l.moveCursorDown(1, 0)
	assert.False(t, moved)
}

func TestSoftWrapLayout_RenderViewLine(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → "abcde" + "fgh"
	}
	l.rebuild(lines, 5)

	assert.Equal(t, "abcde", l.renderViewLine(0, 0, 5))
	assert.Equal(t, "fgh", l.renderViewLine(1, 0, 5))
	assert.Empty(t, l.renderViewLine(-1, 0, 5))
}

func TestSoftWrapLayout_ViewLineStartRune(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → "abcde" + "fgh"
		[]rune("xyz"),
	}
	l.rebuild(lines, 5)

	logLine, startOff := l.viewLineStartRune(0, 0)
	assert.Equal(t, 0, logLine)
	assert.Equal(t, 0, startOff)

	logLine, startOff = l.viewLineStartRune(1, 0)
	assert.Equal(t, 0, logLine)
	assert.Equal(t, 5, startOff)

	logLine, startOff = l.viewLineStartRune(2, 0)
	assert.Equal(t, 1, logLine)
	assert.Equal(t, 0, startOff)
}

func TestSoftWrapLayout_ViewCellToLogical(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	lines := [][]rune{
		[]rune("abcdefgh"), // → "abcde" + "fgh"
	}
	l.rebuild(lines, 5)

	logLine, col := l.viewCellToLogical(0, 2)
	assert.Equal(t, 0, logLine)
	assert.Equal(t, 2, col)

	// 視覚行1のセル1 → 論理行0, col=5+1=6
	logLine, col = l.viewCellToLogical(1, 1)
	assert.Equal(t, 0, logLine)
	assert.Equal(t, 6, col)
}

func TestSoftWrapLayout_ViewCellToLogical_FullWidth(t *testing.T) {
	t.Parallel()

	l := newSoftWrapLayout()
	// "あいうえお" → "あい"(4cells) + "うえ"(4cells) + "お"(2cells)
	lines := [][]rune{
		[]rune("あいうえお"),
	}
	l.rebuild(lines, 5)

	// 視覚行1, セル位置2 → "うえ"の"え" → 論理行0, rune=3
	logLine, col := l.viewCellToLogical(1, 2)
	assert.Equal(t, 0, logLine)
	assert.Equal(t, 3, col)
}
