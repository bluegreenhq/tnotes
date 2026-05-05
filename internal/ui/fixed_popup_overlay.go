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
	kind         PopupKind
	screenWidth  int
	screenHeight int
}

var _ shared.OverlayComponent = (*FixedPopupOverlay)(nil)

// NewFixedPopupOverlay は FixedPopupOverlay を生成する。
func NewFixedPopupOverlay(menu *tui.PopupMenu, origin func() (int, int), kind PopupKind) *FixedPopupOverlay {
	return &FixedPopupOverlay{
		menu:         menu,
		origin:       origin,
		kind:         kind,
		screenWidth:  0,
		screenHeight: 0,
	}
}

// Menu は内部の PopupMenu を返す。
func (p *FixedPopupOverlay) Menu() *tui.PopupMenu { return p.menu }

// Kind はポップアップ種別を返す。
func (p *FixedPopupOverlay) Kind() PopupKind { return p.kind }

// SetScreenSize は画面サイズを設定する。
func (p *FixedPopupOverlay) SetScreenSize(width, height int) {
	p.screenWidth = width
	p.screenHeight = height
}

// Update はメッセージに応じて状態を更新する。
func (p *FixedPopupOverlay) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.handleKey(msg)
	case tea.MouseClickMsg:
		return p.handleClick(msg)
	case tea.MouseMsg:
		mouse := msg.Mouse()
		ox, oy := p.origin()
		p.menu.SetHoverByPos(mouse.X-ox, mouse.Y-oy)
	}

	return nil
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

func (p *FixedPopupOverlay) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.Code {
	case tea.KeyEscape:
		return p.closedCmd()
	case tea.KeyEnter:
		idx := p.menu.SelectHover()
		if idx < 0 {
			return p.closedCmd()
		}

		return p.selectedCmd(idx)
	default:
		p.menu.HandleKeyNav(msg)
	}

	return nil
}

func (p *FixedPopupOverlay) handleClick(msg tea.MouseClickMsg) tea.Cmd {
	if msg.Button != tea.MouseLeft {
		return p.closedCmd()
	}

	ox, oy := p.origin()
	w := p.menu.Width()
	h := p.menu.Height()

	if msg.X < ox || msg.X >= ox+w || msg.Y < oy || msg.Y >= oy+h {
		return p.closedCmd()
	}

	relX := msg.X - ox
	relY := msg.Y - oy
	idx, hit := p.menu.HandleClick(relX, relY)

	if !hit || idx < 0 {
		return p.closedCmd()
	}

	return p.selectedCmd(idx)
}

func (p *FixedPopupOverlay) selectedCmd(idx int) tea.Cmd {
	kind := p.kind

	return func() tea.Msg {
		return PopupMenuSelectedMsg{Kind: kind, Index: idx}
	}
}

func (p *FixedPopupOverlay) closedCmd() tea.Cmd {
	kind := p.kind

	return func() tea.Msg {
		return PopupMenuClosedMsg{Kind: kind}
	}
}
