package ui

import tea "charm.land/bubbletea/v2"

// StartInput はインライン入力モードを開始する（新規作成用）。
func (fl *FolderList) StartInput() tea.Cmd {
	fl.inputMode = true
	fl.lineInput.Reset()

	return fl.blink.Reset()
}

// CommitInput はインライン入力を確定する。値が空なら破棄する。
func (fl *FolderList) CommitInput() tea.Cmd {
	val := fl.lineInput.Value()
	fl.clearInput()

	if val != "" {
		return folderCreateMsg{Name: val}.Cmd()
	}

	return nil
}

// CancelInput はインライン入力を破棄する（Esc用）。
func (fl *FolderList) CancelInput() {
	fl.clearInput()
}

// StartRename はリネーム入力モードを開始する。
func (fl *FolderList) StartRename() tea.Cmd {
	name := fl.SelectedName()
	fl.renameMode = true
	fl.renameName = name
	fl.lineInput.SetValue(name)

	return fl.blink.Reset()
}

// CommitRename はリネーム入力を確定する。値が空または変更なしなら破棄する。
func (fl *FolderList) CommitRename() tea.Cmd {
	val := fl.lineInput.Value()
	oldName := fl.renameName
	fl.clearRename()

	if val != "" && val != oldName {
		return folderRenameMsg{OldName: oldName, NewName: val}.Cmd()
	}

	return nil
}

// CancelRename はリネーム入力を破棄する（Esc用）。
func (fl *FolderList) CancelRename() {
	fl.clearRename()
}

func (fl *FolderList) clearInput() {
	fl.inputMode = false
	fl.lineInput.Reset()
	fl.blink.Stop()
}

func (fl *FolderList) clearRename() {
	fl.renameMode = false
	fl.renameName = ""
	fl.lineInput.Reset()
	fl.blink.Stop()
}

// InputValue は入力中のフォルダ名を返す。
func (fl *FolderList) InputValue() string { return fl.lineInput.Value() }

// HitTestHeader はヘッダー領域のクリック判定を行う.
// 戻り値: headerHitClose, headerHitAdd, "" (該当なし).
func (fl *FolderList) HitTestHeader(x, y int) string {
	if y != 0 {
		return ""
	}

	contentWidth := max(fl.width-folderListBorderWidth, 0)

	// ✕ ボタン (左端)
	if x <= headerCloseBtnWidth {
		return headerHitClose
	}

	// + ボタン (右端)
	if x == contentWidth-headerAddBtnOffset {
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

	// リネーム入力モード
	if fl.renameMode {
		return fl.updateRename(keyMsg)
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
	case 'm':
		return *fl, FolderListMenu.Cmd()
	case tea.KeyEnter, tea.KeyTab:
		return *fl, FolderListFocusNext.Cmd()
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
		return FolderListSelect.Cmd()
	}

	return nil
}

func (fl *FolderList) moveUp() (FolderList, tea.Cmd) {
	if fl.selected > 0 {
		fl.selected--

		return *fl, FolderListSelect.Cmd()
	}

	return *fl, nil
}

func (fl *FolderList) moveDown() (FolderList, tea.Cmd) {
	if fl.selected < len(fl.folders)-1 {
		fl.selected++

		return *fl, FolderListSelect.Cmd()
	}

	return *fl, nil
}

func (fl *FolderList) updateRename(keyMsg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	result := fl.lineInput.handleKey(keyMsg)

	switch result {
	case lineInputNone:
		// blink reset は model_update 側で行う
	case lineInputSubmit:
		return *fl, fl.CommitRename()
	case lineInputCancel:
		fl.CancelRename()

		return *fl, nil
	}

	return *fl, nil
}

func (fl *FolderList) updateInput(keyMsg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	result := fl.lineInput.handleKey(keyMsg)

	switch result {
	case lineInputNone:
		// blink reset は model_update 側で行う
	case lineInputSubmit:
		return *fl, fl.CommitInput()
	case lineInputCancel:
		fl.CancelInput()

		return *fl, nil
	}

	return *fl, nil
}

// SetHeaderHover はヘッダーのホバー状態を更新する。
func (fl *FolderList) SetHeaderHover(x, y int) {
	fl.hoverClose = false
	fl.hoverAdd = false

	if y != 0 {
		return
	}

	hit := fl.HitTestHeader(x, y)

	switch hit {
	case headerHitClose:
		fl.hoverClose = true
	case headerHitAdd:
		fl.hoverAdd = true
	}
}

// ClearHeaderHover はヘッダーのホバーをすべて解除する。
func (fl *FolderList) ClearHeaderHover() {
	fl.hoverClose = false
	fl.hoverAdd = false
}

// HoverAdd は + ボタンがホバー中かを返す。
func (fl *FolderList) HoverAdd() bool { return fl.hoverAdd }
