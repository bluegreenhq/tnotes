package ui_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/ui"
)

func TestPopupMenuItems(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "New"},
		{Label: "Trash"},
		{Label: "Quit"},
	}
	m := ui.NewPopupMenu(items)
	assert.Equal(t, items, m.Items())
}

func TestPopupMenuHeight(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "New"},
		{Label: "Trash"},
	}
	m := ui.NewPopupMenu(items)
	// 上枠1 + 項目2 + 項目間空行1 + 下枠1 = 5
	assert.Equal(t, 5, m.Height())
}

func TestPopupMenuHeightThreeItems(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "New"},
		{Label: "Trash"},
		{Label: "Quit"},
	}
	m := ui.NewPopupMenu(items)
	// 上枠1 + 項目3 + 項目間空行2 + 下枠1 = 7
	assert.Equal(t, 7, m.Height())
}

func TestPopupMenuWidth(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "New"},
		{Label: "LongLabel"},
	}
	m := ui.NewPopupMenu(items)
	// 左枠1 + スペース1 + max("LongLabel"=9) + スペース1 + 右枠1 = 13
	assert.Equal(t, 13, m.Width())
}

func TestPopupMenuHitTest(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "New"},
		{Label: "Trash"},
		{Label: "Quit"},
	}
	m := ui.NewPopupMenu(items)
	// Height = 7: y=0:上枠, y=1:New, y=2:空行, y=3:Trash, y=4:空行, y=5:Quit, y=6:下枠

	// 項目行
	assert.Equal(t, 0, m.HitTest(2, 1)) // "New"
	assert.Equal(t, 1, m.HitTest(2, 3)) // "Trash"
	assert.Equal(t, 2, m.HitTest(2, 5)) // "Quit"

	// 空行
	assert.Equal(t, -1, m.HitTest(2, 2))
	assert.Equal(t, -1, m.HitTest(2, 4))

	// 枠
	assert.Equal(t, -1, m.HitTest(2, 0))
	assert.Equal(t, -1, m.HitTest(2, 6))

	// X が範囲外
	assert.Equal(t, -1, m.HitTest(20, 1))
}

func TestPopupMenuHitTestDisabled(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "Restore", Disabled: true},
		{Label: "Notes"},
	}
	m := ui.NewPopupMenu(items)

	// disabled の項目は -1
	assert.Equal(t, -1, m.HitTest(2, 1))
	// enabled の項目は OK (y=3: 空行を挟んだ2番目の項目)
	assert.Equal(t, 1, m.HitTest(2, 3))
}

func TestPopupMenuHandleClick(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "New"},
		{Label: "Trash"},
	}
	m := ui.NewPopupMenu(items)

	idx, hit := m.HandleClick(2, 1) // "New"
	assert.True(t, hit)
	assert.Equal(t, 0, idx)

	idx, hit = m.HandleClick(2, 0) // 枠
	assert.False(t, hit)
	assert.Equal(t, -1, idx)
}

func TestPopupMenuSetHoverByPos(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "New"},
		{Label: "Trash"},
	}
	m := ui.NewPopupMenu(items)

	m.SetHoverByPos(2, 1) // "New"
	assert.Equal(t, 0, m.Hover())

	m.SetHoverByPos(2, 0) // 枠
	assert.Equal(t, -1, m.Hover())
}

func TestPopupMenuView(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "New"},
		{Label: "Trash"},
	}
	m := ui.NewPopupMenu(items)
	lines := m.View()

	// 5行: 上枠, New, 空行, Trash, 下枠
	assert.Len(t, lines, 5)
	assert.Contains(t, lines[0], "┌")
	assert.Contains(t, lines[1], "New")
	assert.Contains(t, lines[3], "Trash")
	assert.Contains(t, lines[4], "└")
}

func TestPopupMenuMoveHoverDown(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "A"},
		{Label: "B"},
		{Label: "C"},
	}
	m := ui.NewPopupMenu(items)

	assert.Equal(t, -1, m.Hover())
	m.MoveHoverDown()
	assert.Equal(t, 0, m.Hover())
	m.MoveHoverDown()
	assert.Equal(t, 1, m.Hover())
	m.MoveHoverDown()
	assert.Equal(t, 2, m.Hover())
	m.MoveHoverDown() // 末尾を超えない
	assert.Equal(t, 2, m.Hover())
}

func TestPopupMenuMoveHoverUp(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "A"},
		{Label: "B"},
		{Label: "C"},
	}
	m := ui.NewPopupMenu(items)

	// hover=-1 から MoveHoverUp で末尾に移動
	m.MoveHoverUp()
	assert.Equal(t, 2, m.Hover())
	m.MoveHoverUp()
	assert.Equal(t, 1, m.Hover())
	m.MoveHoverUp()
	assert.Equal(t, 0, m.Hover())
	m.MoveHoverUp() // 先頭を超えない
	assert.Equal(t, 0, m.Hover())
}

func TestPopupMenuMoveHoverSkipsDisabled(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "A"},
		{Label: "B", Disabled: true},
		{Label: "C"},
	}
	m := ui.NewPopupMenu(items)

	m.MoveHoverDown() // → A (0)
	assert.Equal(t, 0, m.Hover())
	m.MoveHoverDown() // → C (2), B はスキップ
	assert.Equal(t, 2, m.Hover())

	m.MoveHoverUp() // → A (0), B はスキップ
	assert.Equal(t, 0, m.Hover())
}

func TestPopupMenuSelectHover(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "A"},
		{Label: "B"},
	}
	m := ui.NewPopupMenu(items)

	assert.Equal(t, -1, m.SelectHover())
	m.MoveHoverDown()
	assert.Equal(t, 0, m.SelectHover())
}

func TestPopupMenuViewDisabled(t *testing.T) {
	t.Parallel()

	items := []ui.MenuItem{
		{Label: "Restore", Disabled: true},
		{Label: "Notes", Disabled: false},
	}
	m := ui.NewPopupMenu(items)
	lines := m.View()

	assert.Contains(t, lines[1], "Restore")
	assert.Contains(t, lines[3], "Notes")
}
