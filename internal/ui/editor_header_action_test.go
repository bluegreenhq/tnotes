package ui //nolint:testpackage // 内部アクション関数（unexported）を検査するためのホワイトボックステスト

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// sameAction は 2 つの ModelAction が同じ関数を指しているかを返す。
// トップレベル関数および empty struct のメソッド値で安定して動作する。
func sameAction(a, b ModelAction) bool {
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

func TestEditorHeaderClickNewEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	act := h.HandleClick(1)
	assert.True(t, sameAction(act, createNote), "expected createNote")
}

func TestEditorHeaderClickMoreEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	moreX := h.Width() - 22
	act := h.HandleClick(moreX)
	assert.True(t, sameAction(act, openEditorHeaderMenu), "expected openEditorHeaderMenu")
}

func TestEditorHeaderMenuClickTrashEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.RebuildMenu()
	h.OpenMenu()

	act := h.HandleMenuClick(2, 1)
	assert.True(t, sameAction(act, trashSelectedNote), "expected trashSelectedNote")
	assert.False(t, h.MenuOpen())
}

func TestEditorHeaderMenuClickCopyEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetHasContent(true)
	h.RebuildMenu()
	h.OpenMenu()

	act := h.HandleMenuClick(2, 9)
	assert.True(t, sameAction(act, copyNoteToClipboard), "expected copyNoteToClipboard")
}

func TestEditorHeaderMenuTrashModeEmitsMoveAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetTrashMode(true)
	h.RebuildMenu()
	h.OpenMenu()

	act := h.HandleMenuClick(2, 1)
	assert.True(t, sameAction(act, openNoteMoveMenu), "expected openNoteMoveMenu")
}

func TestEditorHeaderMenuClickPinEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.RebuildMenu()
	h.OpenMenu()

	act := h.HandleMenuClick(2, 3)
	assert.True(t, sameAction(act, pinNote), "expected pinNote")
}

func TestEditorHeaderMenuClickUnpinEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetPinned(true)
	h.RebuildMenu()
	h.OpenMenu()

	act := h.HandleMenuClick(2, 3)
	assert.True(t, sameAction(act, unpinNote), "expected unpinNote")
}
