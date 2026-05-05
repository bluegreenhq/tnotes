package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/ui/shared"
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

var _ shared.OverlayComponent = (*HelpOverlay)(nil)

// helpCloseBtnRow は✕ボタンを配置する行（0始まり）。border上(0) の1つ下 = paddingTop行。
const helpCloseBtnRow = 1

const (
	helpOverlayWidth    = 42
	helpOverlayPaddingH = 2
	helpKeyMinWidth     = 15
)

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

// Update はメッセージに応じて状態を更新し、副作用 Cmd を返す。
func (h *HelpOverlay) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return h.handleKey(msg)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft && h.closeButtonHit(msg.X, msg.Y) {
			return func() tea.Msg { return HelpOverlayCloseMsg{} }
		}
	case tea.MouseMsg:
		mouse := msg.Mouse()
		h.closeHover = h.closeButtonHit(mouse.X, mouse.Y)
	}

	return nil
}

// RenderOn はベース画面上にヘルプオーバーレイを合成する。
func (h *HelpOverlay) RenderOn(base string, _, _ int) string {
	bodyLines := strings.Split(base, "\n")
	overlayLines := strings.Split(h.View(), "\n")
	g := h.geometry()
	tui.OverlayLines(bodyLines, overlayLines, g.StartX, g.StartY)

	return strings.Join(bodyLines, "\n")
}

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

// geometry はオーバーレイの画面上の配置情報を返す。
func (h *HelpOverlay) geometry() tui.OverlayGeometry {
	const (
		borderW = 1
		padLeft = helpOverlayPaddingH
		padTop  = 1
	)

	return tui.CalcOverlayGeometry(h.View(), h.screenWidth, h.bodyHeight, borderW, padLeft, padTop)
}

// closeButtonHit は✕ボタンがクリック/ホバーされたかを判定する。
func (h *HelpOverlay) closeButtonHit(absX, absY int) bool {
	g := h.geometry()
	if g.OverlayW == 0 {
		return false
	}

	btnY := g.StartY + helpCloseBtnRow
	btnX := g.StartX + g.OverlayW - 3 //nolint:mnd // border右(1) + padding右(1) の内側

	return absX == btnX && absY == btnY
}

func (h *HelpOverlay) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case msg.Code == 'q' && msg.Mod&tea.ModCtrl != 0:
		return func() tea.Msg { return HelpOverlayQuitMsg{} }
	case msg.Code == tea.KeyEscape:
		return func() tea.Msg { return HelpOverlayCloseMsg{} }
	case msg.Code == '?' && msg.Mod == 0:
		return func() tea.Msg { return HelpOverlayCloseMsg{} }
	case msg.Code == '/' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		return func() tea.Msg { return HelpOverlayCloseMsg{} }
	}

	return nil
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
