package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// menuKind は開いているメニューの種類を表す。
type menuKind int

const (
	menuKindNone menuKind = iota
	menuKindEditorHeader
	menuKindMoveMenu
	menuKindFolderList
	menuKindEditorContext
	menuKindFooter
)

// menuAnchor は右クリック時のメニュー表示位置を表す。
type menuAnchor struct {
	x int // 画面絶対座標
	y int
}

// popupEntry はコーディネータに登録された個々のメニュー情報。
type popupEntry struct {
	kind              menuKind
	menu              func() *PopupMenu
	isOpen            func() bool
	close             func()
	execute           func(idx int, now time.Time) tea.Cmd
	handleAnchorClick func(relX, relY int) tea.Cmd // nil = アンカー非対応
}

// PopupCoordinator は複数コンポーネントに散在するポップアップメニューを一元管理する。
type PopupCoordinator struct {
	anchor  *menuAnchor
	entries []popupEntry
	layout  *Layout
}

// Active は現在開いているポップアップメニューとその種類を返す。なければ nil。
func (c *PopupCoordinator) Active() (*PopupMenu, menuKind) {
	for i := range c.entries {
		if c.entries[i].isOpen() {
			return c.entries[i].menu(), c.entries[i].kind
		}
	}

	return nil, menuKindNone
}

// CloseAll は全てのメニューとアンカーを閉じる。
func (c *PopupCoordinator) CloseAll() {
	c.anchor = nil

	for i := range c.entries {
		c.entries[i].close()
	}
}

// HandleKey はメニュー表示中のキー入力を処理する。
func (c *PopupCoordinator) HandleKey(msg tea.KeyPressMsg, menu *PopupMenu, kind menuKind, now time.Time) tea.Cmd {
	if msg.Code == tea.KeyEscape {
		c.CloseAll()

		return nil
	}

	if msg.Code == tea.KeyEnter {
		idx := menu.SelectHover()

		c.CloseAll()

		if idx < 0 {
			return nil
		}

		return c.executeAction(idx, kind, now)
	}

	menu.HandleKeyNav(msg)

	return nil
}

// HandleAnchoredClick はアンカー付きメニューのクリックを処理する。
func (c *PopupCoordinator) HandleAnchoredClick(msg tea.MouseClickMsg) tea.Cmd {
	menu := c.anchoredMenu()
	if menu == nil {
		c.anchor = nil

		return nil
	}

	x, y := c.anchoredMenuOrigin(menu)
	relX := msg.X - x
	relY := msg.Y - y

	var cmd tea.Cmd

	for i := range c.entries {
		if c.entries[i].isOpen() && c.entries[i].handleAnchorClick != nil {
			cmd = c.entries[i].handleAnchorClick(relX, relY)

			break
		}
	}

	c.anchor = nil

	return cmd
}

// HandleHover はアンカー付きメニューのホバーを更新する。
func (c *PopupCoordinator) HandleHover(mouse tea.Mouse) {
	if c.anchor == nil {
		return
	}

	menu := c.anchoredMenu()
	if menu == nil {
		return
	}

	x, y := c.anchoredMenuOrigin(menu)
	menu.SetHoverByPos(mouse.X-x, mouse.Y-y)
}

// SetAnchor はメニューのアンカー位置を設定する。
func (c *PopupCoordinator) SetAnchor(x, y int) {
	c.anchor = &menuAnchor{x: x, y: y}
}

// HasAnchor はアンカーが設定されているかを返す。
func (c *PopupCoordinator) HasAnchor() bool {
	return c.anchor != nil
}

// Anchor はアンカーを返す。
func (c *PopupCoordinator) Anchor() *menuAnchor {
	return c.anchor
}

// ClampAnchor はアンカー座標を画面内にクランプする。overlayAtAnchor と同じロジック。
func (c *PopupCoordinator) ClampAnchor(anchor *menuAnchor, menuWidth, menuHeight int) (int, int) {
	bodyHeight := c.layout.BodyHeight()

	x := anchor.x
	y := anchor.y

	if x+menuWidth > c.layout.width {
		x = c.layout.width - menuWidth
	}

	if x < 0 {
		x = 0
	}

	if y+menuHeight > bodyHeight {
		y = bodyHeight - menuHeight
	}

	if y < 0 {
		y = 0
	}

	return x, y
}

// AnchoredMenuOrigin はアンカー付きメニューのクランプ済み描画左上座標を返す。
func (c *PopupCoordinator) AnchoredMenuOrigin(menu *PopupMenu) (int, int) {
	menuLines := menu.View()
	if len(menuLines) == 0 {
		return c.anchor.x, c.anchor.y
	}

	return c.ClampAnchor(c.anchor, lipgloss.Width(menuLines[0]), len(menuLines))
}

func (c *PopupCoordinator) executeAction(idx int, kind menuKind, now time.Time) tea.Cmd {
	for i := range c.entries {
		if c.entries[i].kind == kind {
			return c.entries[i].execute(idx, now)
		}
	}

	return nil
}

func (c *PopupCoordinator) anchoredMenu() *PopupMenu {
	for i := range c.entries {
		if c.entries[i].isOpen() && c.entries[i].handleAnchorClick != nil {
			return c.entries[i].menu()
		}
	}

	return nil
}

func (c *PopupCoordinator) anchoredMenuOrigin(menu *PopupMenu) (int, int) {
	menuLines := menu.View()
	if len(menuLines) == 0 {
		return c.anchor.x, c.anchor.y
	}

	return c.ClampAnchor(c.anchor, lipgloss.Width(menuLines[0]), len(menuLines))
}
