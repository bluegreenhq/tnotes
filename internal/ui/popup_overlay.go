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

// PopupOverlay は popup メニューを画面にオーバーレイ表示する。
// 「固定位置（origin closure 指定）」と「アンカー位置（座標指定 + 画面内クランプ）」の
// 両モードを単一構造体で扱う。anchor が nil ならば固定位置モード。
type PopupOverlay struct {
	menu         *tui.PopupMenu
	origin       func() (int, int) // クランプ済みのメニュー左上座標を返す
	anchor       *menuAnchor       // anchored の場合のみセット（dismiss 時の anchor 保存用）
	onSelect     popupSelectFunc
	onClose      ModelAction
	screenWidth  int
	screenHeight int
}

var _ shared.OverlayComponent = (*PopupOverlay)(nil)

// NewFixedPopupOverlay は origin closure で位置が決まる popup を生成する。
// editor header の「⋯」ボタンや folder list の more ボタンから開く用途。
func NewFixedPopupOverlay(
	menu *tui.PopupMenu,
	origin func() (int, int),
	onSelect popupSelectFunc,
	onClose ModelAction,
) *PopupOverlay {
	return &PopupOverlay{
		menu:         menu,
		origin:       origin,
		anchor:       nil,
		onSelect:     onSelect,
		onClose:      onClose,
		screenWidth:  0,
		screenHeight: 0,
	}
}

// NewAnchoredPopupOverlay は anchor (x, y) を起点に画面内クランプして表示する popup を生成する。
// 右クリックメニュー用途。dismiss 時には anchor を保存する（サブメニュー復元用）。
func NewAnchoredPopupOverlay(
	menu *tui.PopupMenu,
	anchorX, anchorY int,
	onSelect popupSelectFunc,
	onClose ModelAction,
) *PopupOverlay {
	a := menuAnchor{x: anchorX, y: anchorY}
	p := &PopupOverlay{
		menu:         menu,
		origin:       nil, // セット後すぐ下で代入
		anchor:       &a,
		onSelect:     onSelect,
		onClose:      onClose,
		screenWidth:  0,
		screenHeight: 0,
	}
	p.origin = p.anchoredOrigin

	return p
}

// Menu は内部の PopupMenu を返す。
func (p *PopupOverlay) Menu() *tui.PopupMenu { return p.menu }

// OnClose は閉じ時の pane 側後始末アクションを返す。
func (p *PopupOverlay) OnClose() ModelAction { return p.onClose }

// Anchor はアンカー情報を返す（fixed モードでは nil）。
func (p *PopupOverlay) Anchor() *menuAnchor { return p.anchor }

// SetScreenSize は画面サイズを設定する。
func (p *PopupOverlay) SetScreenSize(width, height int) {
	p.screenWidth = width
	p.screenHeight = height
}

// UpdateOverlay はメッセージに応じて状態を更新し、ModelAction と tea.Cmd を返す。
func (p *PopupOverlay) UpdateOverlay(msg tea.Msg) (ModelAction, tea.Cmd) {
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
func (p *PopupOverlay) RenderOn(base string, _, _ int) string {
	bodyLines := strings.Split(base, "\n")
	menuLines := p.menu.View()

	if len(menuLines) == 0 {
		return base
	}

	ox, oy := p.origin()
	tui.OverlayLines(bodyLines, menuLines, ox, oy)

	return strings.Join(bodyLines, "\n")
}

// anchoredOrigin は anchor 座標から画面サイズに収まるメニュー左上座標を計算する。
func (p *PopupOverlay) anchoredOrigin() (int, int) {
	menuLines := p.menu.View()
	if len(menuLines) == 0 {
		return 0, 0
	}

	w := lipgloss.Width(menuLines[0])
	h := len(menuLines)

	return tui.ClampMenuOrigin(w, h, p.anchor.x, p.anchor.y, p.screenWidth, p.screenHeight)
}

func (p *PopupOverlay) handleKey(msg tea.KeyPressMsg) ModelAction {
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

func (p *PopupOverlay) handleClick(msg tea.MouseClickMsg) ModelAction {
	if msg.Button != tea.MouseLeft {
		return popupDismiss(p.onClose)
	}

	ox, oy := p.origin()
	w := p.menu.Width()
	h := p.menu.Height()

	// メニュー領域外のクリックは閉じる
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

// --- PopupOverlay → Model アクション ---

// popupSelect は overlay を閉じてから onClose と onSelect(idx) を順に適用する。
func popupSelect(idx int, onSelect popupSelectFunc, onClose ModelAction) ModelAction {
	return func(m *Model, ctx ActionContext) tea.Cmd {
		m.Overlays.DismissKeepAnchor()

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
		m.Overlays.DismissKeepAnchor()

		if onClose == nil {
			return nil
		}

		return onClose(m, ctx)
	}
}
