package ui //nolint:testpackage // 内部アクション型（unexported）を検査するためのホワイトボックステスト

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFooterClickMoreEmitsAction(t *testing.T) {
	t.Parallel()

	f := NewFooter()
	f.RebuildButtons()

	cmd := f.HandleClick(1)
	assert.NotNil(t, cmd)

	am, ok := cmd().(actionMsg)
	assert.True(t, ok)

	_, isToggle := am.action.(footerToggleMenuAction)
	assert.True(t, isToggle, "expected footerToggleMenuAction")
}

func TestFooterClickMenuQuitEmitsAction(t *testing.T) {
	t.Parallel()

	f := NewFooter()
	f.RebuildButtons()
	f.OpenMenu()

	cmd := f.HandleMenuClick(2, 3)
	assert.NotNil(t, cmd)

	am, ok := cmd().(actionMsg)
	assert.True(t, ok)

	_, isQuit := am.action.(quitAction)
	assert.True(t, isQuit, "expected quitAction")
	assert.False(t, f.MenuOpen())
}

func TestFooterClickMenuShortcutsEmitsAction(t *testing.T) {
	t.Parallel()

	f := NewFooter()
	f.RebuildButtons()
	f.OpenMenu()

	cmd := f.HandleMenuClick(2, 1)
	assert.NotNil(t, cmd)

	am, ok := cmd().(actionMsg)
	assert.True(t, ok)

	_, isOpenHelp := am.action.(openHelpAction)
	assert.True(t, isOpenHelp, "expected openHelpAction")
	assert.False(t, f.MenuOpen())
}
