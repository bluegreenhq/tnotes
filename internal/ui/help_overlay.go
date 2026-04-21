package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

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

// HelpResult はヘルプオーバーレイの操作結果を表す。
type HelpResult int

const (
	// HelpContinue はオーバーレイ継続中。
	HelpContinue HelpResult = iota
	// HelpClose はオーバーレイを閉じる。
	HelpClose
	// HelpQuit はアプリケーション終了を要求する。
	HelpQuit
)

// SetCloseHover は閉じるボタンのホバー状態を設定する。
func (h *HelpOverlay) SetCloseHover(hovered bool) {
	h.closeHover = hovered
}

// Update はキー入力に応じてオーバーレイの状態を更新する。
func (h *HelpOverlay) Update(msg tea.Msg) HelpResult {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return HelpContinue
	}

	switch {
	case keyMsg.Code == 'q' && keyMsg.Mod&tea.ModCtrl != 0:
		return HelpQuit
	case keyMsg.Code == tea.KeyEscape:
		return HelpClose
	case keyMsg.Code == '?' && keyMsg.Mod == 0:
		return HelpClose
	case keyMsg.Code == '/' && keyMsg.Mod == (tea.ModCtrl|tea.ModShift):
		return HelpClose
	}

	return HelpContinue
}

// helpCloseBtnRow は✕ボタンを配置する行（0始まり）。border上(0) の1つ下 = paddingTop行。
const helpCloseBtnRow = 1

const (
	helpOverlayWidth    = 42
	helpOverlayPaddingH = 2
	helpKeyMinWidth     = 15
)

// View はオーバーレイの描画内容を返す。
func (h *HelpOverlay) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	b.WriteString(titleStyle.Render("Shortcuts"))
	b.WriteString("\n")

	sectionTitleStyle := lipgloss.NewStyle().Bold(true)
	hintStyle := lipgloss.NewStyle().Faint(true)

	for i, section := range h.sections {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(sectionTitleStyle.Render(section.Title))
		b.WriteString("\n")

		keyWidth := h.keyWidth(section)

		for _, item := range section.Items {
			padded := item.Key + strings.Repeat(" ", keyWidth-lipgloss.Width(item.Key))
			b.WriteString(padded + item.Description + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("Press Esc/?/Ctrl+Shift+/ to close"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("12")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(helpOverlayPaddingH).
		PaddingRight(helpOverlayPaddingH).
		Width(helpOverlayWidth)

	rendered := boxStyle.Render(b.String())

	return h.overlayCloseButton(rendered)
}

// overlayCloseButton はレンダリング済みオーバーレイの右上（paddingTop行、border右の直前）に✕を重ねる。
func (h *HelpOverlay) overlayCloseButton(rendered string) string {
	lines := strings.Split(rendered, "\n")
	if len(lines) <= helpCloseBtnRow {
		return rendered
	}

	style := buttonStyle
	if h.closeHover {
		style = buttonHoverStyle
	}

	closeStr := style.Render("✕")
	lineWidth := lipgloss.Width(lines[helpCloseBtnRow])
	// border右(1文字) の直前に✕を配置
	btnX := lineWidth - 3 //nolint:mnd // border右(1) + padding右(1) の内側

	lines[helpCloseBtnRow] = tui.ComposeLine(lines[helpCloseBtnRow], closeStr, btnX, 1)

	return strings.Join(lines, "\n")
}

// keyWidth はセクション内のキー列の表示幅を返す。
func (h *HelpOverlay) keyWidth(section HelpSection) int {
	maxLen := 0

	for _, item := range section.Items {
		w := lipgloss.Width(item.Key)
		if w > maxLen {
			maxLen = w
		}
	}

	const keyGap = 2

	return max(maxLen+keyGap, helpKeyMinWidth)
}
