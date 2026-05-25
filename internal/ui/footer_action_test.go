package ui //nolint:testpackage // 内部アクション関数（unexported）を検査するためのホワイトボックステスト

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFooterClickMoreEmitsAction(t *testing.T) {
	t.Parallel()

	f := NewFooter()
	f.RebuildButtons()

	act := f.HandleClick(1)
	assert.True(t, sameAction(act, toggleFooterMenu), "expected toggleFooterMenu")
}

func TestFooterClickMenuQuitEmitsAction(t *testing.T) {
	t.Parallel()

	f := NewFooter()
	f.RebuildButtons()
	f.OpenMenu()

	act := f.HandleMenuClick(2, 3)
	assert.True(t, sameAction(act, quit), "expected quit")
	assert.False(t, f.MenuOpen())
}

func TestFooterClickMenuShortcutsEmitsAction(t *testing.T) {
	t.Parallel()

	f := NewFooter()
	f.RebuildButtons()
	f.OpenMenu()

	act := f.HandleMenuClick(2, 1)
	assert.True(t, sameAction(act, openHelp), "expected openHelp")
	assert.False(t, f.MenuOpen())
}
