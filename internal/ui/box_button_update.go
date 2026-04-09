package ui

// HitTest は startX からの相対位置でボタン範囲内かを判定する。
func (b *BoxButton) HitTest(x, startX int) bool {
	return x >= startX && x < startX+b.DisplayWidth()
}
