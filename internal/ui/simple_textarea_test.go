package ui //nolint:testpackage // 内部フィールドへの直接アクセスが必要

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestSimpleTextArea_SetValueAndValue(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetValue("Hello\nWorld")
	assert.Equal(t, "Hello\nWorld", ta.Value())
	assert.Equal(t, 1, ta.Line())
	assert.Equal(t, 5, ta.Column())
}

func TestSimpleTextArea_InsertText(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("")
	ta.Update(tea.KeyPressMsg{Text: "a"})
	assert.Equal(t, "a", ta.Value())

	ta.Update(tea.KeyPressMsg{Text: "b"})
	assert.Equal(t, "ab", ta.Value())
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 2, ta.Column())
}

func TestSimpleTextArea_Backspace(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("abc")
	ta.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	assert.Equal(t, "ab", ta.Value())
	assert.Equal(t, 2, ta.Column())
}

func TestSimpleTextArea_BackspaceAtLineStart(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("Hello\nWorld")
	ta.row = 1
	ta.col = 0
	ta.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	assert.Equal(t, "HelloWorld", ta.Value())
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 5, ta.Column())
}

func TestSimpleTextArea_Delete(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("abc")
	ta.row = 0
	ta.col = 0
	ta.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	assert.Equal(t, "bc", ta.Value())
}

func TestSimpleTextArea_DeleteAtLineEnd(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("Hello\nWorld")
	ta.row = 0
	ta.col = 5
	ta.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	assert.Equal(t, "HelloWorld", ta.Value())
}

func TestSimpleTextArea_Enter(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("HelloWorld")
	ta.row = 0
	ta.col = 5
	ta.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Equal(t, "Hello\nWorld", ta.Value())
	assert.Equal(t, 1, ta.Line())
	assert.Equal(t, 0, ta.Column())
}

func TestSimpleTextArea_CursorMovement(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("ab\ncd\nef")
	ta.row = 1
	ta.col = 1

	// 上移動
	ta.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 1, ta.Column())

	// 先頭行でさらに上 → 動かない
	ta.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, ta.Line())

	// 下移動
	ta.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, ta.Line())

	// 左移動
	ta.col = 1
	ta.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	assert.Equal(t, 0, ta.Column())

	// 行頭で左 → 前の行の末尾
	ta.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 2, ta.Column())

	// 右移動 → 行末を超えると次の行の先頭
	ta.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.Equal(t, 1, ta.Line())
	assert.Equal(t, 0, ta.Column())
}

func TestSimpleTextArea_HomeEnd(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("Hello")
	ta.row = 0
	ta.col = 3

	ta.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	assert.Equal(t, 0, ta.Column())

	ta.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	assert.Equal(t, 5, ta.Column())
}

func TestSimpleTextArea_ScrollIndependent(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(3)

	ta.SetValue("line0\nline1\nline2\nline3\nline4")
	ta.row = 0
	ta.col = 0

	ta.ScrollDown(2)
	assert.Equal(t, 2, ta.ScrollYOffset())
	assert.Equal(t, 0, ta.Line(), "ScrollDown should not move cursor")
	assert.Equal(t, 0, ta.Column())

	ta.ScrollUp(1)
	assert.Equal(t, 1, ta.ScrollYOffset())
	assert.Equal(t, 0, ta.Line(), "ScrollUp should not move cursor")
}

func TestSimpleTextArea_EnsureVisible(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.Focus()

	ta.SetValue("line0\nline1\nline2\nline3\nline4")
	ta.row = 0
	ta.col = 0
	ta.scrollY = 0

	// カーソルを下に移動するとviewportが追従
	ta.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	ta.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	ta.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 3, ta.Line())
	assert.Equal(t, 1, ta.ScrollYOffset(), "viewport should follow cursor down")

	// カーソルを上に移動するとviewportが追従
	ta.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	ta.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	ta.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 0, ta.ScrollYOffset(), "viewport should follow cursor up")
}

func TestSimpleTextArea_UnfocusedIgnoresInput(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)

	ta.SetValue("hello")
	ta.Update(tea.KeyPressMsg{Text: "x"})
	assert.Equal(t, "hello", ta.Value(), "unfocused textarea should ignore input")
}

func TestSimpleTextArea_LineCount(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	assert.Equal(t, 1, ta.LineCount())

	ta.SetValue("a\nb\nc")
	assert.Equal(t, 3, ta.LineCount())
}

func TestSimpleTextArea_MoveToBegin(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetValue("Hello\nWorld")
	assert.Equal(t, 1, ta.Line())

	ta.MoveToBegin()
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 0, ta.Column())
}

func TestSimpleTextArea_SetCursorColumn(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetValue("Hello")
	ta.row = 0

	ta.SetCursorColumn(3)
	assert.Equal(t, 3, ta.Column())

	ta.SetCursorColumn(100)
	assert.Equal(t, 5, ta.Column(), "should clamp to line length")

	ta.SetCursorColumn(-1)
	assert.Equal(t, 0, ta.Column(), "should clamp to 0")
}

func TestSimpleTextArea_CtrlA_MoveToLineStart(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("Hello")
	ta.row = 0
	ta.col = 3

	ta.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl})
	assert.Equal(t, 0, ta.Column())
}

func TestSimpleTextArea_CtrlE_MoveToLineEnd(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("Hello")
	ta.row = 0
	ta.col = 0

	ta.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	assert.Equal(t, 5, ta.Column())
}

func TestSimpleTextArea_CtrlFB_CursorLeftRight(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("ab\ncd")
	ta.row = 0
	ta.col = 1

	// Ctrl+F: 右移動
	ta.Update(tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl})
	assert.Equal(t, 2, ta.Column())

	// Ctrl+F: 行末→次の行の先頭
	ta.Update(tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl})
	assert.Equal(t, 1, ta.Line())
	assert.Equal(t, 0, ta.Column())

	// Ctrl+B: 左移動→前の行の末尾
	ta.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 2, ta.Column())

	// Ctrl+B: 左移動
	ta.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	assert.Equal(t, 1, ta.Column())
}

func TestSimpleTextArea_CtrlNP_CursorUpDown(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("ab\ncd\nef")
	ta.row = 0
	ta.col = 1

	// Ctrl+N: 下移動
	ta.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	assert.Equal(t, 1, ta.Line())
	assert.Equal(t, 1, ta.Column())

	// Ctrl+P: 上移動
	ta.Update(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 1, ta.Column())
}

func TestSimpleTextArea_CtrlD_Delete(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("abc")
	ta.row = 0
	ta.col = 1

	ta.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	assert.Equal(t, "ac", ta.Value())
	assert.Equal(t, 1, ta.Column())
}

func TestSimpleTextArea_CtrlD_DeleteAtLineEnd(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("ab\ncd")
	ta.row = 0
	ta.col = 2

	ta.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	assert.Equal(t, "abcd", ta.Value())
}

func TestSimpleTextArea_CtrlK_KillLine(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("HelloWorld")
	ta.row = 0
	ta.col = 5

	ta.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	assert.Equal(t, "Hello", ta.Value())
	assert.Equal(t, 5, ta.Column())
}

func TestSimpleTextArea_CtrlK_KillLineAtEnd(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Focus()

	ta.SetValue("Hello\nWorld")
	ta.row = 0
	ta.col = 5

	// 行末でCtrl+K → 次の行と結合
	ta.Update(tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl})
	assert.Equal(t, "HelloWorld", ta.Value())
	assert.Equal(t, 1, ta.LineCount())
}

func TestSimpleTextArea_CursorUpClampsCol(t *testing.T) {
	t.Parallel()

	ta := newSimpleTextArea()
	ta.SetWidth(80)
	ta.SetHeight(10)

	ta.SetValue("Hi\nHello")
	ta.row = 1
	ta.col = 4

	ta.CursorUp()
	assert.Equal(t, 0, ta.Line())
	assert.Equal(t, 2, ta.Column(), "col should be clamped to shorter line")
}
