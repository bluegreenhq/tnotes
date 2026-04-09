package ui

import tea "charm.land/bubbletea/v2"

// Update はメッセージに応じてフォルダ一覧の状態を更新する。
func (fl *FolderList) Update(msg tea.Msg) (FolderList, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return *fl, nil
	}

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

func folderListCmd(msg FolderListMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}
