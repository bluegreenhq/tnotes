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
type PaneComponent interface {
	UpdatePane(msg tea.Msg, ctx PaneContext) tea.Cmd
	ClearHover()
}

// BlinkHandler はカーソル blink 通知を処理するコンポーネントが実装するインターフェース。
type BlinkHandler interface {
	HandleBlinkMsg(msg tui.CursorBlinkMsg) tea.Cmd
}
