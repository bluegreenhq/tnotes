package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestFooterClickMore(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons(ui.FooterState{})

	// [More] は x=1 から "[More]" の6文字
	cmd := f.HandleClick(1)
	assert.Nil(t, cmd)
	assert.True(t, f.MenuOpen())
}

func TestFooterClickMoreToggle(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons(ui.FooterState{})

	f.HandleClick(1) // open
	assert.True(t, f.MenuOpen())

	f.HandleClick(1) // close
	assert.False(t, f.MenuOpen())
}

func TestFooterClickMenuItem(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons(ui.FooterState{})
	f.OpenMenu()

	// メニュー内相対座標 y=1 = "New"
	cmd := f.HandleMenuClick(2, 1)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.FooterNew, msg)
	assert.False(t, f.MenuOpen())
}

func TestFooterClickMenuItemTrash(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons(ui.FooterState{TrashMode: true, TrashCount: 1})
	f.OpenMenu()

	// y=1 = "Restore"
	cmd := f.HandleMenuClick(2, 1)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.FooterRestore, msg)
}

func TestFooterClickDisabledMenuItem(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons(ui.FooterState{TrashMode: true, TrashCount: 0})
	f.OpenMenu()

	// y=1 = "Restore" (disabled)
	cmd := f.HandleMenuClick(2, 1)
	assert.Nil(t, cmd)
}

func TestFooterViewClosed(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons(ui.FooterState{})

	view, lines := f.View("", "", 80)
	assert.Equal(t, 3, lines)
	assert.Contains(t, view, "Menu")
	assert.Contains(t, view, "┌")
	assert.Contains(t, view, "└")
}

func TestFooterViewAlways3Lines(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	f.RebuildButtons(ui.FooterState{})
	f.OpenMenu()

	// メニューはオーバーレイなので Footer.View は常に3行
	_, lines := f.View("", "", 80)
	assert.Equal(t, 3, lines)
}

func TestFooterClickDisabled(t *testing.T) {
	t.Parallel()

	f := ui.NewFooter()
	btn := ui.NewFooterButton("[Restore]", ui.HoverRestore)
	btn.Disabled = true
	f.SetButtons([]ui.FooterButton{btn})
	cmd := f.HandleClick(1)
	assert.Nil(t, cmd)
}
