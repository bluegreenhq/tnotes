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
	assert.Equal(t, ui.FocusNoteList, m.Focus)
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
	assert.Equal(t, ui.FocusNoteList, model.Focus)

	// Tab でエディタへ
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusEditor, model.Focus)
}

func TestNoteListNavigation(t *testing.T) {
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
	assert.Equal(t, 1, model.NoteList.SelectedIndex())
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
	assert.Equal(t, 0, model.NoteList.SelectedIndex())
	assert.Equal(t, ui.FocusNoteList, model.Focus)
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
	assert.Equal(t, 1, model.NoteList.SelectedIndex())

	// 末尾のノートを削除 → 選択は1つ上に
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'd'})
	model = mustModel(t, ret)
	assert.Len(t, model.App.Notes, 1)
	assert.Equal(t, 0, model.NoteList.SelectedIndex())
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

	// エディタ領域でマウスダウン（noteListWidth=32 なので x=33 がエディタ先頭付近）
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
	assert.NotContains(t, view.Content, " Copy ")
	assert.NotContains(t, view.Content, " Cut ")

	// エディタで選択をシミュレート（マウスドラッグ）
	ret, _ = model.Update(tea.MouseClickMsg{X: 33, Y: 0, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseMotionMsg{X: 38, Y: 0, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseReleaseMsg{X: 38, Y: 0, Button: tea.MouseLeft})
	model = mustModel(t, ret)

	// 選択あり: Copy/Cut ボタンが表示される
	view = model.View()
	assert.Contains(t, view.Content, " Copy ")
	assert.Contains(t, view.Content, " Cut ")
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

func TestNoteListResizeDrag(t *testing.T) {
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
	assert.Equal(t, 40, model.NoteListWidth())
}

func TestNoteListResizeMinMax(t *testing.T) {
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
	assert.Equal(t, 20, model.NoteListWidth()) // minNoteListWidth

	// 最大幅以上にドラッグ
	ret, _ = model.Update(tea.MouseClickMsg{X: 20, Y: 5, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseMotionMsg{X: 80, Y: 5, Button: tea.MouseLeft})
	ret, _ = ret.Update(tea.MouseReleaseMsg{X: 80, Y: 5, Button: tea.MouseLeft})
	model = mustModel(t, ret)
	assert.Equal(t, 80, model.NoteListWidth()) // width*80%
}

func TestNoteListWidthClampedOnWindowResize(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel()) // width=100, noteListWidth=32

	// ウィンドウを30に縮小 → maxNoteListWidth=24, noteListWidthは24にクランプ
	ret, _ := m.Update(tea.WindowSizeMsg{Width: 30, Height: 30})
	model := mustModel(t, ret)
	assert.Equal(t, 24, model.NoteListWidth())
}

func TestWheelOnNoteListDoesNotChangeSelection(t *testing.T) {
	t.Parallel()

	m := sized(t, newTestModel())

	// ノートを5つ作成
	var ret tea.Model = m
	for range 5 {
		ret, _ = ret.Update(tea.KeyPressMsg{Code: 'n'})
		ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	}

	model := mustModel(t, ret)
	assert.Equal(t, ui.FocusNoteList, model.Focus)
	assert.Equal(t, 0, model.NoteList.SelectedIndex())

	// ホイールダウンでサイドバーの選択は変わらない
	ret, _ = model.Update(tea.MouseWheelMsg(tea.Mouse{X: 5, Y: 5, Button: tea.MouseWheelDown}))
	model = mustModel(t, ret)

	assert.Equal(t, 0, model.NoteList.SelectedIndex())
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

func TestToggleFolderList(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())

	// 初期状態: フォルダ非表示
	assert.False(t, m.FolderList.Visible())

	// Ctrl+B でフォルダ表示
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	model := mustModel(t, ret)
	assert.True(t, model.FolderList.Visible())
	assert.Equal(t, ui.FocusFolderList, model.Focus)

	// もう一度 Ctrl+B で非表示
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	model = mustModel(t, ret)
	assert.False(t, model.FolderList.Visible())
	assert.Equal(t, ui.FocusNoteList, model.Focus)
}

func TestFolderListSelectTrash(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())

	// ノート作成
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)

	// フォルダ表示
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusFolderList, model.Focus)

	// j で Trash を選択
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'j'})
	model = mustModel(t, ret)
	assert.True(t, model.App.TrashMode)

	// k で Notes に戻る
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'k'})
	model = mustModel(t, ret)
	assert.False(t, model.App.TrashMode)
}

func TestFolderListFocusTransition(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())

	// ノート作成
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model := mustModel(t, ret)

	// フォルダ表示
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusFolderList, model.Focus)

	// Tab → NoteList
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusNoteList, model.Focus)

	// Tab → Editor
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusEditor, model.Focus)

	// Esc → NoteList
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusNoteList, model.Focus)

	// Esc → FolderList
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	model = mustModel(t, ret)
	assert.Equal(t, ui.FocusFolderList, model.Focus)
}

func TestCtrlBInEditorDoesNotToggleFolder(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())

	// ノート作成（エディタにフォーカス）
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	model := mustModel(t, ret)
	assert.Equal(t, ui.FocusEditor, model.Focus)

	// Ctrl+B → フォルダトグルではなくエディタの操作
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	model = mustModel(t, ret)
	assert.False(t, model.FolderList.Visible())
}

func TestTrashModeViaFolder(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())

	// ノート作成して削除
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'd'})
	model := mustModel(t, ret)
	assert.Empty(t, model.App.Notes)

	// フォルダ表示 → Trash 選択
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'j'})
	model = mustModel(t, ret)
	assert.True(t, model.App.TrashMode)

	// Trash のビュー確認
	view := model.View()
	assert.Contains(t, view.Content, "Trash")

	// Notes に戻る
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'k'})
	model = mustModel(t, ret)
	assert.False(t, model.App.TrashMode)
}

func TestRestoreNoteViaFolder(t *testing.T) {
	t.Parallel()
	m := sized(t, newTestModel())

	// ノート作成して削除
	ret, _ := m.Update(tea.KeyPressMsg{Code: 'n'})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'd'})
	model := mustModel(t, ret)

	// フォルダ表示 → Trash 選択
	ret, _ = model.Update(tea.KeyPressMsg{Code: 'b', Mod: tea.ModCtrl})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'j'})
	model = mustModel(t, ret)
	assert.True(t, model.App.TrashMode)

	// Tab → NoteList → r で復元
	ret, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	ret, _ = ret.Update(tea.KeyPressMsg{Code: 'r'})
	model = mustModel(t, ret)
	assert.False(t, model.App.TrashMode)
	assert.Len(t, model.App.Notes, 1)
}
