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

func TestEditorHeaderMenuNoCopyWhenEmpty(t *testing.T) {
	t.Parallel()

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetHasContent(false)
	h.RebuildMenu()

	// メニューに "Copy Note" が含まれない（項目は Delete Note + Pin Note + Move to… + Duplicate）
	items := h.PopupMenu.Items()
	assert.Len(t, items, 4)
	assert.Equal(t, "Delete Note", items[0].Label)
	assert.Equal(t, "Pin Note", items[1].Label)
	assert.Equal(t, "Move to…", items[2].Label)
	assert.Equal(t, "Duplicate", items[3].Label)
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

	h := ui.NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetHover(1) // + ボタン位置
	assert.True(t, h.HoverNew())
	assert.False(t, h.HoverMore())

	moreX := h.Width() - 22 // width - searchFieldWidth(20) - moreButtonOffset(2)
	h.SetHover(moreX)       // ⋯ ボタン位置
	assert.False(t, h.HoverNew())
	assert.True(t, h.HoverMore())
}
