package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/ui/shared"
)

// popupSelectFunc は popup メニュー項目選択時に呼ばれる関数。
// メニュー内の選択 idx に対応する後続 ModelAction を返す。
type popupSelectFunc func(idx int) ModelAction

// AnchoredPopupOverlay は右クリックで開くアンカー位置に追従するポップアップメニュー。
type AnchoredPopupOverlay struct {
	menu         *tui.PopupMenu
	anchorX      int
	anchorY      int
	onSelect     popupSelectFunc
	onClose      ModelAction
	screenWidth  int
	screenHeight int
}

var (
	_ shared.OverlayComponent = (*AnchoredPopupOverlay)(nil)
	_ shared.AnchoredOverlay  = (*AnchoredPopupOverlay)(nil)
	_ popupOverlay            = (*AnchoredPopupOverlay)(nil)
)

// NewAnchoredPopupOverlay は AnchoredPopupOverlay を生成する。
// onSelect はメニュー項目選択時、onClose は閉じ時の pane 側後始末アクション（不要なら nil）。
func NewAnchoredPopupOverlay(
	menu *tui.PopupMenu,
	anchorX, anchorY int,
	onSelect popupSelectFunc,
	onClose ModelAction,
) *AnchoredPopupOverlay {
	return &AnchoredPopupOverlay{
		menu:         menu,
		anchorX:      anchorX,
		anchorY:      anchorY,
		onSelect:     onSelect,
		onClose:      onClose,
		screenWidth:  0,
		screenHeight: 0,
	}
}

// Menu は内部の PopupMenu を返す。
func (p *AnchoredPopupOverlay) Menu() *tui.PopupMenu { return p.menu }

// OnClose は閉じ時の pane 側後始末アクションを返す。
func (p *AnchoredPopupOverlay) OnClose() ModelAction { return p.onClose }

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

// UpdateOverlay はメッセージに応じて状態を更新し、ModelAction と tea.Cmd を返す。
func (p *AnchoredPopupOverlay) UpdateOverlay(msg tea.Msg) (ModelAction, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return p.handleKey(msg), nil
	case tea.MouseClickMsg:
		return p.handleClick(msg), nil
	case tea.MouseMsg:
		mouse := msg.Mouse()
		ox, oy := p.MenuOrigin(p.screenWidth, p.screenHeight)
		p.menu.SetHoverByPos(mouse.X-ox, mouse.Y-oy)
	}

	return nil, nil
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

func (p *AnchoredPopupOverlay) handleKey(msg tea.KeyPressMsg) ModelAction {
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

func (p *AnchoredPopupOverlay) handleClick(msg tea.MouseClickMsg) ModelAction {
	if msg.Button != tea.MouseLeft {
		return popupDismiss(p.onClose)
	}

	ox, oy := p.MenuOrigin(p.screenWidth, p.screenHeight)
	relX := msg.X - ox
	relY := msg.Y - oy
	idx, hit := p.menu.HandleClick(relX, relY)

	if !hit || idx < 0 {
		return popupDismiss(p.onClose)
	}

	return popupSelect(idx, p.onSelect, p.onClose)
}

// --- PopupOverlay → Model アクション ---
// AnchoredPopupOverlay と FixedPopupOverlay が共有する。

// popupSelect は overlay を閉じてから onClose と onSelect(idx) を順に適用する。
func popupSelect(idx int, onSelect popupSelectFunc, onClose ModelAction) ModelAction {
	return func(m *Model, ctx ActionContext) tea.Cmd {
		m.dismissOverlayKeepAnchor()

		var cmds []tea.Cmd

		if onClose != nil {
			if c := onClose(m, ctx); c != nil {
				cmds = append(cmds, c)
			}
		}

		if onSelect != nil {
			if act := onSelect(idx); act != nil {
				if c := act(m, ctx); c != nil {
					cmds = append(cmds, c)
				}
			}
		}

		return tea.Batch(cmds...)
	}
}

// popupDismiss は overlay を閉じる（選択なしで閉じた場合に使う）。
func popupDismiss(onClose ModelAction) ModelAction {
	return func(m *Model, ctx ActionContext) tea.Cmd {
		m.dismissOverlayKeepAnchor()

		if onClose == nil {
			return nil
		}

		return onClose(m, ctx)
	}
}
