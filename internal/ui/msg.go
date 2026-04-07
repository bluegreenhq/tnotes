package ui

// SidebarMsg はサイドバーからモデルへの通知メッセージ。
type SidebarMsg int

const (
	// SidebarSelect はカーソル移動によりノート選択が変わったことを通知する。
	SidebarSelect SidebarMsg = iota
	// SidebarCreate は新規ノート作成を要求する。
	SidebarCreate
	// SidebarTrash はノートのゴミ箱移動を要求する。
	SidebarTrash
	// SidebarEnterTrash はゴミ箱モードへの切り替えを要求する。
	SidebarEnterTrash
	// SidebarExitTrash はゴミ箱モードからの復帰を要求する。
	SidebarExitTrash
	// SidebarRestore はゴミ箱からのノート復元を要求する。
	SidebarRestore
	// SidebarUndo はundo操作を要求する。
	SidebarUndo
	// SidebarRedo はredo操作を要求する。
	SidebarRedo
	// SidebarEdit はエディタへのフォーカス切り替えを要求する。
	SidebarEdit
	// SidebarQuit は終了を要求する。
	SidebarQuit
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
)
