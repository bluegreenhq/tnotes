package ui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestRightClickEditorOpensContextMenu(t *testing.T) {
	t.Parallel()

	m := sized(t, newTestModel())
	ret := createNoteWithText(t, m, "Hello World")
	m = mustModel(t, ret)

	// エディタにフォーカスして全選択
	ret, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = mustModel(t, ret)
	ret, _ = m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl | tea.ModShift})
	m = mustModel(t, ret)

	assert.True(t, m.Editor.HasSelection())

	// エディタ領域で右クリック
	ret, _ = m.Update(tea.MouseClickMsg{
		X: 50, Y: 5, Button: tea.MouseRight,
	})
	m = mustModel(t, ret)
	assert.True(t, m.Editor.IsContextMenuOpen())
}

func TestRightClickEditorNoSelectionOpensMenu(t *testing.T) {
	t.Parallel()

	m := sized(t, newTestModel())
	ret := createNoteWithText(t, m, "Hello World")
	m = mustModel(t, ret)

	// 選択なしでも右クリックでメニューが開く（Paste が使える）
	ret, _ = m.Update(tea.MouseClickMsg{
		X: 50, Y: 5, Button: tea.MouseRight,
	})
	m = mustModel(t, ret)
	assert.True(t, m.Editor.IsContextMenuOpen())
	// Copy と Cut は Disabled
	items := m.Editor.ContextMenu.Items()
	assert.True(t, items[0].Disabled, "Copy should be disabled without selection")
	assert.True(t, items[1].Disabled, "Cut should be disabled without selection")
	assert.False(t, items[2].Disabled, "Paste should be enabled")
}

func TestRightClickNoteListOpensEditorHeaderMenu(t *testing.T) {
	t.Parallel()

	m := sized(t, newTestModel())
	ret := createNoteWithText(t, m, "Hello World")
	m = mustModel(t, ret)

	assert.Len(t, m.App.Notes, 1, "should have 1 note")

	// NoteList 領域で右クリック（Y=4: NoteListヘッダー2行 + セクションヘッダー2行）
	ret, _ = m.Update(tea.MouseClickMsg{
		X: 5, Y: 4, Button: tea.MouseRight,
	})
	m = mustModel(t, ret)
	// EditorHeader のメニューが開く
	assert.True(t, m.Editor.Header.MenuOpen())
}

func TestRightClickClosedByLeftClick(t *testing.T) {
	t.Parallel()

	m := sized(t, newTestModel())
	ret := createNoteWithText(t, m, "Hello World")
	m = mustModel(t, ret)

	// エディタにフォーカスして全選択
	ret, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = mustModel(t, ret)
	ret, _ = m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl | tea.ModShift})
	m = mustModel(t, ret)

	// 右クリックでメニューを開く
	ret, _ = m.Update(tea.MouseClickMsg{
		X: 50, Y: 5, Button: tea.MouseRight,
	})
	m = mustModel(t, ret)
	assert.True(t, m.Editor.IsContextMenuOpen())

	// 左クリックで閉じる
	ret, _ = m.Update(tea.MouseClickMsg{
		X: 0, Y: 0, Button: tea.MouseLeft,
	})
	m = mustModel(t, ret)
	assert.False(t, m.Editor.IsContextMenuOpen())
}

func TestContextMenuOverlayWidthWithJapanese(t *testing.T) {
	t.Parallel()

	const termWidth = 100

	m := sized(t, newTestModel())
	ret := createNoteWithText(t, m, "あいうえおかきくけこさしすせそ")
	m = mustModel(t, ret)

	// エディタにフォーカスして全選択
	ret, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = mustModel(t, ret)
	ret, _ = m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl | tea.ModShift})
	m = mustModel(t, ret)

	// エディタ領域で右クリック（日本語テキスト上）
	ret, _ = m.Update(tea.MouseClickMsg{
		X: 40, Y: 1, Button: tea.MouseRight,
	})
	m = mustModel(t, ret)
	assert.True(t, m.Editor.IsContextMenuOpen())

	// コンテキストメニュー表示時のView幅を取得
	view := m.View()
	lines := strings.Split(view.Content, "\n")

	// メニューが重なる行の幅がターミナル幅を超えていないことを確認
	for i, line := range lines {
		w := lipgloss.Width(line)
		assert.LessOrEqual(t, w, termWidth, "line %d width %d exceeds terminal width %d", i, w, termWidth)
	}
}

func TestRightClickClosedByEscape(t *testing.T) {
	t.Parallel()

	m := sized(t, newTestModel())
	ret := createNoteWithText(t, m, "Hello World")
	m = mustModel(t, ret)

	// エディタにフォーカスして全選択
	ret, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	m = mustModel(t, ret)
	ret, _ = m.Update(tea.KeyPressMsg{Code: 'a', Mod: tea.ModCtrl | tea.ModShift})
	m = mustModel(t, ret)

	// 右クリックでメニューを開く
	ret, _ = m.Update(tea.MouseClickMsg{
		X: 50, Y: 5, Button: tea.MouseRight,
	})
	m = mustModel(t, ret)
	assert.True(t, m.Editor.IsContextMenuOpen())

	// Escape で閉じる
	ret, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = mustModel(t, ret)
	assert.False(t, m.Editor.IsContextMenuOpen())
}
