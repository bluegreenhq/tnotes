package ui

const (
	// editorHeaderHeight はヘッダーの高さ（行数: ボタン行 + セパレーター行）。
	editorHeaderHeight = 2
	// editorHeaderMenuTopY はメニューの開始Y座標（セパレーター行に重ねる）。
	editorHeaderMenuTopY = editorHeaderHeight - 1
)

// EditorHeader はエディタ上部のヘッダーコンポーネント。
type EditorHeader struct {
	width        int
	menuOpen     bool
	PopupMenu    *PopupMenu
	menuMsgs     []EditorHeaderMsg
	hoverNew     bool
	hoverMore    bool
	hasNote      bool
	hasContent   bool
	trashMode    bool
	pinned       bool
	moveMenuOpen bool       // 移動先フォルダメニュー表示中
	MoveMenu     *PopupMenu // 移動先フォルダ一覧
	moveFolders  []string   // 移動先フォルダ名リスト
}

// NewEditorHeader は新しい EditorHeader を生成する。
func NewEditorHeader(width int) *EditorHeader {
	return &EditorHeader{
		width:        width,
		menuOpen:     false,
		PopupMenu:    NewPopupMenu(nil),
		menuMsgs:     nil,
		hoverNew:     false,
		hoverMore:    false,
		hasNote:      false,
		hasContent:   false,
		trashMode:    false,
		pinned:       false,
		moveMenuOpen: false,
		MoveMenu:     NewPopupMenu(nil),
		moveFolders:  nil,
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
