package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

// PaneContext は PaneComponent.UpdatePane に渡す共通コンテキスト。
type PaneContext struct {
	Now       time.Time
	TrashMode bool
}

// PaneComponent はメイン領域の pane（FolderList / NoteList / Editor）が実装するインターフェース。
// UpdatePane は (Model にしてほしい同期アクション, ランタイムに渡す非同期 Cmd) を返す。
// どちらか一方を返すケースが多いが、両方返してもよい。
type PaneComponent interface {
	UpdatePane(msg tea.Msg, ctx PaneContext) (ModelAction, tea.Cmd)
	ClearHover()
}

// BlinkHandler はカーソル blink 通知を処理するコンポーネントが実装するインターフェース。
type BlinkHandler interface {
	HandleBlinkMsg(msg tui.CursorBlinkMsg) tea.Cmd
}

// overlayComponent はオーバーレイ（ヘルプ / ポップアップメニュー / 確認ダイアログ等）が
// 実装する ui-package 内部インターフェース。UpdateOverlay は ModelAction を直接返すため
// shared.OverlayComponent.Update のように tea.Cmd 経由で actionCmd ラップする必要がない。
type overlayComponent interface {
	UpdateOverlay(msg tea.Msg) (ModelAction, tea.Cmd)
	RenderOn(base string, width, height int) string
	SetScreenSize(width, height int)
}

// popupOverlay は AnchoredPopupOverlay / FixedPopupOverlay が実装する共通インターフェース。
// オーバーレイ自身が「閉じ時の pane 側後始末」と「保持するメニュー」を知るため、
// Model 側は overlay 種別による分岐を持たずに済む。
type popupOverlay interface {
	overlayComponent
	Menu() *tui.PopupMenu
	OnClose() ModelAction
}
