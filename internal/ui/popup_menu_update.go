package ui

// HitTest は座標からメニュー項目のインデックスを返す。
// メニュー左上を (0,0) とする相対座標。
// 項目間には空行があり、偶数行(1,3,5...)が項目、奇数行(2,4,...)が空行。
// 項目外、空行、または disabled の場合は -1 を返す。
func (m *PopupMenu) HitTest(x, y int) int {
	if x < 0 || x >= m.Width() {
		return -1
	}

	// y=0: 上枠, y=Height()-1: 下枠
	if y <= 0 || y >= m.Height()-1 {
		return -1
	}

	// 枠の内側の行番号 (1始まり)
	innerY := y - 1

	// 偶数番目(0,2,4...) が項目行、奇数番目(1,3,5...) が空行
	if innerY%menuBorderLines != 0 {
		return -1 // 空行
	}

	idx := innerY / menuBorderLines

	if idx < 0 || idx >= len(m.items) {
		return -1
	}

	if m.items[idx].Disabled {
		return -1
	}

	return idx
}

// SetHoverByPos はマウス座標からホバー状態を更新する。
// メニュー左上を (0,0) とする相対座標。
func (m *PopupMenu) SetHoverByPos(x, y int) {
	m.hover = m.HitTest(x, y)
}

// HandleClick はクリック座標から選択された項目インデックスを返す。
// メニュー左上を (0,0) とする相対座標。
// 戻り値は (インデックス, ヒットしたか)。
func (m *PopupMenu) HandleClick(x, y int) (int, bool) {
	idx := m.HitTest(x, y)
	if idx < 0 {
		return -1, false
	}

	return idx, true
}
