package ui

import "github.com/bluegreenhq/dogubako/tui"

// HelpItem はショートカット1件を表す。
type HelpItem struct {
	Key         string
	Description string
}

// HelpSection はショートカットのグループを表す。
type HelpSection struct {
	Title string
	Items []HelpItem
}

// HelpOverlay はショートカット一覧オーバーレイコンポーネント。
type HelpOverlay struct {
	sections    []HelpSection
	closeHover  bool
	screenWidth int // 画面幅
	bodyHeight  int // ボディ領域の高さ
}

// NewHelpOverlay はフォーカスに応じた HelpOverlay を生成する。
func NewHelpOverlay(focus FocusArea) *HelpOverlay {
	var sections []HelpSection

	switch focus {
	case FocusNoteList:
		sections = append(sections, noteListHelpSection())
	case FocusFolderList:
		sections = append(sections, folderListHelpSection())
	case FocusEditor:
		sections = append(sections, editorHelpSection())
	}

	sections = append(sections, globalHelpSection())

	return &HelpOverlay{sections: sections, closeHover: false, screenWidth: 0, bodyHeight: 0}
}

// SetScreenSize は画面サイズを設定する。
func (h *HelpOverlay) SetScreenSize(screenWidth, bodyHeight int) {
	h.screenWidth = screenWidth
	h.bodyHeight = bodyHeight
}

// Geometry はオーバーレイの画面上の配置情報を返す。
func (h *HelpOverlay) Geometry() tui.OverlayGeometry {
	const (
		borderW = 1
		padLeft = helpOverlayPaddingH
		padTop  = 1
	)

	return tui.CalcOverlayGeometry(h.View(), h.screenWidth, h.bodyHeight, borderW, padLeft, padTop)
}

// CloseButtonHit は✕ボタンがクリック/ホバーされたかを判定する。
func (h *HelpOverlay) CloseButtonHit(absX, absY int) bool {
	g := h.Geometry()
	if g.OverlayW == 0 {
		return false
	}

	btnY := g.StartY + helpCloseBtnRow
	btnX := g.StartX + g.OverlayW - 3 //nolint:mnd // border右(1) + padding右(1) の内側

	return absX == btnX && absY == btnY
}

func noteListHelpSection() HelpSection {
	return HelpSection{
		Title: "Note List",
		Items: []HelpItem{
			{"j/↓/Ctrl+N", "Move down"},
			{"k/↑/Ctrl+P", "Move up"},
			{"n", "New note"},
			{"d/Del", "Delete note"},
			{"m", "Menu"},
			{"Enter", "Edit note"},
			{"Ctrl+C", "Copy note"},
			{"Ctrl+D", "Duplicate note"},
			{"Ctrl+Z", "Undo"},
			{"Ctrl+Shift+Z", "Redo"},
			{"q", "Quit"},
		},
	}
}

func folderListHelpSection() HelpSection {
	return HelpSection{
		Title: "Folder List",
		Items: []HelpItem{
			{"j/↓/Ctrl+N", "Move down"},
			{"k/↑/Ctrl+P", "Move up"},
			{"m", "Menu"},
			{"Enter/Tab", "Next pane"},
			{"q", "Quit"},
		},
	}
}

func editorHelpSection() HelpSection {
	return HelpSection{
		Title: "Editor",
		Items: []HelpItem{
			{"Ctrl+S", "Save"},
			{"Ctrl+Z", "Undo"},
			{"Ctrl+Shift+Z", "Redo"},
			{"Ctrl+Shift+A", "Select all"},
			{"Ctrl+C", "Copy"},
			{"Ctrl+X", "Cut"},
			{"Ctrl+V", "Paste"},
			{"Ctrl+O", "Open URL"},
			{"Ctrl+A", "Beginning of line"},
			{"Ctrl+E", "End of line"},
			{"Ctrl+F", "Move right"},
			{"Ctrl+B", "Move left"},
			{"Ctrl+N", "Move down"},
			{"Ctrl+P", "Move up"},
			{"Ctrl+D", "Delete character"},
			{"Ctrl+K", "Kill to end of line"},
			{"Ctrl+Y", "Yank"},
			{"Shift+Arrow", "Select"},
			{"Esc/Tab", "Back to list"},
		},
	}
}

func globalHelpSection() HelpSection {
	return HelpSection{
		Title: "Global",
		Items: []HelpItem{
			{"Ctrl+Shift+F", "Search"},
			{"Ctrl+Shift+4", "Copy screen"},
			{"Ctrl+Q", "Quit"},
			{"Ctrl+B", "Toggle folders"},
			{"Tab", "Next pane"},
		},
	}
}
