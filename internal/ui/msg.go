package ui

// editorContextMsg はエディタコンテキストメニューのアクション。
type editorContextMsg int

const (
	editorContextCopy editorContextMsg = iota
	editorContextCut
	editorContextPaste
)

// searchDebounceMsg はデバウンスタイマー発火を表すメッセージ。
type searchDebounceMsg struct {
	id    int
	query string
}

// searchClearedMsg は検索テキストがクリアされたことを表すメッセージ。
type searchClearedMsg struct{}

// PopupKind はポップアップメニューの種類を表す。
type PopupKind int

const (
	// PopupKindNone は不明/未指定。
	PopupKindNone PopupKind = iota
	// PopupKindEditorContext はエディタの右クリックコンテキストメニュー。
	PopupKindEditorContext
	// PopupKindEditorHeader はエディタヘッダーの「…」メニュー。
	PopupKindEditorHeader
	// PopupKindMoveMenu はエディタヘッダーの移動先メニュー。
	PopupKindMoveMenu
	// PopupKindFolderList はフォルダ一覧の moreメニュー。
	PopupKindFolderList
	// PopupKindFooter はフッターのメニュー。
	PopupKindFooter
)
