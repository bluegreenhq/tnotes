package ui

import tea "charm.land/bubbletea/v2"

// SetHover はホバーターゲットを設定する。
func (f *Footer) SetHover(h HoverTarget) { f.hover = h }

// SetButtons はフッターに表示するボタンリストを設定する。
func (f *Footer) SetButtons(buttons []FooterButton) { f.buttons = buttons }

// HitTest はフッター行のX座標からホバーターゲットを判定する。
func (f *Footer) HitTest(x int) HoverTarget {
	cursor := 1 // 先頭スペース分

	for i, btn := range f.buttons {
		if i > 0 {
			cursor += 2 // ボタン間スペース
		}

		end := cursor + len(btn.Label)
		if !btn.Disabled && x >= cursor && x < end {
			return btn.Target
		}

		cursor = end
	}

	return HoverNone
}

// HandleClick はフッター行のX座標からクリックされたボタンを判定し、対応する FooterMsg を返す。
func (f *Footer) HandleClick(x int) tea.Cmd {
	target := f.HitTest(x)

	switch target {
	case HoverNew:
		return footerCmd(FooterNew)
	case HoverRestore:
		return footerCmd(FooterRestore)
	case HoverTrashToggle:
		return footerCmd(FooterTrashToggle)
	case HoverQuit:
		return footerCmd(FooterQuit)
	case HoverCopy:
		return footerCmd(FooterCopy)
	case HoverCut:
		return footerCmd(FooterCut)
	case HoverNone:
		return nil
	}

	return nil
}

func footerCmd(msg FooterMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}
