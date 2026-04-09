package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// simpleTextArea は独自テキストエリア。
// カーソルとスクロールオフセットを完全独立管理する。
type simpleTextArea struct {
	lines   [][]rune
	row     int
	col     int
	scrollY int
	scrollX int
	width   int
	height  int
	focused bool
	layout  lineLayout
}

func newSimpleTextArea(noWrap bool) simpleTextArea {
	var lo lineLayout
	if noWrap {
		lo = newNoWrapLayout()
	} else {
		lo = newSoftWrapLayout()
	}

	return simpleTextArea{
		lines:   [][]rune{{}},
		row:     0,
		col:     0,
		scrollY: 0,
		scrollX: 0,
		width:   0,
		height:  0,
		focused: false,
		layout:  lo,
	}
}

// Value は全テキストを返す。
func (t *simpleTextArea) Value() string {
	parts := make([]string, len(t.lines))
	for i, line := range t.lines {
		parts[i] = string(line)
	}

	return strings.Join(parts, "\n")
}

// Line はカーソルの行番号を返す。
func (t *simpleTextArea) Line() int { return t.row }

// Column はカーソルの列番号（ルーンインデックス）を返す。
func (t *simpleTextArea) Column() int { return t.col }

// LineCount は行数を返す。
func (t *simpleTextArea) LineCount() int { return len(t.lines) }

// ScrollYOffset は表示先頭行を返す。
func (t *simpleTextArea) ScrollYOffset() int { return t.scrollY }

// ScrollXOffset は表示先頭列（セル単位）を返す。
func (t *simpleTextArea) ScrollXOffset() int { return t.scrollX }

// SetWidth は表示幅を設定する。
func (t *simpleTextArea) SetWidth(w int) {
	t.width = w
	t.layout.rebuild(t.lines, t.width)
}

// SetHeight は表示高さを設定する。
func (t *simpleTextArea) SetHeight(h int) { t.height = h }

// Focus はフォーカスを設定する。
func (t *simpleTextArea) Focus() tea.Cmd {
	t.focused = true

	return nil
}

// Blur はフォーカスを解除する。
func (t *simpleTextArea) Blur() { t.focused = false }

// Focused はフォーカス状態を返す。
func (t *simpleTextArea) Focused() bool { return t.focused }

func (t *simpleTextArea) ensureVisible() {
	t.scrollY, t.scrollX = t.layout.adjustScroll(t.row, t.col, t.scrollY, t.scrollX, t.width, t.height)
}
