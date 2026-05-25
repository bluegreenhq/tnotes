package shared

// OverlayComponent はオーバーレイの描画インターフェース。
// 入力（Update）の扱いはホストアプリ側で定義する（ModelAction を返す等）。
type OverlayComponent interface {
	RenderOn(base string, width, height int) string
	SetScreenSize(width, height int)
}
