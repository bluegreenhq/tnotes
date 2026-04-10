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

// FolderListMsg はフォルダ一覧からモデルへの通知メッセージ。
type FolderListMsg int

const (
	// FolderListSelect はフォルダ選択変更を通知する。
	FolderListSelect FolderListMsg = iota
	// FolderListFocusNext はノート一覧へのフォーカス移動を要求する。
	FolderListFocusNext
)

// FooterMsg はフッターからモデルへの通知メッセージ。
type FooterMsg int

const (
	// FooterRestore は復元ボタンがクリックされたことを通知する。
	FooterRestore FooterMsg = iota
	// FooterQuit は終了ボタンがクリックされたことを通知する。
	FooterQuit
	// FooterCopy はコピーボタンがクリックされたことを通知する。
	FooterCopy
	// FooterCut はカットボタンがクリックされたことを通知する。
	FooterCut
	// FooterMore はMoreボタンがクリックされたことを通知する。
	FooterMore
)

// EditorHeaderMsg はエディタヘッダーからモデルへの通知メッセージ。
type EditorHeaderMsg int

const (
	// EditorHeaderNew は新規ノート作成を要求する。
	EditorHeaderNew EditorHeaderMsg = iota
	// EditorHeaderTrash はノートのゴミ箱移動を要求する。
	EditorHeaderTrash
	// EditorHeaderCopy はノート内容のクリップボードコピーを要求する。
	EditorHeaderCopy
	// EditorHeaderRestore はゴミ箱からのノート復元を要求する。
	EditorHeaderRestore
)
