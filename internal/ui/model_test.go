package ui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/ui"
)

func mustModel(t *testing.T, m tea.Model) *ui.Model {
	t.Helper()

	model, ok := m.(*ui.Model)
	if !ok {
		t.Fatal("unexpected model type")
	}

	return model
}

func sized(t *testing.T, m *ui.Model) *ui.Model {
	t.Helper()

	ret, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	return mustModel(t, ret)
}

func newTestModel() *ui.Model {
	a, _ := app.New(nil)

	return ui.InitialModel(a, false)
}

func TestModelEmpty(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	assert.Empty(t, m.App.Notes)
	assert.Equal(t, ui.FocusSidebar, m.Focus)
}

func TestCreateNote(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	model := mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)
	assert.Equal(t, ui.FocusEditor, model.Focus)
}

func TestToggleFocus(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// ノートを作成
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	model := mustModel(t, ret)
	assert.Equal(t, ui.FocusEditor, model.Focus)

	// Esc でサイドバーへ（EditorEscape msg が cmd 経由で返る）
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusSidebar, model.Focus)

	// Tab でエディタへ
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusEditor, model.Focus)
}

func TestSidebarNavigation(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// 2つノートを作成
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)
	assert.Len(t, model.App.Notes, 2)

	// j で下に移動
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'j'})
	model = mustModel(t, ret)
	assert.Equal(t, 1, model.Sidebar.SelectedIndex())
}

func TestRenderView(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	view := m.View()
	assert.Contains(t, view.Content, "Notes (0)")
	assert.Contains(t, view.Content, "Press 'n'")
}

func TestDeleteNote(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// 2つノートを作成
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)
	assert.Len(t, model.App.Notes, 2)

	// d で先頭のノートを削除
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'd'})
	model = mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)
	assert.Equal(t, 0, model.Sidebar.SelectedIndex())
	assert.Equal(t, ui.FocusSidebar, model.Focus)
}

func TestDeleteLastNote(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)

	// 最後の1件を削除 → エディタクリア
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'd'})
	model = mustModel(t, ret)
	assert.Empty(t, model.App.Notes)
}

func TestDeleteNoteFromEnd(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// 2つノートを作成
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	// 末尾に移動
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'j'})
	model := mustModel(t, ret)
	assert.Equal(t, 1, model.Sidebar.SelectedIndex())

	// 末尾のノートを削除 → 選択は1つ上に
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'd'})
	model = mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)
	assert.Equal(t, 0, model.Sidebar.SelectedIndex())
}

func TestTrashMode(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// ノート作成して削除
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'd'})
	model := mustModel(t, ret)
	assert.Empty(t, model.App.Notes)
	assert.False(t, model.App.TrashMode)

	// g でゴミ箱モードに入る
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'g'})
	model = mustModel(t, ret)
	assert.True(t, model.App.TrashMode)

	// g で戻る
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'g'})
	model = mustModel(t, ret)
	assert.False(t, model.App.TrashMode)
}

func TestTrashModeNoFocusToEditor(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// ノート作成→削除→ゴミ箱モード
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'd'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'g'})
	model := mustModel(t, ret)
	assert.True(t, model.App.TrashMode)
	assert.Equal(t, ui.FocusSidebar, model.Focus)

	// Tab でエディタに移動できない
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusSidebar, model.Focus)

	// Enter でもエディタに移動できない
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusSidebar, model.Focus)
}

func TestRestoreNote(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// ノート2つ作成
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	// 1つ削除
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'd'})
	model := mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)

	// ゴミ箱モードに入る
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'g'})
	model = mustModel(t, ret)
	assert.True(t, model.App.TrashMode)

	// r で復元 → 通常モードに戻り、ノート数が2に
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'r'})
	model = mustModel(t, ret)
	assert.False(t, model.App.TrashMode)
	assert.Len(t, model.App.Notes, 2)
}

func TestTrashModeView(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// ノート作成→削除→ゴミ箱モード
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'd'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'g'})
	model := mustModel(t, ret)

	view := model.View()
	assert.Contains(t, view.Content, "Trash (1)")
}

func TestMouseDragSelection(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// ノート作成してテキストを入力
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	for _, ch := range "Hello World" {
		ret, _ = ret.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}

	model := mustModel(t, ret)

	// エディタ領域でマウスダウン（sidebarWidthPx=32 なので x=33 がエディタ先頭付近）
	ret, _ = model.Update(tea.MouseClickMsg{X: 33, Y: 0, Button: tea.MouseLeft})
	model = mustModel(t, ret)

	// マウスドラッグ（MouseMotionMsg）
	ret, _ = model.Update(tea.MouseMotionMsg{X: 38, Y: 0, Button: tea.MouseLeft})
	model = mustModel(t, ret)

	// マウスリリース
	ret, _ = model.Update(tea.MouseReleaseMsg{X: 38, Y: 0, Button: tea.MouseLeft})
	model = mustModel(t, ret)

	assert.True(t, model.Editor.HasSelection())
}

func TestFooterShowsCopyCutWhenSelected(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// ノート作成してテキストを入力
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	for _, ch := range "Hello World" {
		ret, _ = ret.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}

	model := mustModel(t, ret)

	// 選択なし: Copy/Cut ボタンなし
	view := model.View()
	assert.NotContains(t, view.Content, "[Copy]")
	assert.NotContains(t, view.Content, "[Cut]")

	// エディタで選択をシミュレート（マウスドラッグ）
	ret, _ = model.Update(tea.MouseClickMsg{X: 33, Y: 0, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseMotionMsg{X: 38, Y: 0, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseReleaseMsg{X: 38, Y: 0, Button: tea.MouseLeft})
	model = mustModel(t, ret)

	// 選択あり: Copy/Cut ボタンが表示される
	view = model.View()
	assert.Contains(t, view.Content, "[Copy]")
	assert.Contains(t, view.Content, "[Cut]")
}

func TestSelectionClearedOnNoteSwitch(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	// 2つノート作成（テキスト入力あり）
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	for _, ch := range "First note" {
		ret, _ = ret.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}

	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'n'})

	for _, ch := range "Second note" {
		ret, _ = ret.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}

	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)

	// Tab でエディタへ移動し選択
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	ret, _ = ret.Update(tea.MouseClickMsg{X: 33, Y: 0, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseMotionMsg{X: 40, Y: 0, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseReleaseMsg{X: 40, Y: 0, Button: tea.MouseLeft})
	model = mustModel(t, ret)
	assert.True(t, model.Editor.HasSelection())

	// Esc でサイドバーに戻り、j で別のノートに移動
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'j'})
	model = mustModel(t, ret)

	// 選択がクリアされている
	assert.False(t, model.Editor.HasSelection())
}

func TestNoteUndoAfterTrash(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)

	ret, _ = model.Update(tea.KeyPressMsg{Code: 'd'})
	model = mustModel(t, ret)
	assert.Empty(t, model.App.Notes)

	// サイドバーフォーカスでCtrl+Z → undo
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	model = mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)
	// フッターにredo案内が表示される
	view := model.View()
	assert.Contains(t, view.Content, "Redo: Ctrl+Shift+Z")
}

func TestNoteRedoAfterUndo(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'd'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	model := mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)

	// Ctrl+Shift+Z → redo
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl | tea.ModShift})
	model = mustModel(t, ret)
	assert.Empty(t, model.App.Notes)
}

func TestNoteUndoAfterCreate(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)

	// サイドバーでCtrl+Z → 作成の取り消し
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	model = mustModel(t, ret)
	assert.Empty(t, model.App.Notes)
}

func TestEditorUndoViaCtrlZ(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	model := mustModel(t, ret)
	assert.Equal(t, ui.FocusEditor, model.Focus)

	for _, ch := range "Hello" {
		ret, _ = ret.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}

	model = mustModel(t, ret)
	assert.Contains(t, model.Editor.Value(), "Hello")

	// Ctrl+Z でundo
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	model = mustModel(t, ret)
	assert.True(t, len(model.Editor.Value()) < len("Hello") || model.Editor.Value() == "")
}

func TestEditorRedoViaCtrlShiftZ(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})

	for _, ch := range "Hi" {
		ret, _ = ret.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
	}

	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl})
	model := mustModel(t, ret)
	beforeRedo := model.Editor.Value()

	ret, _ = model.Update(tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl | tea.ModShift})
	model = mustModel(t, ret)
	assert.NotEqual(t, beforeRedo, model.Editor.Value())
}

func TestInfoMsgClearedOnArrowKey(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)

	// 削除
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'd'})
	model = mustModel(t, ret)
	view := model.View()
	assert.Contains(t, view.Content, "Undo: Ctrl+Z")

	// j キーで移動 → infoMsg が消える
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'j'})
	model = mustModel(t, ret)
	view = model.View()
	assert.NotContains(t, view.Content, "Undo: Ctrl+Z")
}

func TestSidebarResizeDrag(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())

	// ノート作成（サイドバーに内容がある状態）
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)

	// 境界付近（X=32付近）でマウスクリック
	ret, _ = model.Update(tea.MouseClickMsg{X: 32, Y: 5, Button: tea.MouseLeft})
	model = mustModel(t, ret)

	// ドラッグでX=40に移動
	ret, _ = model.Update(tea.MouseMotionMsg{X: 40, Y: 5, Button: tea.MouseLeft})
	model = mustModel(t, ret)

	// リリース
	ret, _ = model.Update(tea.MouseReleaseMsg{X: 40, Y: 5, Button: tea.MouseLeft})
	model = mustModel(t, ret)

	// サイドバー幅が変わっている
	assert.Equal(t, 40, model.SidebarWidth())
}

func TestSidebarResizeMinMax(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel()) // width=100

	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)

	// 最小幅以下にドラッグ
	ret, _ = model.Update(tea.MouseClickMsg{X: 32, Y: 5, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseMotionMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseReleaseMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	model = mustModel(t, ret)
	assert.Equal(t, 20, model.SidebarWidth()) // minSidebarWidth

	// 最大幅以上にドラッグ
	ret, _ = model.Update(tea.MouseClickMsg{X: 20, Y: 5, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseMotionMsg{X: 80, Y: 5, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseReleaseMsg{X: 80, Y: 5, Button: tea.MouseLeft})
	model = mustModel(t, ret)
	assert.Equal(t, 80, model.SidebarWidth()) // width*80%
}

func TestSidebarWidthClampedOnWindowResize(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel()) // width=100, sidebarWidth=32

	// ウィンドウを30に縮小 → maxSidebarWidth=24, sidebarWidthは24にクランプ
	ret, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 30})
	model := mustModel(t, ret)
	assert.Equal(t, 24, model.SidebarWidth())
}

func TestWheelOnSidebarDoesNotChangeSelection(t *testing.T) {
	t.Parallel()

	m := sized(t, newTestModel())

	// ノートを5つ作成
	var ret tea.Model = m
	for range 5 {
		ret, _ = ret.Update(tea.KeyPressMsg{Code: 'n'})
		ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	}

	model := mustModel(t, ret)
	assert.Equal(t, ui.FocusSidebar, model.Focus)
	assert.Equal(t, 0, model.Sidebar.SelectedIndex())

	// ホイールダウンでサイドバーの選択は変わらない
	ret, _ = model.Update(tea.MouseWheelMsg(tea.Mouse{X: 5, Y: 5, Button: tea.MouseWheelDown}))
	model = mustModel(t, ret)

	assert.Equal(t, 0, model.Sidebar.SelectedIndex())
}

func TestWheelOnEditor(t *testing.T) {
	t.Parallel()

	m := sized(t, newTestModel())

	// ノートを作成
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	model := mustModel(t, ret)
	assert.Equal(t, ui.FocusEditor, model.Focus)

	// エディタ領域でホイールイベント送信（エラーにならないことを確認）
	ret, _ = model.Update(tea.MouseWheelMsg(tea.Mouse{X: 50, Y: 5, Button: tea.MouseWheelDown}))
	model = mustModel(t, ret)
	assert.NotNil(t, model)
}
