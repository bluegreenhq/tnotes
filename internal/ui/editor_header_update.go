package ui

import tea "charm.land/bubbletea/v2"

// newButtonX は + ボタンの X 座標（左端スペース含む）。
const newButtonX = 1

// RebuildMenu はメニュー項目を現在の状態に基づいて再構築する。
func (h *EditorHeader) RebuildMenu() {
	var menuItems []MenuItem

	if h.trashMode {
		menuItems = []MenuItem{
			{Label: "Restore", Disabled: false},
		}
		h.menuMsgs = []EditorHeaderMsg{EditorHeaderRestore}
	} else {
		menuItems = []MenuItem{
			{Label: "Delete Note", Disabled: false},
		}
		h.menuMsgs = []EditorHeaderMsg{EditorHeaderTrash}

		if h.hasContent {
			menuItems = append(menuItems, MenuItem{Label: "Copy Note", Disabled: false})
			h.menuMsgs = append(h.menuMsgs, EditorHeaderCopy)
		}
	}

	prevHover := h.PopupMenu.hover
	h.PopupMenu = NewPopupMenu(menuItems)
	h.PopupMenu.hover = prevHover
}

// OpenMenu はメニューを開く。
func (h *EditorHeader) OpenMenu() {
	h.RebuildMenu()
	h.menuOpen = true
}

// CloseMenu はメニューを閉じる。
func (h *EditorHeader) CloseMenu() {
	h.menuOpen = false
	h.PopupMenu.hover = -1
}

// MenuHeight はメニューの高さを返す。
func (h *EditorHeader) MenuHeight() int {
	if !h.menuOpen {
		return 0
	}

	return h.PopupMenu.Height()
}

// HandleClick はヘッダー行のクリックを処理する。
// x はヘッダー内の相対 X 座標。
func (h *EditorHeader) HandleClick(x int) tea.Cmd {
	// + ボタン判定
	if !h.trashMode && x == newButtonX {
		return editorHeaderCmd(EditorHeaderNew)
	}

	// ⋯ ボタン判定
	if h.hasNote && h.isMoreButtonX(x) {
		if h.menuOpen {
			h.CloseMenu()
		} else {
			h.OpenMenu()
		}

		return nil
	}

	return nil
}

// HandleMenuClick はメニュー領域のクリックを処理する。
// x, y はメニュー左上を原点とする相対座標。
func (h *EditorHeader) HandleMenuClick(x, y int) tea.Cmd {
	idx, hit := h.PopupMenu.HandleClick(x, y)
	h.CloseMenu()

	if !hit || idx < 0 || idx >= len(h.menuMsgs) {
		return nil
	}

	return editorHeaderCmd(h.menuMsgs[idx])
}

// SetMenuHover はメニュー領域のホバーを更新する。
func (h *EditorHeader) SetMenuHover(x, y int) {
	h.PopupMenu.SetHoverByPos(x, y)
}

// SetHover は X 座標からホバー状態を更新する。
func (h *EditorHeader) SetHover(x int) {
	h.hoverNew = !h.trashMode && x == newButtonX
	h.hoverMore = h.hasNote && h.isMoreButtonX(x)
}

// ClearHover はホバーをすべて解除する。
func (h *EditorHeader) ClearHover() {
	h.hoverNew = false
	h.hoverMore = false
}

// HoverNew は + ボタンがホバー中かを返す。
func (h *EditorHeader) HoverNew() bool { return h.hoverNew }

// HoverMore は ⋯ ボタンがホバー中かを返す。
func (h *EditorHeader) HoverMore() bool { return h.hoverMore }

// moreButtonOffset は "⋯ " の "⋯" 位置（右端からのオフセット）。
const moreButtonOffset = 2

func (h *EditorHeader) isMoreButtonX(x int) bool {
	moreX := h.width - moreButtonOffset

	return x == moreX
}

func editorHeaderCmd(msg EditorHeaderMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}
