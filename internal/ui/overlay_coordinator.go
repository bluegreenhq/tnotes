package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

// OverlayCoordinator は表示中の overlay と画面サイズを一元管理する。
// Model から overlay 制御の責務を分離する。
type OverlayCoordinator struct {
	overlay      overlayComponent
	screenWidth  int
	screenHeight int
}

// NewOverlayCoordinator は新しい OverlayCoordinator を生成する。
func NewOverlayCoordinator() OverlayCoordinator {
	return OverlayCoordinator{
		overlay:      nil,
		screenWidth:  0,
		screenHeight: 0,
	}
}

// SetScreenSize は overlay の画面サイズを設定する。表示中の overlay にも反映する。
func (c *OverlayCoordinator) SetScreenSize(width, height int) {
	c.screenWidth = width
	c.screenHeight = height

	if c.overlay != nil {
		c.overlay.SetScreenSize(width, height)
	}
}

// IsOpen は何らかの overlay が表示中かを返す。
func (c *OverlayCoordinator) IsOpen() bool { return c.overlay != nil }

// HelpVisible はヘルプ overlay が表示中かを返す。
func (c *OverlayCoordinator) HelpVisible() bool {
	_, ok := c.overlay.(*HelpOverlay)

	return ok
}

// Menu は popup overlay の menu を返す。popup でなければ nil。テスト用ヘルパー。
func (c *OverlayCoordinator) Menu() *tui.PopupMenu {
	if pop, ok := c.overlay.(*PopupOverlay); ok {
		return pop.Menu()
	}

	return nil
}

// OpenHelp はヘルプ overlay を開く。
func (c *OverlayCoordinator) OpenHelp(focus FocusArea) {
	c.setOverlay(NewHelpOverlay(focus))
}

// OpenAnchoredPopup はアンカー付き popup overlay を開く。
func (c *OverlayCoordinator) OpenAnchoredPopup(
	menu *tui.PopupMenu,
	anchorX, anchorY int,
	onSelect popupSelectFunc,
	onClose ModelAction,
) {
	c.setOverlay(NewAnchoredPopupOverlay(menu, anchorX, anchorY, onSelect, onClose))
}

// OpenFixedPopup は固定位置 popup overlay を開く。
func (c *OverlayCoordinator) OpenFixedPopup(
	menu *tui.PopupMenu,
	origin func() (int, int),
	onSelect popupSelectFunc,
	onClose ModelAction,
) {
	c.setOverlay(NewFixedPopupOverlay(menu, origin, onSelect, onClose))
}

// OpenConfirmDeleteFolder はフォルダ削除確認ダイアログを開く。
func (c *OverlayCoordinator) OpenConfirmDeleteFolder(name string, noteCount int) {
	c.setOverlay(NewConfirmDeleteFolderDialog(name, noteCount))
}

// Dismiss は次に別の overlay を開く前の片付け。
// popup overlay の場合は onClose を呼んで pane 側の状態も整える。
func (c *OverlayCoordinator) Dismiss(m *Model) {
	if pop, ok := c.overlay.(*PopupOverlay); ok {
		if oc := pop.OnClose(); oc != nil {
			_ = oc(m, ActionContext{Now: time.Now()})
		}
	}

	c.overlay = nil
}

// Clear は overlay を強制的にクリアする（pane 側の cleanup は呼ばない）。
// popup overlay の選択/Esc 経路や、pane 自身が menu 状態を整えた上で overlay も
// 閉じたいケースで使う。
func (c *OverlayCoordinator) Clear() {
	c.overlay = nil
}

// Update は overlay にメッセージを委譲する。overlay が無ければ (nil, nil)。
func (c *OverlayCoordinator) Update(msg tea.Msg) (ModelAction, tea.Cmd) {
	if c.overlay == nil {
		return nil, nil
	}

	return c.overlay.UpdateOverlay(msg)
}

// RenderOn は overlay をベース画面に合成する。overlay が無ければ base をそのまま返す。
func (c *OverlayCoordinator) RenderOn(base string) string {
	if c.overlay == nil {
		return base
	}

	return c.overlay.RenderOn(base, c.screenWidth, c.screenHeight)
}

func (c *OverlayCoordinator) setOverlay(ov overlayComponent) {
	ov.SetScreenSize(c.screenWidth, c.screenHeight)
	c.overlay = ov
}
