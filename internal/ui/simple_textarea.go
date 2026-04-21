package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/utils"
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
	killBuf []rune // Ctrl+K で削除した内容を保持（Ctrl+Y でペースト）
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
		killBuf: nil,
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

// SetValue はテキストを設定し、カーソルを末尾に移動する。
func (t *simpleTextArea) SetValue(s string) {
	raw := strings.Split(s, "\n")
	t.lines = make([][]rune, len(raw))

	for i, l := range raw {
		t.lines[i] = []rune(l)
	}

	if len(t.lines) == 0 {
		t.lines = [][]rune{{}}
	}

	t.layout.rebuild(t.lines, t.width)
	t.row = len(t.lines) - 1
	t.col = len(t.lines[t.row])
	t.ensureVisible()
}

// SetCursorColumn はカーソル列を設定する。
func (t *simpleTextArea) SetCursorColumn(col int) {
	maxCol := len(t.lines[t.row])

	if col < 0 {
		col = 0
	}

	if col > maxCol {
		col = maxCol
	}

	t.col = col
	t.ensureVisible()
}

// InsertText はカーソル位置にテキストを挿入する。改行を含むテキストにも対応する。
func (t *simpleTextArea) InsertText(s string) {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if i > 0 {
			t.insertNewline()
		}

		t.insertText(l)
	}

	t.ensureVisible()
}

// MoveToBegin はカーソルをテキスト先頭に移動する。
func (t *simpleTextArea) MoveToBegin() {
	t.row = 0
	t.col = 0
	t.ensureVisible()
}

// Update はキー入力に応じてテキストを編集する。
// ポインタレシーバのため再代入不要。
func (t *simpleTextArea) Update(msg tea.Msg) tea.Cmd {
	if !t.focused {
		return nil
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}

	return t.handleKey(keyMsg)
}

// ScrollUp はカーソルを動かさずに表示を n 行上にスクロールする。
func (t *simpleTextArea) ScrollUp(n int) {
	t.scrollY = max(t.scrollY-n, 0)
}

// ScrollDown はカーソルを動かさずに表示を n 行下にスクロールする。
func (t *simpleTextArea) ScrollDown(n int) {
	maxScroll := max(t.layout.totalVisualLines()-t.height, 0)
	t.scrollY = min(t.scrollY+n, maxScroll)
}

// CursorUp はカーソルを1つ上の視覚行に移動する。
func (t *simpleTextArea) CursorUp() {
	newRow, newCol, moved := t.layout.moveCursorUp(t.row, t.col)
	if !moved {
		return
	}

	t.row = newRow
	t.col = newCol
	t.ensureVisible()
}

// CursorDown はカーソルを1つ下の視覚行に移動する。
func (t *simpleTextArea) CursorDown() {
	newRow, newCol, moved := t.layout.moveCursorDown(t.row, t.col)
	if !moved {
		return
	}

	t.row = newRow
	t.col = newCol
	t.ensureVisible()
}

// MoveTo はカーソルを指定の論理行・列に移動する。
func (t *simpleTextArea) MoveTo(line, col int) {
	line = max(line, 0)
	line = min(line, len(t.lines)-1)
	t.row = line

	col = max(col, 0)
	col = min(col, len(t.lines[t.row]))
	t.col = col

	t.ensureVisible()
}

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

func (t *simpleTextArea) ensureVisible() {
	t.scrollY, t.scrollX = t.layout.adjustScroll(t.row, t.col, t.scrollY, t.scrollX, t.width, t.height)
}

func (t *simpleTextArea) handleKey(msg tea.KeyPressMsg) tea.Cmd { //nolint:cyclop,gocyclo,funlen // キーバインド分岐
	switch {
	case msg.Code == 'a' && msg.Mod == tea.ModCtrl:
		t.col = 0
	case msg.Code == 'e' && msg.Mod == tea.ModCtrl:
		t.col = len(t.lines[t.row])
	case msg.Code == 'f' && msg.Mod == tea.ModCtrl:
		t.cursorRight()
	case msg.Code == 'b' && msg.Mod == tea.ModCtrl:
		t.cursorLeft()
	case msg.Code == 'n' && msg.Mod == tea.ModCtrl:
		t.CursorDown()
	case msg.Code == 'p' && msg.Mod == tea.ModCtrl:
		t.CursorUp()
	case msg.Code == 'd' && msg.Mod == tea.ModCtrl:
		t.delete()
	case msg.Code == 'k' && msg.Mod == tea.ModCtrl:
		t.killLine()
	case msg.Code == 'y' && msg.Mod == tea.ModCtrl:
		t.yank()
	case msg.Text != "" && (msg.Mod == 0 || msg.Mod == tea.ModShift):
		t.insertText(msg.Text)
	case msg.Code == tea.KeyEnter:
		t.insertNewline()
	case msg.Code == tea.KeyBackspace, msg.Code == 'h' && msg.Mod == tea.ModCtrl:
		t.backspace()
	case msg.Code == tea.KeyDelete:
		t.delete()
	case msg.Code == tea.KeyLeft:
		t.cursorLeft()
	case msg.Code == tea.KeyRight:
		t.cursorRight()
	case msg.Code == tea.KeyUp:
		t.CursorUp()
	case msg.Code == tea.KeyDown:
		t.CursorDown()
	case msg.Code == tea.KeyHome:
		t.col = 0
	case msg.Code == tea.KeyEnd:
		t.col = len(t.lines[t.row])
	}

	t.ensureVisible()

	return nil
}

func (t *simpleTextArea) insertText(s string) {
	runes := []rune(s)
	line := t.lines[t.row]
	newLine := make([]rune, 0, len(line)+len(runes))
	newLine = append(newLine, line[:t.col]...)
	newLine = append(newLine, runes...)
	newLine = append(newLine, line[t.col:]...)
	t.lines[t.row] = newLine
	t.col += len(runes)
	t.layout.rebuild(t.lines, t.width)
}

func (t *simpleTextArea) insertNewline() {
	line := t.lines[t.row]
	before := make([]rune, t.col)
	copy(before, line[:t.col])

	after := make([]rune, len(line)-t.col)
	copy(after, line[t.col:])

	newLines := make([][]rune, 0, len(t.lines)+1)
	newLines = append(newLines, t.lines[:t.row]...)
	newLines = append(newLines, before, after)
	newLines = append(newLines, t.lines[t.row+1:]...)
	t.lines = newLines
	t.row++
	t.col = 0
	t.layout.rebuild(t.lines, t.width)
	t.ensureVisible()
}

func (t *simpleTextArea) backspace() {
	if t.col > 0 {
		line := t.lines[t.row]
		t.lines[t.row] = append(line[:t.col-1], line[t.col:]...)
		t.col--
		t.layout.rebuild(t.lines, t.width)
	} else if t.row > 0 {
		prevLen := len(t.lines[t.row-1])
		t.lines[t.row-1] = append(t.lines[t.row-1], t.lines[t.row]...)
		t.lines = append(t.lines[:t.row], t.lines[t.row+1:]...)
		t.row--
		t.col = prevLen
		t.layout.rebuild(t.lines, t.width)
		t.ensureVisible()
	}
}

func (t *simpleTextArea) delete() {
	line := t.lines[t.row]
	if t.col < len(line) {
		t.lines[t.row] = append(line[:t.col], line[t.col+1:]...)
		t.layout.rebuild(t.lines, t.width)
	} else if t.row < len(t.lines)-1 {
		t.lines[t.row] = append(t.lines[t.row], t.lines[t.row+1]...)
		t.lines = append(t.lines[:t.row+1], t.lines[t.row+2:]...)
		t.layout.rebuild(t.lines, t.width)
	}
}

func (t *simpleTextArea) killLine() {
	line := t.lines[t.row]
	if t.col < len(line) {
		killed := make([]rune, len(line)-t.col)
		copy(killed, line[t.col:])
		t.killBuf = killed
		t.lines[t.row] = line[:t.col]
		t.layout.rebuild(t.lines, t.width)
	} else if t.row < len(t.lines)-1 {
		t.killBuf = []rune{'\n'}
		t.lines[t.row] = append(t.lines[t.row], t.lines[t.row+1]...)
		t.lines = append(t.lines[:t.row+1], t.lines[t.row+2:]...)
		t.layout.rebuild(t.lines, t.width)
	}
}

func (t *simpleTextArea) yank() {
	if len(t.killBuf) == 0 {
		return
	}

	for _, r := range t.killBuf {
		if r == '\n' {
			t.insertNewline()
		} else {
			t.insertText(string(r))
		}
	}
}

func (t *simpleTextArea) cursorLeft() {
	if t.col > 0 {
		t.col--
	} else if t.row > 0 {
		t.row--
		t.col = len(t.lines[t.row])
		t.ensureVisible()
	}
}

func (t *simpleTextArea) cursorRight() {
	if t.col < len(t.lines[t.row]) {
		t.col++
	} else if t.row < len(t.lines)-1 {
		t.row++
		t.col = 0
		t.ensureVisible()
	}
}

// positionFromCell は視覚行・セル列から論理行・ルーン列を返す。
// 視覚行が範囲外の場合はクランプする。
func (t *simpleTextArea) positionFromCell(visualRow, cellCol int) (int, int) {
	total := t.layout.totalVisualLines()
	if total == 0 {
		return 0, 0
	}

	if visualRow >= total {
		visualRow = total - 1
	}

	return t.layout.viewCellToLogical(visualRow, cellCol)
}

// visualLineLength は指定視覚行のルーン数を返す。
func (t *simpleTextArea) visualLineLength(visualRow int) int {
	logLine, startRune := t.layout.visualToLogical(visualRow)
	for _, v := range t.layout.visualLinesFor(logLine) {
		if v.startRune == startRune {
			return v.length
		}
	}

	return 0
}

// truncateLineWithScroll は水平スクロール位置から幅分のテキストを返す。
func truncateLineWithScroll(line []rune, scrollX, width int) string {
	if width <= 0 {
		return string(line)
	}

	startRune := utils.CellToRuneIndex(line, scrollX)
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
