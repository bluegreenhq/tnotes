package ui

import tea "charm.land/bubbletea/v2"

// HelpResult はヘルプオーバーレイの操作結果を表す。
type HelpResult int

const (
	// HelpContinue はオーバーレイ継続中。
	HelpContinue HelpResult = iota
	// HelpClose はオーバーレイを閉じる。
	HelpClose
	// HelpQuit はアプリケーション終了を要求する。
	HelpQuit
)

// SetCloseHover は閉じるボタンのホバー状態を設定する。
func (h *HelpOverlay) SetCloseHover(hovered bool) {
	h.closeHover = hovered
}

// Update はキー入力に応じてオーバーレイの状態を更新する。
func (h *HelpOverlay) Update(msg tea.Msg) HelpResult {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return HelpContinue
	}

	switch {
	case keyMsg.Code == 'q' && keyMsg.Mod&tea.ModCtrl != 0:
		return HelpQuit
	case keyMsg.Code == tea.KeyEscape:
		return HelpClose
	case keyMsg.Code == '?' && keyMsg.Mod == 0:
		return HelpClose
	case keyMsg.Code == '/' && keyMsg.Mod == (tea.ModCtrl|tea.ModShift):
		return HelpClose
	}

	return HelpContinue
}
