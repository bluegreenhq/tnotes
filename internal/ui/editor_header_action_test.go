package ui //nolint:testpackage // 内部アクション型（unexported）を検査するためのホワイトボックステスト

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// actionTypeFromCmd は cmd を実行して actionMsg.action の具体型をテストで判定するためのヘルパー。
func actionTypeFromCmd[T ModelAction](t *testing.T, cmd func() any) (T, bool) {
	t.Helper()

	if cmd == nil {
		var zero T

		return zero, false
	}

	msg := cmd()

	am, ok := msg.(actionMsg)
	if !ok {
		var zero T

		return zero, false
	}

	a, ok := am.action.(T)

	return a, ok
}

func TestEditorHeaderClickNewEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	cmd := h.HandleClick(1)
	assert.NotNil(t, cmd)

	_, ok := actionTypeFromCmd[editorHeaderNewAction](t, func() any { return cmd() })
	assert.True(t, ok, "expected editorHeaderNewAction")
}

func TestEditorHeaderClickMoreEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	moreX := h.Width() - 22
	cmd := h.HandleClick(moreX)
	assert.NotNil(t, cmd)

	_, ok := actionTypeFromCmd[editorHeaderOpenMenuAction](t, func() any { return cmd() })
	assert.True(t, ok, "expected editorHeaderOpenMenuAction")
}

func TestEditorHeaderMenuClickTrashEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.RebuildMenu()
	h.OpenMenu()

	cmd := h.HandleMenuClick(2, 1)
	assert.NotNil(t, cmd)

	_, ok := actionTypeFromCmd[editorHeaderTrashAction](t, func() any { return cmd() })
	assert.True(t, ok, "expected editorHeaderTrashAction")
	assert.False(t, h.MenuOpen())
}

func TestEditorHeaderMenuClickCopyEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetHasContent(true)
	h.RebuildMenu()
	h.OpenMenu()

	cmd := h.HandleMenuClick(2, 9)
	assert.NotNil(t, cmd)

	_, ok := actionTypeFromCmd[editorHeaderCopyAction](t, func() any { return cmd() })
	assert.True(t, ok, "expected editorHeaderCopyAction")
}

func TestEditorHeaderMenuTrashModeEmitsMoveAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetTrashMode(true)
	h.RebuildMenu()
	h.OpenMenu()

	cmd := h.HandleMenuClick(2, 1)
	assert.NotNil(t, cmd)

	_, ok := actionTypeFromCmd[editorHeaderMoveAction](t, func() any { return cmd() })
	assert.True(t, ok, "expected editorHeaderMoveAction")
}

func TestEditorHeaderMenuClickPinEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.RebuildMenu()
	h.OpenMenu()

	cmd := h.HandleMenuClick(2, 3)
	assert.NotNil(t, cmd)

	_, ok := actionTypeFromCmd[editorHeaderPinAction](t, func() any { return cmd() })
	assert.True(t, ok, "expected editorHeaderPinAction")
}

func TestEditorHeaderMenuClickUnpinEmitsAction(t *testing.T) {
	t.Parallel()

	h := NewEditorHeader(60)
	h.SetHasNote(true)
	h.SetPinned(true)
	h.RebuildMenu()
	h.OpenMenu()

	cmd := h.HandleMenuClick(2, 3)
	assert.NotNil(t, cmd)

	_, ok := actionTypeFromCmd[editorHeaderUnpinAction](t, func() any { return cmd() })
	assert.True(t, ok, "expected editorHeaderUnpinAction")
}
