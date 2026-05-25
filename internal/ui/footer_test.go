package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/ui"
)

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
