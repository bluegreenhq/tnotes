package ui

// FolderKind はフォルダの種類を表す。
type FolderKind int

const (
	// FolderNotes は通常のNotesフォルダ。
	FolderNotes FolderKind = iota
	// FolderTrash はゴミ箱フォルダ。
	FolderTrash
)

const (
	folderListHeaderLines = 2 // タイトル + 区切り線
	folderListBorderWidth = 2 // 左右ボーダー分
	defaultFolderListW    = 20
	minFolderListWidth    = 15
)

// Folder はフォルダ1件を表す。
type Folder struct {
	Name  string
	Kind  FolderKind
	Count int
}

// FolderList はフォルダ一覧の状態を表す。
type FolderList struct {
	folders  []Folder
	selected int
	width    int
	height   int
	visible  bool
}

// NewFolderList は新しい FolderList を生成する。
func NewFolderList(width, height int) FolderList {
	return FolderList{
		folders: []Folder{
			{Name: "Notes", Kind: FolderNotes, Count: 0},
			{Name: "Trash", Kind: FolderTrash, Count: 0},
		},
		selected: 0,
		width:    width,
		height:   height,
		visible:  false,
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

// UpdateCounts はフォルダのノート件数を更新する。
func (fl *FolderList) UpdateCounts(notesCount, trashCount int) {
	for i := range fl.folders {
		switch fl.folders[i].Kind {
		case FolderNotes:
			fl.folders[i].Count = notesCount
		case FolderTrash:
			fl.folders[i].Count = trashCount
		}
	}
}

// Width は現在の幅を返す。
func (fl *FolderList) Width() int { return fl.width }

// SetSize はフォルダ一覧のサイズを設定する。
func (fl *FolderList) SetSize(width, height int) {
	fl.width = width
	fl.height = height
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

	if contentY < len(fl.folders) {
		return contentY
	}

	return -1
}
