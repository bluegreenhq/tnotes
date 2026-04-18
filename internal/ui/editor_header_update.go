package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

// newButtonX は + ボタンの X 座標（左端スペース含む）。
const newButtonX = 1

// RebuildMenu はメニュー項目を現在の状態に基づいて再構築する。
func (h *EditorHeader) RebuildMenu() {
	var menuItems []tui.MenuItem

	if h.trashMode {
		menuItems = []tui.MenuItem{
			tui.NewMenuItem("Move to…"),
		}
		h.menuMsgs = []EditorHeaderMsg{EditorHeaderMove}
	} else {
		menuItems = []tui.MenuItem{
			tui.NewMenuItem("Delete Note"),
		}
		h.menuMsgs = []EditorHeaderMsg{EditorHeaderTrash}

		if h.pinned {
			menuItems = append(menuItems, tui.NewMenuItem("Unpin Note"))
			h.menuMsgs = append(h.menuMsgs, EditorHeaderUnpin)
		} else {
			menuItems = append(menuItems, tui.NewMenuItem("Pin Note"))
			h.menuMsgs = append(h.menuMsgs, EditorHeaderPin)
		}

		menuItems = append(menuItems, tui.NewMenuItem("Move to…"))
		h.menuMsgs = append(h.menuMsgs, EditorHeaderMove)

		menuItems = append(menuItems, tui.NewMenuItem("Duplicate"))
		h.menuMsgs = append(h.menuMsgs, EditorHeaderDuplicate)

		if h.hasContent {
			menuItems = append(menuItems, tui.NewMenuItem("Copy Note"))
			h.menuMsgs = append(h.menuMsgs, EditorHeaderCopy)
		}
	}

	prevHover := h.PopupMenu.Hover()
	h.PopupMenu = tui.NewPopupMenu(menuItems)
	h.PopupMenu.SetHover(prevHover)
}

// OpenMenu はメニューを開く。
func (h *EditorHeader) OpenMenu() {
	h.RebuildMenu()
	h.menuOpen = true
}

// CloseMenu はメニューを閉じる。
func (h *EditorHeader) CloseMenu() {
	h.menuOpen = false
	h.PopupMenu.SetHover(-1)
	h.moveMenuOpen = false
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
		return EditorHeaderNew.Cmd()
	}

	// ⋯ ボタン判定（検索フィールドより優先）
	if h.hasNote && h.isMoreButtonX(x) {
		if h.menuOpen {
			h.CloseMenu()
		} else {
			h.OpenMenu()
		}

		return nil
	}

	// 検索フィールド判定
	focused, cleared := h.HandleSearchClick(x)
	if cleared {
		return func() tea.Msg { return searchClearedMsg{} }
	}

	if focused {
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

	return h.menuMsgs[idx].Cmd()
}

// ExecuteMenuAction はインデックスに対応するメニューアクションのコマンドを返す。
func (h *EditorHeader) ExecuteMenuAction(idx int) tea.Cmd {
	if idx < 0 || idx >= len(h.menuMsgs) {
		return nil
	}

	return h.menuMsgs[idx].Cmd()
}

// SetMenuHover はメニュー領域のホバーを更新する。
func (h *EditorHeader) SetMenuHover(x, y int) {
	h.PopupMenu.SetHoverByPos(x, y)
}

// SetHover は X 座標からホバー状態を更新する。
func (h *EditorHeader) SetHover(x int) {
	h.hoverNew = !h.trashMode && x == newButtonX
	h.hoverMore = h.hasNote && h.isMoreButtonX(x)
	h.SetSearchHover(x)
}

// ClearHover はホバーをすべて解除する。
func (h *EditorHeader) ClearHover() {
	h.hoverNew = false
	h.hoverMore = false
	h.hoverSearch = false
}

// HoverNew は + ボタンがホバー中かを返す。
func (h *EditorHeader) HoverNew() bool { return h.hoverNew }

// HoverMore は ⋯ ボタンがホバー中かを返す。
func (h *EditorHeader) HoverMore() bool { return h.hoverMore }

// moreButtonOffset は "⋯ " の "⋯" 位置（右端からのオフセット）。
const moreButtonOffset = 2

func (h *EditorHeader) isMoreButtonX(x int) bool {
	moreX := h.width - searchFieldWidth - moreButtonOffset

	return x == moreX
}

// OpenMoveMenu は移動先フォルダ選択メニューを開く。
func (h *EditorHeader) OpenMoveMenu(folders []string) {
	h.moveFolders = folders
	items := make([]tui.MenuItem, len(folders))

	for i, name := range folders {
		items[i] = tui.NewMenuItem(name)
	}

	h.MoveMenu = tui.NewPopupMenu(items)
	h.moveMenuOpen = true
	h.menuOpen = false
}

// CloseMoveMenu は移動先メニューを閉じる。
func (h *EditorHeader) CloseMoveMenu() {
	h.moveMenuOpen = false
	h.MoveMenu.SetHover(-1)
}

// HandleMoveMenuClick は移動先メニューのクリックを処理する。
func (h *EditorHeader) HandleMoveMenuClick(x, y int) tea.Cmd {
	idx, hit := h.MoveMenu.HandleClick(x, y)
	h.CloseMoveMenu()

	if !hit || idx < 0 || idx >= len(h.moveFolders) {
		return nil
	}

	dest := h.moveFolders[idx]

	return noteMoveMsg{DestFolder: dest}.Cmd()
}

// ExecuteMoveMenuAction はインデックスに対応する移動先フォルダのコマンドを返す。
func (h *EditorHeader) ExecuteMoveMenuAction(idx int) tea.Cmd {
	if idx < 0 || idx >= len(h.moveFolders) {
		return nil
	}

	dest := h.moveFolders[idx]

	return noteMoveMsg{DestFolder: dest}.Cmd()
}

// MoveMenuHeight は移動先メニューの高さを返す。
func (h *EditorHeader) MoveMenuHeight() int {
	if !h.moveMenuOpen {
		return 0
	}

	return h.MoveMenu.Height()
}

// SetMoveMenuHover は移動先メニューのホバーを更新する。
func (h *EditorHeader) SetMoveMenuHover(x, y int) {
	h.MoveMenu.SetHoverByPos(x, y)
}

// HandleSearchKey は検索フィールドのキー入力を処理する。
// 戻り値: (handled bool, cmd tea.Cmd).
func (h *EditorHeader) HandleSearchKey(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	if !h.searchFocused {
		return false, nil
	}

	switch msg.Code {
	case tea.KeyEscape, tea.KeyEnter:
		h.searchFocused = false

		return true, nil
	default:
		h.searchInput.HandleKey(msg)

		return true, nil
	}
}

// HandleSearchClick は検索フィールド領域のクリックを処理する。
// x はヘッダー内の相対X座標。
// 戻り値: focused=フォーカス取得, cleared=クリアボタン押下。
func (h *EditorHeader) HandleSearchClick(x int) (bool, bool) {
	searchStart := h.width - searchFieldWidth

	if x < searchStart || x >= h.width {
		return false, false
	}

	h.searchFocused = true

	return true, false
}

// SetSearchHover は検索フィールド領域のホバーを更新する。
func (h *EditorHeader) SetSearchHover(x int) {
	searchStart := h.width - searchFieldWidth
	h.hoverSearch = x >= searchStart && x < h.width
}
