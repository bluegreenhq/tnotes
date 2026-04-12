package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestFooterClickMore(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons()

	// [More] は x=1 から "[More]" の6文字
	cmd := f.HandleClick(1)
	assert.Nil(t, cmd)
	assert.True(t, f.MenuOpen())
}

func TestFooterClickMoreToggle(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons()

	f.HandleClick(1) // open
	assert.True(t, f.MenuOpen())

	f.HandleClick(1) // close
	assert.False(t, f.MenuOpen())
}

func TestFooterClickMenuItem(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons()
	f.OpenMenu()

	// メニュー内相対座標 y=3 = "Quit" (y=1=Shortcuts, y=2=空行, y=3=Quit)
	cmd := f.HandleMenuClick(2, 3)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.FooterQuit, msg)
	assert.False(t, f.MenuOpen())
}

func TestFooterClickMenuItemShortcuts(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons()
	f.OpenMenu()

	// メニュー内相対座標 y=1 = "Shortcuts"
	cmd := f.HandleMenuClick(2, 1)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.FooterHelp, msg)
	assert.False(t, f.MenuOpen())
}

func TestFooterClickMenuItemTrash(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons()
	f.OpenMenu()

	// y=3 = "Quit"
	cmd := f.HandleMenuClick(2, 3)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.FooterQuit, msg)
}

func TestFooterViewClosed(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons()

	view, lines := f.View("", "", 80)
	assert.Equal(t, 3, lines)
	assert.Contains(t, view, "Menu")
	assert.Contains(t, view, "┌")
	assert.Contains(t, view, "└")
}

func TestFooterViewAlways3Lines(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons()
	f.OpenMenu()

	// メニューはオーバーレイなので Footer.View は常に3行
	_, lines := f.View("", "", 80)
	assert.Equal(t, 3, lines)
}

func TestFooterClickDisabled(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	btn := ui.NewFooterButton("[Quit]", ui.HoverQuit)
	btn.Disabled = true
	f.SetButtons([]ui.FooterButton{btn})
	cmd := f.HandleClick(1)
	assert.Nil(t, cmd)
}
