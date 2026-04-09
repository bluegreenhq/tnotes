package ui

// HoverTarget はフッターボタンのホバーターゲット。
type HoverTarget int

const (
	HoverNone HoverTarget = iota
	HoverNew
	HoverQuit
	HoverRestore
	HoverTrashToggle
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
	popupMenu PopupMenu
	menuMsgs  []FooterMsg // menuItems[i] に対応する FooterMsg
}

// NewFooter は新しい Footer を生成する。
func NewFooter() Footer {
	return Footer{
		hover:     HoverNone,
		buttons:   nil,
		menuOpen:  false,
		popupMenu: PopupMenu{items: nil, hover: -1},
		menuMsgs:  nil,
	}
}

// Hover はホバーターゲットを返す。
func (f *Footer) Hover() HoverTarget { return f.hover }

// MenuOpen はメニューが開いているかを返す。
func (f *Footer) MenuOpen() bool { return f.menuOpen }

// PopupMenuRef はポップアップメニューへの参照を返す。
func (f *Footer) PopupMenuRef() *PopupMenu { return &f.popupMenu }

// FooterState はフッターのボタン構築に必要な状態。
type FooterState struct {
	TrashMode    bool
	TrashCount   int
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
	var menuItems []MenuItem

	if s.TrashMode {
		menuItems = []MenuItem{
			{Label: "Restore", Disabled: s.TrashCount == 0},
			{Label: "Notes", Disabled: false},
			{Label: "Quit", Disabled: false},
		}
		f.menuMsgs = []FooterMsg{FooterRestore, FooterTrashToggle, FooterQuit}
	} else {
		menuItems = []MenuItem{
			{Label: "New", Disabled: false},
			{Label: "Trash", Disabled: false},
			{Label: "Quit", Disabled: false},
		}
		f.menuMsgs = []FooterMsg{FooterNew, FooterTrashToggle, FooterQuit}
	}

	prevHover := f.popupMenu.hover
	f.popupMenu = NewPopupMenu(menuItems)
	f.popupMenu.hover = prevHover
}
