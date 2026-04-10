package ui

import tea "charm.land/bubbletea/v2"

// Update はキー入力に応じてダイアログの状態を更新する。
func (d *ConfirmDialog) Update(msg tea.Msg) ConfirmResult {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return ConfirmContinue
	}

	switch keyMsg.Code {
	case 'y', 'Y', tea.KeyEnter:
		return ConfirmYes
	case 'n', 'N', tea.KeyEscape:
		return ConfirmNo
	}

	return ConfirmContinue
}

// HandleClick は相対座標でのクリックを処理する。
// relY はダイアログコンテンツ内の行番号。
func (d *ConfirmDialog) HandleClick(relX, relY int) ConfirmResult {
	if relY != d.contentLines()-1 {
		return ConfirmContinue
	}

	if relX < len(d.confirmBtn) {
		return ConfirmYes
	}

	return ConfirmNo
}

// HandleMotion は相対座標でのマウスホバーを処理する。
func (d *ConfirmDialog) HandleMotion(relX, relY int) {
	d.hoverYes = false
	d.hoverNo = false

	if relY != d.contentLines()-1 {
		return
	}

	if relX < len(d.confirmBtn) {
		d.hoverYes = true
	} else {
		d.hoverNo = true
	}
}

// ClearHover はホバー状態をリセットする。
func (d *ConfirmDialog) ClearHover() {
	d.hoverYes = false
	d.hoverNo = false
}
