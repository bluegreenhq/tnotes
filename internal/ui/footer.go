package ui

// HoverTarget はフッターボタンのホバーターゲット。
type HoverTarget int

const (
	HoverNone HoverTarget = iota
	HoverQuit
	HoverCopy
	HoverCut
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
	PopupMenu *PopupMenu
	menuMsgs  []FooterMsg // menuItems[i] に対応する FooterMsg
}

// NewFooter は新しい Footer を生成する。
func NewFooter() Footer {
	return Footer{
		hover:     HoverNone,
		buttons:   nil,
		menuOpen:  false,
		PopupMenu: NewPopupMenu(nil),
		menuMsgs:  nil,
	}
}

// Hover はホバーターゲットを返す。
func (f *Footer) Hover() HoverTarget { return f.hover }

// MenuOpen はメニューが開いているかを返す。
func (f *Footer) MenuOpen() bool { return f.menuOpen }

// FooterState はフッターのボタン構築に必要な状態。
type FooterState struct {
	TrashMode    bool
	HasSelection bool
	EditorDirty  bool
}

// RebuildButtons はフッターのボタンリストを再構築する。
func (f *Footer) RebuildButtons(s FooterState) {
	f.buttons = []FooterButton{
		NewFooterButton("Menu", HoverMore),
	}

	if !s.TrashMode {
		if s.HasSelection {
			f.buttons = append(f.buttons,
				NewFooterButton("Copy", HoverCopy),
				NewFooterButton("Cut", HoverCut),
			)
		} else if s.EditorDirty {
			f.buttons = append(f.buttons,
				FooterButton{Label: "● Modified", Target: HoverNone, Disabled: true},
			)
		}
	}

	// メニュー項目を構築
	var menuItems = []MenuItem{
		{Label: "Quit", Disabled: false},
	}

	f.menuMsgs = []FooterMsg{FooterQuit}

	prevHover := f.PopupMenu.hover
	f.PopupMenu = NewPopupMenu(menuItems)
	f.PopupMenu.hover = prevHover
}
