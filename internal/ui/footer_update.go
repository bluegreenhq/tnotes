package ui

import tea "charm.land/bubbletea/v2"

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
