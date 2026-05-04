package ui

import tea "charm.land/bubbletea/v2"

// OverlayComponent はオーバーレイコンポーネントの共通インターフェース。
// Help / PopupMenu / ConfirmDialog などモーダル UI が実装する。
type OverlayComponent interface {
	Update(msg tea.Msg) tea.Cmd
	RenderOn(base string, width, height int) string
	SetScreenSize(width, height int)
}

// AnchoredOverlay はアンカー座標を持つオーバーレイ。
// 右クリックメニューなどクリック位置に追従するメニューが実装する。
type AnchoredOverlay interface {
	AnchorX() int
	AnchorY() int
	// MenuOrigin はクランプ済みのメニュー描画左上座標を返す。
	MenuOrigin(screenWidth, screenHeight int) (int, int)
}
