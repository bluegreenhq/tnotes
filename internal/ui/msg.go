package ui

// NoteListMsg はノート一覧からモデルへの通知メッセージ。
type NoteListMsg int

const (
	// NoteListSelect はカーソル移動によりノート選択が変わったことを通知する。
	NoteListSelect NoteListMsg = iota
	// NoteListCreate は新規ノート作成を要求する。
	NoteListCreate
	// NoteListTrash はノートのゴミ箱移動を要求する。
	NoteListTrash
	// NoteListEnterTrash はゴミ箱モードへの切り替えを要求する。
	NoteListEnterTrash
	// NoteListExitTrash はゴミ箱モードからの復帰を要求する。
	NoteListExitTrash
	// NoteListRestore はゴミ箱からのノート復元を要求する。
	NoteListRestore
	// NoteListUndo はundo操作を要求する。
	NoteListUndo
	// NoteListRedo はredo操作を要求する。
	NoteListRedo
	// NoteListEdit はエディタへのフォーカス切り替えを要求する。
	NoteListEdit
	// NoteListCopy はノート内容のクリップボードコピーを要求する。
	NoteListCopy
	// NoteListQuit は終了を要求する。
	NoteListQuit
)

// cursorBlinkMsg はカーソルの点滅状態を切り替えるメッセージ。
type cursorBlinkMsg struct {
	tag int
}

// EditorMsg はエディタからモデルへの通知メッセージ。
type EditorMsg int

const (
	// EditorBlur はノート一覧へのフォーカス切り替えを要求する。
	EditorBlur EditorMsg = iota
	// EditorSave はノート保存を要求する。
	EditorSave
)

// FooterMsg はフッターからモデルへの通知メッセージ。
type FooterMsg int

const (
	// FooterNew は新規ノート生成ボタンがクリックされたことを通知する。
	FooterNew FooterMsg = iota
	// FooterTrashToggle はゴミ箱モード切り替えボタンがクリックされたことを通知する。
	FooterTrashToggle
	// FooterRestore は復元ボタンがクリックされたことを通知する。
	FooterRestore
	// FooterQuit は終了ボタンがクリックされたことを通知する。
	FooterQuit
	// FooterCopy はコピーボタンがクリックされたことを通知する。
	FooterCopy
	// FooterCut はカットボタンがクリックされたことを通知する。
	FooterCut
	// FooterMore はMoreボタンがクリックされたことを通知する。
	FooterMore
)
