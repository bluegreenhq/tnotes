package ui

import tea "charm.land/bubbletea/v2"

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
	PopupMenu     *PopupMenu
	menuMsgs      []EditorHeaderMsg
	hoverNew      bool
	hoverMore     bool
	hasNote       bool
	hasContent    bool
	trashMode     bool
	pinned        bool
	moveMenuOpen  bool        // 移動先フォルダメニュー表示中
	MoveMenu      *PopupMenu  // 移動先フォルダ一覧
	moveFolders   []string    // 移動先フォルダ名リスト
	searchInput   lineInput   // 検索入力
	searchFocused bool        // 検索フィールドにフォーカスがあるか
	searchBlink   cursorBlink // 検索カーソル点滅
	hoverSearch   bool        // 検索フィールドのホバー状態
}

// NewEditorHeader は新しい EditorHeader を生成する。
func NewEditorHeader(width int) *EditorHeader {
	return &EditorHeader{
		width:         width,
		menuOpen:      false,
		PopupMenu:     NewPopupMenu(nil),
		menuMsgs:      nil,
		hoverNew:      false,
		hoverMore:     false,
		hasNote:       false,
		hasContent:    false,
		trashMode:     false,
		pinned:        false,
		moveMenuOpen:  false,
		MoveMenu:      NewPopupMenu(nil),
		moveFolders:   nil,
		searchInput:   newLineInput(),
		searchFocused: false,
		searchBlink:   newCursorBlink(blinkOwnerSearch),
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

// MenuLeftX はメニュー左端のヘッダー相対X座標を返す。
func (h *EditorHeader) MenuLeftX() int {
	return h.width - searchFieldWidth - h.PopupMenu.Width()
}

// MoveMenuLeftX は移動先メニュー左端のヘッダー相対X座標を返す。
func (h *EditorHeader) MoveMenuLeftX() int {
	return h.width - searchFieldWidth - moreButtonOffset + 1 - h.MoveMenu.Width()
}
