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

func newMenuAnchor(x, y int) *menuAnchor {
	return &menuAnchor{x: x, y: y}
}

// popupEntry はコーディネータに登録された個々のメニュー情報。
type popupEntry struct {
	kind              menuKind
	menu              func() *PopupMenu
	isOpen            func() bool
	close             func()
	execute           func(idx int, now time.Time) tea.Cmd
	handleAnchorClick func(relX, relY int) tea.Cmd // nil = アンカー非対応
	origin            func() (x, y int)            // nil = 固定位置なし（アンカーのみ）
}

// PopupCoordinator は複数コンポーネントに散在するポップアップメニューを一元管理する。
type PopupCoordinator struct {
	anchor     *menuAnchor
	lastAnchor *menuAnchor // 直前のアンカー（サブメニュー復元用）
	entries    []popupEntry
	layout     *Layout
}

// NewPopupCoordinator は新しい PopupCoordinator を生成する。
func NewPopupCoordinator(layout *Layout, entries []popupEntry) PopupCoordinator {
	return PopupCoordinator{
		anchor:     nil,
		lastAnchor: nil,
		entries:    entries,
		layout:     layout,
	}
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

		// サブメニュー復元用にアンカーを保存してからクリア
		c.lastAnchor = c.anchor
		c.CloseAll()

		if idx < 0 {
			c.lastAnchor = nil

			return nil
		}

		cmd := c.executeAction(idx, kind, now)

		// ハンドラが TakeLastAnchor で消費しなかった場合、stale 防止のためクリア
		c.lastAnchor = nil

		return cmd
	}

	menu.HandleKeyNav(msg)

	return nil
}

// HandleFixedClick は固定位置メニュー（origin コールバック付き）のクリックを処理する。
// メニュー領域内ならクリックを処理し、メニュー外なら閉じる。
// 処理した場合は (cmd, true)、開いている固定位置メニューがなければ (nil, false)。
func (c *PopupCoordinator) HandleFixedClick(msg tea.MouseClickMsg, now time.Time) (tea.Cmd, bool) {
	for i := range c.entries {
		if !c.entries[i].isOpen() || c.entries[i].origin == nil {
			continue
		}

		menu := c.entries[i].menu()
		ox, oy := c.entries[i].origin()
		w := menu.Width()
		h := menu.Height()

		if msg.X >= ox && msg.X < ox+w && msg.Y >= oy && msg.Y < oy+h {
			relX := msg.X - ox
			relY := msg.Y - oy
			idx, hit := menu.HandleClick(relX, relY)

			c.lastAnchor = c.anchor
			c.CloseAll()

			if !hit || idx < 0 {
				c.lastAnchor = nil

				return nil, true
			}

			cmd := c.executeAction(idx, c.entries[i].kind, now)
			c.lastAnchor = nil

			return cmd, true
		}

		// メニュー外クリック → 閉じる
		c.entries[i].close()

		return nil, true
	}

	return nil, false
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

	// lastAnchor をハンドラ呼び出し前に保存する。
	// ハンドラ内で TakeLastAnchor() → SetAnchor() によりサブメニューを同じ位置に表示できるようにするため。
	c.lastAnchor = c.anchor
	c.anchor = nil

	var cmd tea.Cmd

	for i := range c.entries {
		if c.entries[i].isOpen() && c.entries[i].handleAnchorClick != nil {
			cmd = c.entries[i].handleAnchorClick(relX, relY)

			break
		}
	}

	// ハンドラが TakeLastAnchor で消費しなかった場合、stale 防止のためクリア
	c.lastAnchor = nil

	return cmd
}

// HandleHover はアンカー付きまたは固定位置メニューのホバーを更新する。
// ホバーを処理した場合は true を返す。
func (c *PopupCoordinator) HandleHover(mouse tea.Mouse) bool {
	if c.anchor != nil {
		menu := c.anchoredMenu()
		if menu != nil {
			x, y := c.anchoredMenuOrigin(menu)
			menu.SetHoverByPos(mouse.X-x, mouse.Y-y)

			return true
		}
	}

	for i := range c.entries {
		if !c.entries[i].isOpen() || c.entries[i].origin == nil {
			continue
		}

		menu := c.entries[i].menu()
		ox, oy := c.entries[i].origin()
		menu.SetHoverByPos(mouse.X-ox, mouse.Y-oy)

		return true
	}

	return false
}

// SetAnchor はメニューのアンカー位置を設定する。
func (c *PopupCoordinator) SetAnchor(x, y int) {
	c.anchor = newMenuAnchor(x, y)
}

// HasAnchor はアンカーが設定されているかを返す。
func (c *PopupCoordinator) HasAnchor() bool {
	return c.anchor != nil
}

// Anchor はアンカーを返す。
func (c *PopupCoordinator) Anchor() *menuAnchor {
	return c.anchor
}

// TakeLastAnchor は直前のアンカーを返し、消費する。サブメニュー復元用。
func (c *PopupCoordinator) TakeLastAnchor() *menuAnchor {
	a := c.lastAnchor
	c.lastAnchor = nil

	return a
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
