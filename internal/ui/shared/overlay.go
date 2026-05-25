package shared

// OverlayComponent はオーバーレイの描画インターフェース。
// 入力（Update）の扱いはホストアプリ側で定義する（ModelAction を返す等）。
type OverlayComponent interface {
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
