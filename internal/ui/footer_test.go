package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestFooterClickNew(t *testing.T) {
	t.Parallel()

	f := ui.Footer{}
	f.SetButtons([]ui.FooterButton{
		ui.NewFooterButton("[New]", ui.HoverNew),
	})
	cmd := f.HandleClick(1)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, ui.FooterNew, msg)
}

func TestFooterClickDisabled(t *testing.T) {
	t.Parallel()

	f := ui.Footer{}
	btn := ui.NewFooterButton("[Restore]", ui.HoverRestore)
	btn.Disabled = true
	f.SetButtons([]ui.FooterButton{btn})
	cmd := f.HandleClick(1)
	assert.Nil(t, cmd)
}
