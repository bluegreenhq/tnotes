package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

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
	maxScroll := max(len(t.lines)-t.height, 0)
	t.scrollY = min(t.scrollY+n, maxScroll)
}

// CursorUp はカーソルを1行上に移動する。
func (t *simpleTextArea) CursorUp() {
	if t.row <= 0 {
		return
	}

	t.row--
	t.clampCol()
	t.ensureVisible()
}

// CursorDown はカーソルを1行下に移動する。
func (t *simpleTextArea) CursorDown() {
	if t.row >= len(t.lines)-1 {
		return
	}

	t.row++
	t.clampCol()
	t.ensureVisible()
}

func (t *simpleTextArea) clampCol() {
	maxCol := len(t.lines[t.row])
	if t.col > maxCol {
		t.col = maxCol
	}
}

func (t *simpleTextArea) handleKey(msg tea.KeyPressMsg) tea.Cmd { //nolint:cyclop // キーバインド分岐
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
	case msg.Text != "" && msg.Mod == 0:
		t.insertText(msg.Text)
	case msg.Code == tea.KeyEnter:
		t.insertNewline()
	case msg.Code == tea.KeyBackspace:
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
	t.ensureVisible()
}

func (t *simpleTextArea) backspace() {
	if t.col > 0 {
		line := t.lines[t.row]
		t.lines[t.row] = append(line[:t.col-1], line[t.col:]...)
		t.col--
	} else if t.row > 0 {
		prevLen := len(t.lines[t.row-1])
		t.lines[t.row-1] = append(t.lines[t.row-1], t.lines[t.row]...)
		t.lines = append(t.lines[:t.row], t.lines[t.row+1:]...)
		t.row--
		t.col = prevLen
		t.ensureVisible()
	}
}

func (t *simpleTextArea) delete() {
	line := t.lines[t.row]
	if t.col < len(line) {
		t.lines[t.row] = append(line[:t.col], line[t.col+1:]...)
	} else if t.row < len(t.lines)-1 {
		t.lines[t.row] = append(t.lines[t.row], t.lines[t.row+1]...)
		t.lines = append(t.lines[:t.row+1], t.lines[t.row+2:]...)
	}
}

func (t *simpleTextArea) killLine() {
	line := t.lines[t.row]
	if t.col < len(line) {
		t.lines[t.row] = line[:t.col]
	} else if t.row < len(t.lines)-1 {
		t.lines[t.row] = append(t.lines[t.row], t.lines[t.row+1]...)
		t.lines = append(t.lines[:t.row+1], t.lines[t.row+2:]...)
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
