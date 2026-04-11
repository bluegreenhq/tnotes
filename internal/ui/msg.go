package ui

import tea "charm.land/bubbletea/v2"

// NoteListMsg はノート一覧からモデルへの通知メッセージ。
type NoteListMsg int

// Cmd は NoteListMsg を返す tea.Cmd を生成する。
func (m NoteListMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

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

// EditorMsg はエディタからモデルへの通知メッセージ。
type EditorMsg int

// Cmd は EditorMsg を返す tea.Cmd を生成する。
func (m EditorMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

const (
	// EditorBlur はノート一覧へのフォーカス切り替えを要求する。
	EditorBlur EditorMsg = iota
	// EditorSave はノート保存を要求する。
	EditorSave
)

// editorContextMsg はエディタコンテキストメニューのアクション。
type editorContextMsg int

const (
	editorContextCopy editorContextMsg = iota
	editorContextCut
	editorContextPaste
)

// editorOpenURLMsg はカーソル位置のURLをブラウザで開くことを要求するメッセージ。
type editorOpenURLMsg struct {
	URL string
}

func (m editorOpenURLMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

// FolderListMsg はフォルダ一覧からモデルへの通知メッセージ。
type FolderListMsg int

// Cmd は FolderListMsg を返す tea.Cmd を生成する。
func (m FolderListMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

const (
	// FolderListSelect はフォルダ選択変更を通知する。
	FolderListSelect FolderListMsg = iota
	// FolderListFocusNext はノート一覧へのフォーカス移動を要求する。
	FolderListFocusNext
)

// folderCreateMsg はインライン入力で確定されたフォルダ名を運ぶメッセージ。
type folderCreateMsg struct {
	Name string
}

func (m folderCreateMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

// folderDeleteMsg はフォルダ削除を運ぶメッセージ。
type folderDeleteMsg struct {
	Name string
}

func (m folderDeleteMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

// folderRenameMsg はフォルダリネームを運ぶメッセージ。
type folderRenameMsg struct {
	OldName string
	NewName string
}

func (m folderRenameMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

// FooterMsg はフッターからモデルへの通知メッセージ。
type FooterMsg int

// Cmd は FooterMsg を返す tea.Cmd を生成する。
func (m FooterMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

const (
	// FooterQuit は終了ボタンがクリックされたことを通知する。
	FooterQuit FooterMsg = iota
	// FooterCopy はコピーボタンがクリックされたことを通知する。
	FooterCopy
	// FooterCut はカットボタンがクリックされたことを通知する。
	FooterCut
	// FooterMore はMoreボタンがクリックされたことを通知する。
	FooterMore
)

// EditorHeaderMsg はエディタヘッダーからモデルへの通知メッセージ。
type EditorHeaderMsg int

// Cmd は EditorHeaderMsg を返す tea.Cmd を生成する。
func (m EditorHeaderMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

const (
	// EditorHeaderNew は新規ノート作成を要求する。
	EditorHeaderNew EditorHeaderMsg = iota
	// EditorHeaderTrash はノートのゴミ箱移動を要求する。
	EditorHeaderTrash
	// EditorHeaderCopy はノート内容のクリップボードコピーを要求する。
	EditorHeaderCopy
	// EditorHeaderRestore はゴミ箱からのノート復元を要求する。
	EditorHeaderRestore
	// EditorHeaderPin はノートのピン留めを要求する。
	EditorHeaderPin
	// EditorHeaderUnpin はノートのピン留め解除を要求する。
	EditorHeaderUnpin
	// EditorHeaderMove はノートのフォルダ移動を要求する。
	EditorHeaderMove
)

// noteMoveMsg はノートを別フォルダに移動するメッセージ。
type noteMoveMsg struct {
	DestFolder string
}

func (m noteMoveMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}
