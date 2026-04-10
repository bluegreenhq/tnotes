package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestEditorHeaderNew(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	assert.Equal(t, 60, h.Width())
	assert.False(t, h.MenuOpen())
	assert.False(t, h.HasNote())
	assert.False(t, h.TrashMode())
}

func TestEditorHeaderSetState(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	assert.True(t, h.HasNote())

	h.SetTrashMode(true)
	assert.True(t, h.TrashMode())
}

func TestEditorHeaderSetWidth(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetWidth(80)
	assert.Equal(t, 80, h.Width())
}

func TestEditorHeaderViewNormal(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(20)
	h.SetHasNote(true)
	view := h.View()
	assert.Contains(t, view, "+")
	assert.Contains(t, view, "⋯")
}

func TestEditorHeaderViewNoNote(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(20)
	view := h.View()
	assert.Contains(t, view, "+")
	assert.NotContains(t, view, "⋯")
}

func TestEditorHeaderViewTrashMode(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(20)
	h.SetHasNote(true)
	h.SetTrashMode(true)
	view := h.View()
	assert.NotContains(t, view, "+")
	assert.Contains(t, view, "⋯")
}

func TestEditorHeaderClickNew(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	// "+" は x=1 の位置
	cmd := h.HandleClick(1)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.EditorHeaderNew, msg)
}

func TestEditorHeaderClickMore(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(20)
	h.SetHasNote(true)
	// "⋯" は右端付近
	cmd := h.HandleClick(h.Width() - 2)
	assert.Nil(t, cmd) // メニューを開くだけ、Cmd は返さない
	assert.True(t, h.MenuOpen())
}

func TestEditorHeaderClickMoreToggle(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(20)
	h.SetHasNote(true)
	h.HandleClick(h.Width() - 2) // open
	assert.True(t, h.MenuOpen())
	h.HandleClick(h.Width() - 2) // close
	assert.False(t, h.MenuOpen())
}

func TestEditorHeaderMenuClickTrash(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.RebuildMenu()
	h.OpenMenu()

	// メニュー内: y=1 = "Delete Note"
	cmd := h.HandleMenuClick(2, 1)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.EditorHeaderTrash, msg)
	assert.False(t, h.MenuOpen())
}

func TestEditorHeaderMenuClickCopy(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetHasContent(true)
	h.RebuildMenu()
	h.OpenMenu()

	// メニュー内: y=7 = "Copy Note"（Delete Note=1, sep=2, Pin Note=3, sep=4, Move to…=5, sep=6, Copy Note=7）
	cmd := h.HandleMenuClick(2, 7)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.EditorHeaderCopy, msg)
}

func TestEditorHeaderMenuTrashMode(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetTrashMode(true)
	h.RebuildMenu()
	h.OpenMenu()

	// メニュー内: y=1 = "Restore"
	cmd := h.HandleMenuClick(2, 1)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.EditorHeaderRestore, msg)
}

func TestEditorHeaderMenuNoCopyWhenEmpty(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetHasContent(false)
	h.RebuildMenu()

	// メニューに "Copy Note" が含まれない（項目は Delete Note + Pin Note + Move to…）
	items := h.PopupMenu.Items()
	assert.Len(t, items, 3)
	assert.Equal(t, "Delete Note", items[0].Label)
	assert.Equal(t, "Pin Note", items[1].Label)
	assert.Equal(t, "Move to…", items[2].Label)
}

func TestEditorHeaderMenuClickPin(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.RebuildMenu()
	h.OpenMenu()

	// メニュー内: y=3 = "Pin Note"
	cmd := h.HandleMenuClick(2, 3)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.EditorHeaderPin, msg)
}

func TestEditorHeaderMenuClickUnpin(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetPinned(true)
	h.RebuildMenu()
	h.OpenMenu()

	// メニュー内: y=3 = "Unpin Note"
	cmd := h.HandleMenuClick(2, 3)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.EditorHeaderUnpin, msg)
}

func TestEditorHeaderClickNewInTrashMode(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetTrashMode(true)
	// trashMode では + は非表示なので x=1 クリックは何もしない
	cmd := h.HandleClick(1)
	assert.Nil(t, cmd)
}

func TestEditorHeaderHover(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(20)
	h.SetHasNote(true)
	h.SetHover(1) // + ボタン位置
	assert.True(t, h.HoverNew())
	assert.False(t, h.HoverMore())

	h.SetHover(h.Width() - 2) // ⋯ ボタン位置
	assert.False(t, h.HoverNew())
	assert.True(t, h.HoverMore())
}
