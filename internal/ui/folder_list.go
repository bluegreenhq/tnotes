package ui

// FolderKind はフォルダの種類を表す。
type FolderKind int

const (
	// FolderNotes は通常のNotesフォルダ。
	FolderNotes FolderKind = iota
	// FolderTrash はゴミ箱フォルダ。
	FolderTrash
	// FolderUser はユーザー定義フォルダ。
	FolderUser
)

const (
	folderListHeaderLines = 2 // タイトル + 区切り線
	folderListBorderWidth = 1 // 右ボーダー分
	folderListItemHeight  = 2 // 名前 + 空行
	defaultFolderListW    = 20
	minFolderListWidth    = 15

	// ヘッダーボタンHitTest用定数.
	// レイアウト（moreあり）: ... " + ⋯ "
	//                  右端から:  4 3 2 1
	// レイアウト（moreなし）: ... " + "
	//                  右端から:  2 1
	headerCloseBtnWidth      = 2
	headerMoreBtnOffset      = 2 // contentWidth - 2 = ⋯
	headerAddBtnOffsetNoMore = 2 // contentWidth - 2 = +（moreなし時）
	headerAddBtnOffsetMore   = 4 // contentWidth - 4 = +（moreあり時）
	headerHitClose           = "close"
	headerHitAdd             = "add"
	headerHitMore            = "more"
	folderListSystemAndTrash = 2 // Notes + Trash
)

// Folder はフォルダ1件を表す。
type Folder struct {
	Name  string
	Kind  FolderKind
	Count int
}

// FolderList はフォルダ一覧の状態を表す。
type FolderList struct {
	folders    []Folder
	selected   int
	width      int
	height     int
	visible    bool
	inputMode  bool   // インライン入力中かどうか（新規作成）
	inputValue string // 入力中のフォルダ名
	renameMode bool   // リネーム入力中かどうか
	renameName string // リネーム元のフォルダ名
	menuOpen   bool   // moreメニュー表示中かどうか
	PopupMenu  *PopupMenu
	hoverClose bool
	hoverAdd   bool
	hoverMore  bool
}

// NewFolderList は新しい FolderList を生成する。
func NewFolderList(width, height int) FolderList {
	return FolderList{
		folders: []Folder{
			{Name: "Notes", Kind: FolderNotes, Count: 0},
			{Name: "Trash", Kind: FolderTrash, Count: 0},
		},
		selected:   0,
		width:      width,
		height:     height,
		visible:    false,
		inputMode:  false,
		inputValue: "",
		renameMode: false,
		renameName: "",
		menuOpen:   false,
		PopupMenu:  NewPopupMenu(nil),
		hoverClose: false,
		hoverAdd:   false,
		hoverMore:  false,
	}
}

// Visible はフォルダ一覧の表示状態を返す。
func (fl *FolderList) Visible() bool { return fl.visible }

// SetVisible はフォルダ一覧の表示状態を設定する。
func (fl *FolderList) SetVisible(v bool) { fl.visible = v }

// ToggleVisible はフォルダ一覧の表示状態をトグルする。
func (fl *FolderList) ToggleVisible() { fl.visible = !fl.visible }

// SelectedKind は選択中のフォルダの種類を返す。
func (fl *FolderList) SelectedKind() FolderKind {
	return fl.folders[fl.selected].Kind
}

// SelectedIndex は選択中のインデックスを返す。
func (fl *FolderList) SelectedIndex() int { return fl.selected }

// IndexByKind は指定KindのフォルダのインデックスをFolderList内から検索して返す。該当なしは -1。
func (fl *FolderList) IndexByKind(kind FolderKind) int {
	for i, f := range fl.folders {
		if f.Kind == kind {
			return i
		}
	}

	return -1
}

// IndexByName は指定名のフォルダのインデックスを返す。該当なしは -1。
func (fl *FolderList) IndexByName(name string) int {
	for i, f := range fl.folders {
		if f.Name == name {
			return i
		}
	}

	return -1
}

// Width は現在の幅を返す。
func (fl *FolderList) Width() int { return fl.width }

// SetSize はフォルダ一覧のサイズを設定する。
func (fl *FolderList) SetSize(width, height int) {
	fl.width = width
	fl.height = height
}

// InputMode はインライン入力中かどうかを返す（新規作成）。
func (fl *FolderList) InputMode() bool { return fl.inputMode }

// RenameMode はリネーム入力中かどうかを返す。
func (fl *FolderList) RenameMode() bool { return fl.renameMode }

// MenuOpen はmoreメニュー表示中かどうかを返す。
func (fl *FolderList) MenuOpen() bool { return fl.menuOpen }

// OpenMenu はmoreメニューを開く。
func (fl *FolderList) OpenMenu() {
	fl.PopupMenu = NewPopupMenu([]MenuItem{
		{Label: "Rename", Disabled: false},
		{Label: "Delete", Disabled: false},
	})
	fl.menuOpen = true
}

// CloseMenu はmoreメニューを閉じる。
func (fl *FolderList) CloseMenu() {
	fl.menuOpen = false
	fl.PopupMenu.hover = -1
}

// MenuHeight はメニューの高さを返す。
func (fl *FolderList) MenuHeight() int {
	if !fl.menuOpen {
		return 0
	}

	return fl.PopupMenu.Height()
}

// IsUserFolder は選択中のフォルダがユーザー定義フォルダかどうかを返す。
func (fl *FolderList) IsUserFolder() bool {
	if fl.selected < 0 || fl.selected >= len(fl.folders) {
		return false
	}

	return fl.folders[fl.selected].Kind == FolderUser
}

// SelectedName は選択中のフォルダ名を返す。
func (fl *FolderList) SelectedName() string {
	if fl.selected < 0 || fl.selected >= len(fl.folders) {
		return ""
	}

	return fl.folders[fl.selected].Name
}

// SetFolders はフォルダ一覧を再構成する。表示順: Notes → ユーザーフォルダ（アルファベット順）→ Trash。
func (fl *FolderList) SetFolders(userFolders []string, notesCount, trashCount int, folderCounts map[string]int) {
	folders := make([]Folder, 0, len(userFolders)+folderListSystemAndTrash)
	folders = append(folders, Folder{Name: "Notes", Kind: FolderNotes, Count: notesCount})

	for _, name := range userFolders {
		folders = append(folders, Folder{Name: name, Kind: FolderUser, Count: folderCounts[name]})
	}

	folders = append(folders, Folder{Name: "Trash", Kind: FolderTrash, Count: trashCount})
	fl.folders = folders

	if fl.selected >= len(fl.folders) {
		fl.selected = len(fl.folders) - 1
	}
}

// HitTest は座標からクリックされたフォルダのインデックスを返す。該当なしは -1。
func (fl *FolderList) HitTest(x, y int) int {
	if x < 0 || x >= fl.width {
		return -1
	}

	contentY := y - folderListHeaderLines
	if contentY < 0 {
		return -1
	}

	// 各フォルダは folderListItemHeight 行(名前+空行)を占める。空行の場合は該当なし。
	folderIdx := contentY / folderListItemHeight
	if contentY%folderListItemHeight != 0 {
		return -1
	}

	if folderIdx < len(fl.folders) {
		return folderIdx
	}

	return -1
}
