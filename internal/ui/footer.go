package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

// HoverTarget はフッターボタンのホバーターゲット。
type HoverTarget int

const (
	HoverNone HoverTarget = iota
	HoverQuit
	HoverMore
)

// FooterButton はフッターに表示する1つのボタンを表す。
type FooterButton struct {
	Label    string
	Target   HoverTarget
	Disabled bool
}

// NewFooterButton は新しい FooterButton を生成する。
func NewFooterButton(label string, target HoverTarget) FooterButton {
	return FooterButton{Label: label, Target: target, Disabled: false}
}

// Footer はフッターバーの状態を表す。
type Footer struct {
	hover     HoverTarget
	buttons   []FooterButton
	menuOpen  bool
	PopupMenu *tui.PopupMenu
	menuMsgs  []FooterMsg // menuItems[i] に対応する FooterMsg
}

// NewFooter は新しい Footer を生成する。
func NewFooter() Footer {
	return Footer{
		hover:     HoverNone,
		buttons:   nil,
		menuOpen:  false,
		PopupMenu: tui.NewPopupMenu(nil),
		menuMsgs:  nil,
	}
}

// MenuOpen はメニューが開いているかを返す。
func (f *Footer) MenuOpen() bool { return f.menuOpen }

// RebuildButtons はフッターのボタンリストを再構築する。
func (f *Footer) RebuildButtons() {
	f.buttons = []FooterButton{
		NewFooterButton("Menu", HoverMore),
	}

	// メニュー項目を構築
	menuItems := []tui.MenuItem{
		tui.NewMenuItem("Shortcuts"),
		tui.NewMenuItem("Quit"),
	}

	f.menuMsgs = []FooterMsg{FooterHelp, FooterQuit}

	prevHover := f.PopupMenu.Hover()
	f.PopupMenu = tui.NewPopupMenu(menuItems)
	f.PopupMenu.SetHover(prevHover)
}

// SetHover はホバーターゲットを設定する。
func (f *Footer) SetHover(h HoverTarget) { f.hover = h }

// SetButtons はフッターに表示するボタンリストを設定する。
func (f *Footer) SetButtons(buttons []FooterButton) { f.buttons = buttons }

// OpenMenu はメニューを開く。
func (f *Footer) OpenMenu() { f.menuOpen = true }

// CloseMenu はメニューを閉じる。
func (f *Footer) CloseMenu() {
	f.menuOpen = false
	f.PopupMenu.SetHover(-1)
}

// MenuHeight はメニューが開いている場合のメニュー部分の高さを返す。
func (f *Footer) MenuHeight() int {
	if !f.menuOpen {
		return 0
	}

	return f.PopupMenu.Height()
}

// HitTest はフッター行のX座標からホバーターゲットを判定する。
func (f *Footer) HitTest(x int) HoverTarget {
	cursor := 1 // 先頭スペース分

	for i, btn := range f.buttons {
		if i > 0 {
			cursor += 2 // ボタン間スペース
		}

		var w int
		if btn.Target == HoverNone && btn.Disabled {
			// disabled な非ボタン（● Modified）はラベル幅のみ
			w = len(btn.Label)
		} else {
			bb := tui.NewBoxButton(btn.Label)
			w = bb.DisplayWidth()
		}

		end := cursor + w
		if !btn.Disabled && x >= cursor && x < end {
			return btn.Target
		}

		cursor = end
	}

	return HoverNone
}

// HandleClick はフッター行のクリックを処理する。
// [More] クリックでメニュー開閉をトグルし、他のボタンはコマンドを返す。
func (f *Footer) HandleClick(x int) tea.Cmd {
	target := f.HitTest(x)

	if target == HoverMore {
		if f.menuOpen {
			f.CloseMenu()
		} else {
			f.OpenMenu()
		}

		return nil
	}

	switch target {
	case HoverQuit:
		return FooterQuit.Cmd()
	case HoverNone, HoverMore:
		return nil
	}

	return nil
}

// HandleMenuClick はメニュー領域のクリックを処理する。
// x, y はメニュー左上を原点とする相対座標。
func (f *Footer) HandleMenuClick(x, y int) tea.Cmd {
	idx, hit := f.PopupMenu.HandleClick(x, y)
	f.CloseMenu()

	if !hit || idx < 0 || idx >= len(f.menuMsgs) {
		return nil
	}

	return f.menuMsgs[idx].Cmd()
}

// ExecuteMenuAction はインデックスに対応するメニューアクションのコマンドを返す。
func (f *Footer) ExecuteMenuAction(idx int) tea.Cmd {
	if idx < 0 || idx >= len(f.menuMsgs) {
		return nil
	}

	return f.menuMsgs[idx].Cmd()
}

// SetMenuHover はメニュー領域のホバーを更新する。
// x, y はメニュー左上を原点とする相対座標。
func (f *Footer) SetMenuHover(x, y int) {
	f.PopupMenu.SetHoverByPos(x, y)
}

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)

const footerLineCount = 3

// View はフッター行を描画する。infoMsgがあればボタンの右にシアン色で表示する。
// errMsgがある場合はエラーが優先される。
// 戻り値は描画文字列と行数。
func (f *Footer) View(errMsg, infoMsg string, width int) (string, int) {
	if errMsg != "" {
		f.menuOpen = false

		return renderErrorLines(errMsg, width), errorLineCount(errMsg, width)
	}

	btns := f.collectBoxButtons()
	topLine := renderFooterTopLine(btns)
	midLine := renderFooterMidLine(btns, infoMsg)
	botLine := renderFooterBotLine(btns)

	return topLine + "\n" + midLine + "\n" + botLine, footerLineCount
}

func (f *Footer) collectBoxButtons() []tui.BoxButton {
	btns := make([]tui.BoxButton, 0, len(f.buttons))

	for _, btn := range f.buttons {
		bb := tui.NewBoxButton(btn.Label)
		bb.SetHovered(f.hover == btn.Target)
		btns = append(btns, bb)
	}

	return btns
}

func renderFooterTopLine(btns []tui.BoxButton) string {
	var buf strings.Builder

	buf.WriteString(" ")

	for i := range btns {
		if i > 0 {
			buf.WriteString("  ")
		}

		buf.WriteString(btns[i].ViewTop())
	}

	return buf.String()
}

func renderFooterMidLine(btns []tui.BoxButton, infoMsg string) string {
	var buf strings.Builder

	buf.WriteString(" ")

	for i := range btns {
		if i > 0 {
			buf.WriteString("  ")
		}

		buf.WriteString(btns[i].ViewMiddle())
	}

	if infoMsg != "" {
		buf.WriteString("  ")
		buf.WriteString(infoStyle.Render(infoMsg))
	}

	return buf.String()
}

func renderFooterBotLine(btns []tui.BoxButton) string {
	var buf strings.Builder

	buf.WriteString(" ")

	for i := range btns {
		if i > 0 {
			buf.WriteString("  ")
		}

		buf.WriteString(btns[i].ViewBottom())
	}

	return buf.String()
}

func errorLineCount(msg string, width int) int {
	contentWidth := width - 1 // 先頭スペース分
	if contentWidth <= 0 {
		return 1
	}

	runes := []rune(msg)
	lines := (len(runes) + contentWidth - 1) / contentWidth
	lines = max(lines, 1)

	return lines
}

func renderErrorLines(msg string, width int) string {
	contentWidth := width - 1
	if contentWidth <= 0 || len([]rune(msg)) <= contentWidth {
		return " " + errorStyle.Render(msg)
	}

	runes := []rune(msg)

	var buf strings.Builder

	for i := 0; i < len(runes); i += contentWidth {
		end := min(i+contentWidth, len(runes))
		line := " " + errorStyle.Render(string(runes[i:end]))

		if i > 0 {
			buf.WriteString("\n")
		}

		buf.WriteString(line)
	}

	return buf.String()
}
