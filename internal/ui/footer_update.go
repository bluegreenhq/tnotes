package ui

import tea "charm.land/bubbletea/v2"

// SetHover はホバーターゲットを設定する。
func (f *Footer) SetHover(h HoverTarget) { f.hover = h }

// SetButtons はフッターに表示するボタンリストを設定する。
func (f *Footer) SetButtons(buttons []FooterButton) { f.buttons = buttons }

// OpenMenu はメニューを開く。
func (f *Footer) OpenMenu() { f.menuOpen = true }

// CloseMenu はメニューを閉じる。
func (f *Footer) CloseMenu() {
	f.menuOpen = false
	f.PopupMenu.hover = -1
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
			bb := NewBoxButton(btn.Label)
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
	case HoverNew:
		return footerCmd(FooterNew)
	case HoverRestore:
		return footerCmd(FooterRestore)
	case HoverQuit:
		return footerCmd(FooterQuit)
	case HoverCopy:
		return footerCmd(FooterCopy)
	case HoverCut:
		return footerCmd(FooterCut)
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

	return footerCmd(f.menuMsgs[idx])
}

// SetMenuHover はメニュー領域のホバーを更新する。
// x, y はメニュー左上を原点とする相対座標。
func (f *Footer) SetMenuHover(x, y int) {
	f.PopupMenu.SetHoverByPos(x, y)
}

func footerCmd(msg FooterMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}
