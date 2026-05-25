package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/ui/shared"
)

// AnchoredPopupOverlay は右クリックで開くアンカー位置に追従するポップアップメニュー。
type AnchoredPopupOverlay struct {
	menu         *tui.PopupMenu
	anchorX      int
	anchorY      int
	kind         PopupKind
	screenWidth  int
	screenHeight int
}

var (
	_ shared.OverlayComponent = (*AnchoredPopupOverlay)(nil)
	_ shared.AnchoredOverlay  = (*AnchoredPopupOverlay)(nil)
)

// NewAnchoredPopupOverlay は AnchoredPopupOverlay を生成する。
func NewAnchoredPopupOverlay(menu *tui.PopupMenu, anchorX, anchorY int, kind PopupKind) *AnchoredPopupOverlay {
	return &AnchoredPopupOverlay{
		menu:         menu,
		anchorX:      anchorX,
		anchorY:      anchorY,
		kind:         kind,
		screenWidth:  0,
		screenHeight: 0,
	}
}

// Menu は内部の PopupMenu を返す。
func (p *AnchoredPopupOverlay) Menu() *tui.PopupMenu { return p.menu }

// Kind はポップアップ種別を返す。
func (p *AnchoredPopupOverlay) Kind() PopupKind { return p.kind }

// AnchorX はアンカーX座標を返す。
func (p *AnchoredPopupOverlay) AnchorX() int { return p.anchorX }

// AnchorY はアンカーY座標を返す。
func (p *AnchoredPopupOverlay) AnchorY() int { return p.anchorY }

// MenuOrigin はクランプ済みのメニュー描画左上座標を返す。
func (p *AnchoredPopupOverlay) MenuOrigin(screenWidth, screenHeight int) (int, int) {
	w, h := p.menuSize()

	return tui.ClampMenuOrigin(w, h, p.anchorX, p.anchorY, screenWidth, screenHeight)
}

// SetScreenSize は画面サイズを設定する。
func (p *AnchoredPopupOverlay) SetScreenSize(width, height int) {
	p.screenWidth = width
	p.screenHeight = height
}

// Update はメッセージに応じて状態を更新する。
func (p *AnchoredPopupOverlay) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.handleKey(msg)
	case tea.MouseClickMsg:
		return p.handleClick(msg)
	case tea.MouseMsg:
		mouse := msg.Mouse()
		ox, oy := p.MenuOrigin(p.screenWidth, p.screenHeight)
		p.menu.SetHoverByPos(mouse.X-ox, mouse.Y-oy)
	}

	return nil
}

// RenderOn はベース画面上にメニューを合成する。
func (p *AnchoredPopupOverlay) RenderOn(base string, width, height int) string {
	bodyLines := strings.Split(base, "\n")
	menuLines := p.menu.View()

	if len(menuLines) == 0 {
		return base
	}

	x, y := p.MenuOrigin(width, height)
	tui.OverlayLines(bodyLines, menuLines, x, y)

	return strings.Join(bodyLines, "\n")
}

func (p *AnchoredPopupOverlay) menuSize() (int, int) {
	menuLines := p.menu.View()
	if len(menuLines) == 0 {
		return 0, 0
	}

	return lipgloss.Width(menuLines[0]), len(menuLines)
}

func (p *AnchoredPopupOverlay) handleKey(msg tea.KeyPressMsg) tea.Cmd {
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

func (p *AnchoredPopupOverlay) handleClick(msg tea.MouseClickMsg) tea.Cmd {
	if msg.Button != tea.MouseLeft {
		return p.closedCmd()
	}

	ox, oy := p.MenuOrigin(p.screenWidth, p.screenHeight)
	relX := msg.X - ox
	relY := msg.Y - oy
	idx, hit := p.menu.HandleClick(relX, relY)

	if !hit {
		return p.closedCmd()
	}

	if idx < 0 {
		return p.closedCmd()
	}

	return p.selectedCmd(idx)
}

func (p *AnchoredPopupOverlay) selectedCmd(idx int) tea.Cmd {
	return actionCmd(popupMenuSelectedAction{Kind: p.kind, Index: idx})
}

func (p *AnchoredPopupOverlay) closedCmd() tea.Cmd {
	return actionCmd(popupMenuClosedAction{Kind: p.kind})
}

// --- PopupOverlay → Model アクション ---
// AnchoredPopupOverlay と FixedPopupOverlay が共有する。

// popupMenuSelectedAction はポップアップメニュー項目選択を Model に通知する。
type popupMenuSelectedAction struct {
	Kind  PopupKind
	Index int
}

// Apply はポップアップを閉じ、対応するメニューアクションを実行する。
func (a popupMenuSelectedAction) Apply(m *Model, ctx ActionContext) tea.Cmd {
	m.closePopupOverlay(a.Kind)

	return m.executePopupAction(a.Kind, a.Index, ctx.Now)
}

// popupMenuClosedAction はポップアップメニューが選択無しで閉じられたことを Model に通知する。
type popupMenuClosedAction struct {
	Kind PopupKind
}

// Apply はポップアップを閉じる。
func (a popupMenuClosedAction) Apply(m *Model, _ ActionContext) tea.Cmd {
	m.closePopupOverlay(a.Kind)

	return nil
}
