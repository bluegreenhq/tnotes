package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

// NoteListMsg はノート一覧からモデルへの通知メッセージ。
type NoteListMsg int

// Cmd は NoteListMsg を返す tea.Cmd を生成する。
func (m NoteListMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

const (
	// NoteListSelect はカーソル移動によりノート選択が変わったことを通知する。
	NoteListSelect NoteListMsg = iota
	// NoteListClickSelect はクリックによるノート選択変更を通知する（フォーカス変更を伴う）。
	NoteListClickSelect
	// NoteListCreate は新規ノート作成を要求する。
	NoteListCreate
	// NoteListTrash はノートのゴミ箱移動を要求する。
	NoteListTrash
	// NoteListUndo はundo操作を要求する。
	NoteListUndo
	// NoteListRedo はredo操作を要求する。
	NoteListRedo
	// NoteListEdit はエディタへのフォーカス切り替えを要求する。
	NoteListEdit
	// NoteListCopy はノート内容のクリップボードコピーを要求する。
	NoteListCopy
	// NoteListDuplicate はノート複製を要求する。
	NoteListDuplicate
	// NoteListMenu はコンテキストメニュー表示を要求する。
	NoteListMenu
	// NoteListFocusPrev はフォルダ一覧へのフォーカス移動を要求する。
	NoteListFocusPrev
)

// NoteListRightClickMsg はノート一覧での右クリックを通知する。
type NoteListRightClickMsg struct {
	NoteIndex int
	AnchorX   int
	AnchorY   int
}

// Cmd は NoteListRightClickMsg を返す tea.Cmd を生成する。
func (m NoteListRightClickMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

// FolderListRightClickMsg はフォルダ一覧での右クリックを通知する。
type FolderListRightClickMsg struct {
	FolderIndex int
	AnchorX     int
	AnchorY     int
}

// Cmd は FolderListRightClickMsg を返す tea.Cmd を生成する。
func (m FolderListRightClickMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

// EditorRightClickMsg はエディタでの右クリックを通知する。
// Menu は右クリック時点で構築されたコンテキストメニュー。Editor はこのメニューを保持しない。
type EditorRightClickMsg struct {
	Menu    *tui.PopupMenu
	AnchorX int
	AnchorY int
}

// Cmd は EditorRightClickMsg を返す tea.Cmd を生成する。
func (m EditorRightClickMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

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
	// EditorSearchChanged は検索テキストが変更されたことを通知する。
	EditorSearchChanged
	// EditorSearchBlur は検索フィールドからフォーカスが外れたことを通知する。
	EditorSearchBlur
	// EditorClickBody はエディタ本文クリックによるフォーカス取得を要求する。
	EditorClickBody
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
	// FolderListMenu はコンテキストメニュー表示を要求する。
	FolderListMenu
	// FolderListStartInput はフォルダ新規作成入力の開始を要求する。
	FolderListStartInput
)

// folderMenuActionMsg はフォルダメニューのアクション実行を運ぶメッセージ。
type folderMenuActionMsg struct {
	idx int
}

func (m folderMenuActionMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

// actionResultMsg はコンポーネント操作の結果を運ぶ汎用メッセージ。
type actionResultMsg struct {
	Err  error
	Info string
}

func (m actionResultMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

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
	// EditorHeaderPin はノートのピン留めを要求する。
	EditorHeaderPin
	// EditorHeaderUnpin はノートのピン留め解除を要求する。
	EditorHeaderUnpin
	// EditorHeaderMove はノートのフォルダ移動を要求する。
	EditorHeaderMove
	// EditorHeaderDuplicate はノート複製を要求する。
	EditorHeaderDuplicate
)

// searchDebounceMsg はデバウンスタイマー発火を表すメッセージ。
type searchDebounceMsg struct {
	id    int
	query string
}

// searchClearedMsg は検索テキストがクリアされたことを表すメッセージ。
type searchClearedMsg struct{}

// noteMoveMsg はノートを別フォルダに移動するメッセージ。
type noteMoveMsg struct {
	DestFolder string
}

func (m noteMoveMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}

// HelpOverlayCloseMsg はヘルプオーバーレイを閉じることを要求する。
type HelpOverlayCloseMsg struct{}

// HelpOverlayQuitMsg はヘルプオーバーレイ表示中にアプリ終了が要求されたことを通知する。
type HelpOverlayQuitMsg struct{}

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

// PopupMenuSelectedMsg はポップアップメニューで項目が選択されたことを通知する。
type PopupMenuSelectedMsg struct {
	Kind  PopupKind
	Index int
}

// PopupMenuClosedMsg はポップアップメニューが選択無しで閉じられたことを通知する。
type PopupMenuClosedMsg struct {
	Kind PopupKind
}

// EditorHeaderOpenMenuMsg はエディタヘッダーの「…」ボタンクリックでメニューを開くことを要求する。
type EditorHeaderOpenMenuMsg struct{}

// FooterToggleMenuMsg はフッターメニューの開閉トグルを要求する。
type FooterToggleMenuMsg struct{}

// OpenConfirmDeleteFolderMsg はフォルダ削除確認ダイアログの表示を要求する。
type OpenConfirmDeleteFolderMsg struct {
	Name      string
	NoteCount int
}

// QuitMsg はアプリケーション終了を要求する。
// 各 pane / フッターから emit され、Model 側で同期保存後に tea.Quit を返す。
type QuitMsg struct{}

// OpenHelpMsg はショートカットヘルプ表示を要求する。
type OpenHelpMsg struct{}

// ToggleFolderListMsg はフォルダ一覧の表示切り替えを要求する。
type ToggleFolderListMsg struct{}
