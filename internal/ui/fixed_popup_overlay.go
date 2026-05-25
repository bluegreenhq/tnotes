package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/ui/shared"
)

// FixedPopupOverlay は UI 要素から決まる固定位置に表示されるポップアップメニュー。
// editor header の「…」ボタンや folder list の moreボタンから開かれる。
type FixedPopupOverlay struct {
	menu         *tui.PopupMenu
	origin       func() (int, int) // 描画時に呼ばれて (x, y) を返す
	onSelect     popupSelectFunc
	onClose      ModelAction
	screenWidth  int
	screenHeight int
}

var (
	_ shared.OverlayComponent = (*FixedPopupOverlay)(nil)
	_ popupOverlay            = (*FixedPopupOverlay)(nil)
)

// NewFixedPopupOverlay は FixedPopupOverlay を生成する。
// onSelect はメニュー項目選択時、onClose は閉じ時の pane 側後始末アクション（不要なら nil）。
func NewFixedPopupOverlay(
	menu *tui.PopupMenu,
	origin func() (int, int),
	onSelect popupSelectFunc,
	onClose ModelAction,
) *FixedPopupOverlay {
	return &FixedPopupOverlay{
		menu:         menu,
		origin:       origin,
		onSelect:     onSelect,
		onClose:      onClose,
		screenWidth:  0,
		screenHeight: 0,
	}
}

// Menu は内部の PopupMenu を返す。
func (p *FixedPopupOverlay) Menu() *tui.PopupMenu { return p.menu }

// OnClose は閉じ時の pane 側後始末アクションを返す。
func (p *FixedPopupOverlay) OnClose() ModelAction { return p.onClose }

// SetScreenSize は画面サイズを設定する。
func (p *FixedPopupOverlay) SetScreenSize(width, height int) {
	p.screenWidth = width
	p.screenHeight = height
}

// UpdateOverlay はメッセージに応じて状態を更新し、ModelAction と tea.Cmd を返す。
func (p *FixedPopupOverlay) UpdateOverlay(msg tea.Msg) (ModelAction, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.handleKey(msg), nil
	case tea.MouseClickMsg:
		return p.handleClick(msg), nil
	case tea.MouseMsg:
		mouse := msg.Mouse()
		ox, oy := p.origin()
		p.menu.SetHoverByPos(mouse.X-ox, mouse.Y-oy)
	}

	return nil, nil
}

// RenderOn はベース画面上にメニューを合成する。
func (p *FixedPopupOverlay) RenderOn(base string, _, _ int) string {
	bodyLines := strings.Split(base, "\n")
	menuLines := p.menu.View()

	if len(menuLines) == 0 {
		return base
	}

	ox, oy := p.origin()
	tui.OverlayLines(bodyLines, menuLines, ox, oy)

	return strings.Join(bodyLines, "\n")
}

func (p *FixedPopupOverlay) handleKey(msg tea.KeyPressMsg) ModelAction {
	switch msg.Code {
	case tea.KeyEscape:
		return popupDismiss(p.onClose)
	case tea.KeyEnter:
		idx := p.menu.SelectHover()
		if idx < 0 {
			return popupDismiss(p.onClose)
		}

		return popupSelect(idx, p.onSelect, p.onClose)
	default:
		p.menu.HandleKeyNav(msg)
	}

	return nil
}

func (p *FixedPopupOverlay) handleClick(msg tea.MouseClickMsg) ModelAction {
	if msg.Button != tea.MouseLeft {
		return popupDismiss(p.onClose)
	}

	ox, oy := p.origin()
	w := p.menu.Width()
	h := p.menu.Height()

	if msg.X < ox || msg.X >= ox+w || msg.Y < oy || msg.Y >= oy+h {
		return popupDismiss(p.onClose)
	}

	relX := msg.X - ox
	relY := msg.Y - oy
	idx, hit := p.menu.HandleClick(relX, relY)

	if !hit || idx < 0 {
		return popupDismiss(p.onClose)
	}

	return popupSelect(idx, p.onSelect, p.onClose)
}
