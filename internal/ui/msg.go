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
