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
	buttonTop := d.contentLines() - confirmButtonRows // ボタン3行の先頭行

	if relY < buttonTop || relY > buttonTop+2 {
		return ConfirmContinue
	}

	padLeft := d.buttonPadLeft()
	noStartX := padLeft + d.yesBtn.DisplayWidth() + confirmButtonGapCols

	if d.yesBtn.HitTest(relX, padLeft) {
		return ConfirmYes
	}

	if d.noBtn.HitTest(relX, noStartX) {
		return ConfirmNo
	}

	return ConfirmContinue
}

// HandleMotion は相対座標でのマウスホバーを処理する。
func (d *ConfirmDialog) HandleMotion(relX, relY int) {
	d.yesBtn.SetHovered(false)
	d.noBtn.SetHovered(false)

	buttonTop := d.contentLines() - confirmButtonRows

	if relY < buttonTop || relY > buttonTop+2 {
		return
	}

	padLeft := d.buttonPadLeft()
	noStartX := padLeft + d.yesBtn.DisplayWidth() + confirmButtonGapCols

	if d.yesBtn.HitTest(relX, padLeft) {
		d.yesBtn.SetHovered(true)
	} else if d.noBtn.HitTest(relX, noStartX) {
		d.noBtn.SetHovered(true)
	}
}

// ClearHover はホバー状態をリセットする。
func (d *ConfirmDialog) ClearHover() {
	d.yesBtn.SetHovered(false)
	d.noBtn.SetHovered(false)
}
