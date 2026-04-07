package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// simpleTextArea は折り返しなしの独自テキストエリア。
// カーソルとスクロールオフセットを完全独立管理する。
type simpleTextArea struct {
	lines   [][]rune
	row     int
	col     int
	scrollY int
	width   int
	height  int
	focused bool
}

func newSimpleTextArea() simpleTextArea {
	return simpleTextArea{
		lines:   [][]rune{{}},
		row:     0,
		col:     0,
		scrollY: 0,
		width:   0,
		height:  0,
		focused: false,
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

// SetWidth は表示幅を設定する。
func (t *simpleTextArea) SetWidth(w int) { t.width = w }

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
	if t.row < t.scrollY {
		t.scrollY = t.row
	}

	if t.height > 0 && t.row >= t.scrollY+t.height {
		t.scrollY = t.row - t.height + 1
	}
}
