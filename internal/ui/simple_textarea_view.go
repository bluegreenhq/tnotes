package ui

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// View はテキストエリアの描画内容をプレーンテキストで返す。
// ANSIエスケープは含まない。カーソル表示は呼び出し側の責務。
func (t *simpleTextArea) View() string {
	var b strings.Builder

	totalVisual := t.layout.totalVisualLines()
	endVisual := min(t.scrollY+t.height, totalVisual)

	for vi := t.scrollY; vi < endVisual; vi++ {
		if vi > t.scrollY {
			b.WriteString("\n")
		}

		b.WriteString(t.layout.renderViewLine(vi, t.scrollX, t.width))
	}

	// 行数が height に満たない場合は空行で埋める
	for i := endVisual - t.scrollY; i < t.height; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

// truncateLineWithScroll は水平スクロール位置から幅分のテキストを返す。
func truncateLineWithScroll(line []rune, scrollX, width int) string {
	if width <= 0 {
		return string(line)
	}

	startRune := cellToRuneIndex(line, scrollX)
	remaining := line[startRune:]

	cellWidth := 0

	for i, r := range remaining {
		rw := runewidth.RuneWidth(r)
		if cellWidth+rw > width {
			return string(remaining[:i])
		}

		cellWidth += rw
	}

	return string(remaining)
}
