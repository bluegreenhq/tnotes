package ui

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// View はテキストエリアの描画内容をプレーンテキストで返す。
// ANSIエスケープは含まない。カーソル表示は呼び出し側の責務。
func (t *simpleTextArea) View() string {
	var b strings.Builder

	endLine := min(t.scrollY+t.height, len(t.lines))

	for i := t.scrollY; i < endLine; i++ {
		if i > t.scrollY {
			b.WriteString("\n")
		}

		line := t.lines[i]
		b.WriteString(t.truncateLine(line))
	}

	// 行数が height に満たない場合は空行で埋める
	for i := endLine - t.scrollY; i < t.height; i++ {
		b.WriteString("\n")
	}

	return b.String()
}

func (t *simpleTextArea) truncateLine(line []rune) string {
	if t.width <= 0 {
		return string(line)
	}

	startRune := cellToRuneIndex(line, t.scrollX)
	remaining := line[startRune:]

	cellWidth := 0

	for i, r := range remaining {
		rw := runewidth.RuneWidth(r)
		if cellWidth+rw > t.width {
			return string(remaining[:i])
		}

		cellWidth += rw
	}

	return string(remaining)
}
