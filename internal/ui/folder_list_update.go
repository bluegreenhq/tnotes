package ui

import tea "charm.land/bubbletea/v2"

// StartInput はインライン入力モードを開始する。
func (fl *FolderList) StartInput() {
	fl.inputMode = true
	fl.inputValue = ""
}

// CancelInput はインライン入力をキャンセルする。
func (fl *FolderList) CancelInput() {
	fl.inputMode = false
	fl.inputValue = ""
}

// InputValue は入力中のフォルダ名を返す。
func (fl *FolderList) InputValue() string { return fl.inputValue }

// HitTestHeader はヘッダー領域のクリック判定を行う.
// 戻り値: headerHitClose, headerHitAdd, headerHitMore, "" (該当なし).
func (fl *FolderList) HitTestHeader(x, y int) string {
	if y != 0 {
		return ""
	}

	contentWidth := fl.width - folderListBorderWidth

	// ✕ ボタン (左端)
	if x <= headerCloseBtnWidth {
		return headerHitClose
	}

	// ボタンは右端に配置
	if fl.IsUserFolder() {
		if x == contentWidth-headerMoreBtnOffset {
			return headerHitMore
		}

		if x == contentWidth-headerAddBtnOffsetMore {
			return headerHitAdd
		}
	} else if x == contentWidth-headerAddBtnOffsetNoMore {
		return headerHitAdd
	}

	return ""
}

// Update はメッセージに応じてフォルダ一覧の状態を更新する。
func (fl *FolderList) Update(msg tea.Msg) (FolderList, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return *fl, nil
	}

	// メニュー表示中
	if fl.menuOpen {
		if keyMsg.Code == tea.KeyEscape {
			fl.CloseMenu()

			return *fl, nil
		}

		fl.CloseMenu()
	}

	// インライン入力モード
	if fl.inputMode {
		return fl.updateInput(keyMsg)
	}

	return fl.handleKeyNav(keyMsg)
}

func (fl *FolderList) handleKeyNav(keyMsg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	if keyMsg.Mod&tea.ModCtrl != 0 {
		switch keyMsg.Code {
		case 'p':
			return fl.moveUp()
		case 'n':
			return fl.moveDown()
		}

		return *fl, nil
	}

	switch keyMsg.Code {
	case tea.KeyUp, 'k':
		return fl.moveUp()
	case tea.KeyDown, 'j':
		return fl.moveDown()
	case tea.KeyEnter, tea.KeyTab:
		return *fl, folderListCmd(FolderListFocusNext)
	}

	return *fl, nil
}

// SelectIndex はインデックスを指定して選択する。
func (fl *FolderList) SelectIndex(idx int) tea.Cmd {
	if idx < 0 || idx >= len(fl.folders) {
		return nil
	}

	prev := fl.selected
	fl.selected = idx

	if prev != fl.selected {
		return folderListCmd(FolderListSelect)
	}

	return nil
}

func (fl *FolderList) moveUp() (FolderList, tea.Cmd) {
	if fl.selected > 0 {
		fl.selected--

		return *fl, folderListCmd(FolderListSelect)
	}

	return *fl, nil
}

func (fl *FolderList) moveDown() (FolderList, tea.Cmd) {
	if fl.selected < len(fl.folders)-1 {
		fl.selected++

		return *fl, folderListCmd(FolderListSelect)
	}

	return *fl, nil
}

func (fl *FolderList) updateInput(keyMsg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	switch keyMsg.Code {
	case tea.KeyEnter:
		if fl.inputValue != "" {
			fl.inputMode = false
			name := fl.inputValue
			fl.inputValue = ""

			return *fl, func() tea.Msg {
				return folderCreateMsg{Name: name}
			}
		}

		return *fl, nil
	case tea.KeyEscape:
		fl.inputMode = false
		fl.inputValue = ""

		return *fl, nil
	case tea.KeyBackspace:
		if len(fl.inputValue) > 0 {
			runes := []rune(fl.inputValue)
			fl.inputValue = string(runes[:len(runes)-1])
		}

		return *fl, nil
	default:
		if keyMsg.Text != "" {
			fl.inputValue += keyMsg.Text
		}

		return *fl, nil
	}
}

// SetHeaderHover はヘッダーのホバー状態を更新する。
func (fl *FolderList) SetHeaderHover(x, y int) {
	fl.hoverClose = false
	fl.hoverAdd = false
	fl.hoverMore = false

	if y != 0 {
		return
	}

	hit := fl.HitTestHeader(x, y)

	switch hit {
	case headerHitClose:
		fl.hoverClose = true
	case headerHitAdd:
		fl.hoverAdd = true
	case headerHitMore:
		fl.hoverMore = true
	}
}

// ClearHeaderHover はヘッダーのホバーをすべて解除する。
func (fl *FolderList) ClearHeaderHover() {
	fl.hoverClose = false
	fl.hoverAdd = false
	fl.hoverMore = false
}

// HoverAdd は + ボタンがホバー中かを返す。
func (fl *FolderList) HoverAdd() bool { return fl.hoverAdd }

// HoverMore は ⋯ ボタンがホバー中かを返す。
func (fl *FolderList) HoverMore() bool { return fl.hoverMore }

func folderListCmd(msg FolderListMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}
