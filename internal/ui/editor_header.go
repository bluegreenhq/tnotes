package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

const (
	// editorHeaderHeight はヘッダーの高さ（行数: ボタン行 + セパレーター行）。
	editorHeaderHeight = 2
	// editorHeaderMenuTopY はメニューの開始Y座標（セパレーター行に重ねる）。
	editorHeaderMenuTopY = editorHeaderHeight - 1
)

// EditorHeader はエディタ上部のヘッダーコンポーネント。
type EditorHeader struct {
	width         int
	menuOpen      bool
	PopupMenu     *tui.PopupMenu
	menuActions   []ModelAction
	moveAnchor    *menuAnchor // メニューが anchored で開かれた位置（Move サブメニューを同位置に展開するため）
	hoverNew      bool
	hoverMore     bool
	hasNote       bool
	hasContent    bool
	trashMode     bool
	pinned        bool
	moveMenuOpen  bool            // 移動先フォルダメニュー表示中
	MoveMenu      *tui.PopupMenu  // 移動先フォルダ一覧
	moveFolders   []string        // 移動先フォルダ名リスト
	searchInput   tui.LineInput   // 検索入力
	searchFocused bool            // 検索フィールドにフォーカスがあるか
	searchBlink   tui.CursorBlink // 検索カーソル点滅
	hoverSearch   bool            // 検索フィールドのホバー状態
}

// NewEditorHeader は新しい EditorHeader を生成する。
func NewEditorHeader(width int) *EditorHeader {
	return &EditorHeader{
		width:         width,
		menuOpen:      false,
		PopupMenu:     tui.NewPopupMenu(nil),
		menuActions:   nil,
		moveAnchor:    nil,
		hoverNew:      false,
		hoverMore:     false,
		hasNote:       false,
		hasContent:    false,
		trashMode:     false,
		pinned:        false,
		moveMenuOpen:  false,
		MoveMenu:      tui.NewPopupMenu(nil),
		moveFolders:   nil,
		searchInput:   tui.NewLineInput(),
		searchFocused: false,
		searchBlink:   tui.NewCursorBlink(blinkOwnerSearch),
		hoverSearch:   false,
	}
}

// Width はヘッダーの幅を返す。
func (h *EditorHeader) Width() int { return h.width }

// SetWidth はヘッダーの幅を設定する。
func (h *EditorHeader) SetWidth(w int) { h.width = w }

// MenuOpen はメニューが開いているかを返す。
func (h *EditorHeader) MenuOpen() bool { return h.menuOpen }

// HasNote はノートが選択されているかを返す。
func (h *EditorHeader) HasNote() bool { return h.hasNote }

// SetHasNote はノート選択状態を設定する。
func (h *EditorHeader) SetHasNote(v bool) { h.hasNote = v }

// TrashMode はゴミ箱モードかを返す。
func (h *EditorHeader) TrashMode() bool { return h.trashMode }

// SetHasContent はコンテンツの有無を設定する。
func (h *EditorHeader) SetHasContent(v bool) { h.hasContent = v }

// SetTrashMode はゴミ箱モードを設定する。
func (h *EditorHeader) SetTrashMode(v bool) { h.trashMode = v }

// SetPinned はピン留め状態を設定する。
func (h *EditorHeader) SetPinned(v bool) { h.pinned = v }

// MoveMenuOpen は移動先メニューが開いているかを返す。
func (h *EditorHeader) MoveMenuOpen() bool { return h.moveMenuOpen }

const searchFieldWidth = 20

// SearchFocused は検索フィールドにフォーカスがあるかを返す。
func (h *EditorHeader) SearchFocused() bool { return h.searchFocused }

// SetSearchFocused は検索フィールドのフォーカスを設定する。
func (h *EditorHeader) SetSearchFocused(v bool) { h.searchFocused = v }

// SearchQuery は検索テキストを返す。
func (h *EditorHeader) SearchQuery() string { return h.searchInput.Value() }

// ClearSearch は検索テキストをクリアする。
func (h *EditorHeader) ClearSearch() { h.searchInput.Reset() }

// HasSearchQuery は検索テキストがあるかを返す。
func (h *EditorHeader) HasSearchQuery() bool { return h.searchInput.Value() != "" }

// SearchBlinkVisible は検索カーソルの表示状態を返す。
func (h *EditorHeader) SearchBlinkVisible() bool { return h.searchBlink.Visible() }

// ResetSearchBlink は検索カーソルの blink をリセットする。
func (h *EditorHeader) ResetSearchBlink() tea.Cmd { return h.searchBlink.Reset() }

// HandleBlinkMsg は BlinkHandler インターフェース実装（検索カーソル用）。
func (h *EditorHeader) HandleBlinkMsg(msg tui.CursorBlinkMsg) tea.Cmd {
	return h.searchBlink.HandleMsg(msg)
}

// MenuLeftX はメニュー左端のヘッダー相対X座標を返す。
func (h *EditorHeader) MenuLeftX() int {
	return h.width - searchFieldWidth - h.PopupMenu.Width()
}

// MoveMenuLeftX は移動先メニュー左端のヘッダー相対X座標を返す。
func (h *EditorHeader) MoveMenuLeftX() int {
	return h.width - searchFieldWidth - moreButtonOffset + 1 - h.MoveMenu.Width()
}

// newButtonX は + ボタンの X 座標（左端スペース含む）。
const newButtonX = 1

// RebuildMenu はメニュー項目を現在の状態に基づいて再構築する。
func (h *EditorHeader) RebuildMenu() {
	var menuItems []tui.MenuItem

	if h.trashMode {
		menuItems = []tui.MenuItem{
			tui.NewMenuItem("Move to…"),
		}
		h.menuActions = []ModelAction{openNoteMoveMenu}
	} else {
		menuItems = []tui.MenuItem{
			tui.NewMenuItem("Delete Note"),
		}
		h.menuActions = []ModelAction{trashSelectedNote}

		if h.pinned {
			menuItems = append(menuItems, tui.NewMenuItem("Unpin Note"))
			h.menuActions = append(h.menuActions, unpinNote)
		} else {
			menuItems = append(menuItems, tui.NewMenuItem("Pin Note"))
			h.menuActions = append(h.menuActions, pinNote)
		}

		menuItems = append(menuItems, tui.NewMenuItem("Move to…"))
		h.menuActions = append(h.menuActions, openNoteMoveMenu)

		menuItems = append(menuItems, tui.NewMenuItem("Duplicate"))
		h.menuActions = append(h.menuActions, duplicateSelectedNote)

		if h.hasContent {
			menuItems = append(menuItems, tui.NewMenuItem("Copy Note"))
			h.menuActions = append(h.menuActions, copyNoteToClipboard)
		}
	}

	prevHover := h.PopupMenu.Hover()
	h.PopupMenu = tui.NewPopupMenu(menuItems)
	h.PopupMenu.SetHover(prevHover)
}

// OpenMenu は固定位置モードでメニューを開く。Move サブメニューも固定位置に展開される。
func (h *EditorHeader) OpenMenu() {
	h.moveAnchor = nil
	h.RebuildMenu()
	h.menuOpen = true
}

// OpenMenuAtAnchor はアンカー位置モードでメニューを開く。Move サブメニューも同じアンカー位置に展開される。
func (h *EditorHeader) OpenMenuAtAnchor(anchorX, anchorY int) {
	a := menuAnchor{x: anchorX, y: anchorY}
	h.moveAnchor = &a
	h.RebuildMenu()
	h.menuOpen = true
}

// MoveAnchor は Move サブメニュー展開時に使うアンカー位置を返す。固定位置モードでは nil。
func (h *EditorHeader) MoveAnchor() *menuAnchor { return h.moveAnchor }

// CloseMenu はメニューを閉じる。
// moveAnchor はクリアしない（popupSelect chain で onClose → onSelect(openMoveMenu) が
// 順次走るとき、openMoveMenu が anchor を必要とする）。次の OpenMenu/OpenMenuAtAnchor で
// 上書きされるため、メニュー open 中以外の値は意味を持たない。
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
func (h *EditorHeader) HandleClick(x int) ModelAction {
	// + ボタン判定
	if !h.trashMode && x == newButtonX {
		return createNote
	}

	// ⋯ ボタン判定（検索フィールドより優先）
	if h.hasNote && h.isMoreButtonX(x) {
		return openEditorHeaderMenu
	}

	// 検索フィールド判定
	h.handleSearchClick(x)

	return nil
}

// HandleMenuClick はメニュー領域のクリックを処理する。
// x, y はメニュー左上を原点とする相対座標。
func (h *EditorHeader) HandleMenuClick(x, y int) ModelAction {
	idx, hit := h.PopupMenu.HandleClick(x, y)
	h.CloseMenu()

	if !hit || idx < 0 || idx >= len(h.menuActions) {
		return nil
	}

	return h.menuActions[idx]
}

// ExecuteMenuAction はインデックスに対応するメニューアクションを返す。
func (h *EditorHeader) ExecuteMenuAction(idx int) ModelAction {
	if idx < 0 || idx >= len(h.menuActions) {
		return nil
	}

	return h.menuActions[idx]
}

// SetMenuHover はメニュー領域のホバーを更新する。
func (h *EditorHeader) SetMenuHover(x, y int) {
	h.PopupMenu.SetHoverByPos(x, y)
}

// SetHover は X 座標からホバー状態を更新する。
func (h *EditorHeader) SetHover(x int) {
	h.hoverNew = !h.trashMode && x == newButtonX
	h.hoverMore = h.hasNote && h.isMoreButtonX(x)
	h.setSearchHover(x)
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

// ExecuteMoveMenuAction はインデックスに対応する移動先フォルダのアクションを返す。
func (h *EditorHeader) ExecuteMoveMenuAction(idx int) ModelAction {
	if idx < 0 || idx >= len(h.moveFolders) {
		return nil
	}

	return moveNoteTo(h.moveFolders[idx])
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

// View はヘッダー行を描画する。
func (h *EditorHeader) View() string {
	showNew := !h.trashMode
	showMore := h.hasNote

	var left, right string

	if showNew {
		style := buttonStyle
		if h.hoverNew {
			style = buttonHoverStyle
		}

		left = " " + style.Render("+")
	}

	if showMore {
		style := buttonStyle
		if h.hoverMore {
			style = buttonHoverStyle
		}

		right = style.Render("⋯") + " "
	}

	searchField := h.renderSearchField()

	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)
	searchLen := lipgloss.Width(searchField)
	gap := h.width - leftLen - rightLen - searchLen

	gap = max(gap, 0)

	buttonLine := left + strings.Repeat(" ", gap) + right + searchField
	separator := strings.Repeat("─", max(h.width, 0))

	return buttonLine + "\n" + separator
}

const (
	searchFieldPadding = 4 // 左右スペース×2
	searchIconWidth    = 2 // "⚲" + space
)

func (h *EditorHeader) renderSearchField() string {
	icon := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("⚲")

	if h.searchFocused {
		content := h.searchInput.ViewWithWidth(searchFieldWidth-searchFieldPadding-searchIconWidth, h.searchBlink.Visible())
		inner := " " + icon + " " + content + " "

		return padOrTruncate(inner, searchFieldWidth)
	}

	if h.searchInput.Value() != "" {
		content := h.searchInput.ViewWithWidth(searchFieldWidth-searchFieldPadding-searchIconWidth, false)
		inner := " " + icon + " " + content + " "

		return padOrTruncate(inner, searchFieldWidth)
	}

	// プレースホルダー
	placeholder := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Search")
	inner := " " + icon + " " + placeholder + " "

	return padOrTruncate(inner, searchFieldWidth)
}

func padOrTruncate(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}

	return s + strings.Repeat(" ", width-w)
}

func (h *EditorHeader) isMoreButtonX(x int) bool {
	moreX := h.width - searchFieldWidth - moreButtonOffset

	return x == moreX
}

func (h *EditorHeader) handleSearchClick(x int) bool {
	searchStart := h.width - searchFieldWidth

	if x < searchStart || x >= h.width {
		return false
	}

	h.searchFocused = true

	return true
}

func (h *EditorHeader) setSearchHover(x int) {
	searchStart := h.width - searchFieldWidth
	h.hoverSearch = x >= searchStart && x < h.width
}

// --- EditorHeader → Model アクション ---

// pinNote はノートをピン留めする。
func pinNote(m *Model, _ ActionContext) tea.Cmd {
	return m.setNotePin(true)
}

// unpinNote はノートのピン留めを解除する。
func unpinNote(m *Model, _ ActionContext) tea.Cmd {
	return m.setNotePin(false)
}

// openNoteMoveMenu は移動先フォルダ選択メニューを開く。
func openNoteMoveMenu(m *Model, _ ActionContext) tea.Cmd {
	return m.openMoveMenu()
}

// openEditorHeaderMenu はエディタヘッダーの「⋯」メニューを固定位置ポップアップとして開く。
func openEditorHeaderMenu(m *Model, _ ActionContext) tea.Cmd {
	m.Editor.Header.OpenMenu()
	m.Overlays.OpenFixedPopup(m.Editor.Header.PopupMenu, m.editorHeaderMenuOrigin,
		m.Editor.Header.ExecuteMenuAction, closeEditorHeaderMenu)

	return nil
}

// closeEditorHeaderMenu はエディタヘッダーの「⋯」メニューを閉じる（popup overlay の onClose 用）。
func closeEditorHeaderMenu(m *Model, _ ActionContext) tea.Cmd {
	m.Editor.Header.CloseMenu()

	return nil
}

// closeMoveMenu は移動先メニューを閉じる（popup overlay の onClose 用）。
func closeMoveMenu(m *Model, _ ActionContext) tea.Cmd {
	m.Editor.Header.CloseMoveMenu()

	return nil
}

// moveNoteTo はノートを指定フォルダに移動し、UI を切り替える。
func moveNoteTo(dest string) ModelAction {
	return func(m *Model, ctx ActionContext) tea.Cmd {
		return m.handleNoteMove(dest, ctx.Now)
	}
}
