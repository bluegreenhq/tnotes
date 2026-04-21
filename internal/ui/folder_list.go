package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"
)

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
	// レイアウト: ... " + "
	//       右端から:  2 1
	headerCloseBtnWidth      = 2
	headerAddBtnOffset       = 2 // contentWidth - 2 = +
	headerHitClose           = "close"
	headerHitAdd             = "add"
	folderListSystemAndTrash = 2 // Notes + Trash
)

// Folder はフォルダ1件を表す。
type Folder struct {
	Name  string
	Kind  FolderKind
	Count int
}

// NewFolder は新しい Folder を生成する。
func NewFolder(name string, kind FolderKind, count int) Folder {
	return Folder{Name: name, Kind: kind, Count: count}
}

// FolderList はフォルダ一覧の状態を表す。
type FolderList struct {
	folders    []Folder
	selected   int
	width      int
	height     int
	layout     *Layout
	visible    bool
	inputMode  bool            // インライン入力中かどうか（新規作成）
	renameMode bool            // リネーム入力中かどうか
	renameName string          // リネーム元のフォルダ名
	lineInput  tui.LineInput   // インライン入力の状態
	blink      tui.CursorBlink // カーソル点滅状態
	menuOpen   bool            // moreメニュー表示中かどうか
	PopupMenu  *tui.PopupMenu
	hoverClose bool
	hoverAdd   bool
}

// NewFolderList は新しい FolderList を生成する。
func NewFolderList(width, height int) FolderList {
	return FolderList{
		folders: []Folder{
			NewFolder("Notes", FolderNotes, 0),
			NewFolder("Trash", FolderTrash, 0),
		},
		selected:   0,
		width:      width,
		height:     height,
		layout:     nil,
		visible:    false,
		inputMode:  false,
		renameMode: false,
		renameName: "",
		lineInput:  tui.NewLineInput(),
		blink:      tui.NewCursorBlink(blinkOwnerFolderList),
		menuOpen:   false,
		PopupMenu:  tui.NewPopupMenu(nil),
		hoverClose: false,
		hoverAdd:   false,
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
	fl.PopupMenu = tui.NewPopupMenu([]tui.MenuItem{
		tui.NewMenuItem("Rename"),
		tui.NewMenuItem("Delete"),
	})
	fl.menuOpen = true
}

// CloseMenu はmoreメニューを閉じる。
func (fl *FolderList) CloseMenu() {
	fl.menuOpen = false
	fl.PopupMenu.SetHover(-1)
}

// MenuHeight はメニューの高さを返す。
func (fl *FolderList) MenuHeight() int {
	if !fl.menuOpen {
		return 0
	}

	return fl.PopupMenu.Height()
}

// MenuLeftX はメニュー左端のフォルダリスト相対X座標を返す。
func (fl *FolderList) MenuLeftX() int {
	return fl.Width() - folderListBorderWidth - fl.PopupMenu.Width()
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
	folders = append(folders, NewFolder("Notes", FolderNotes, notesCount))

	for _, name := range userFolders {
		folders = append(folders, NewFolder(name, FolderUser, folderCounts[name]))
	}

	folders = append(folders, NewFolder("Trash", FolderTrash, trashCount))
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

// StartInput はインライン入力モードを開始する（新規作成用）。
func (fl *FolderList) StartInput() tea.Cmd {
	fl.inputMode = true
	fl.lineInput.Reset()

	return fl.blink.Reset()
}

// CommitInput はインライン入力を確定する。値が空なら破棄する。
func (fl *FolderList) CommitInput() tea.Cmd {
	val := fl.lineInput.Value()
	fl.clearInput()

	if val != "" {
		return folderCreateMsg{Name: val}.Cmd()
	}

	return nil
}

// CancelInput はインライン入力を破棄する（Esc用）。
func (fl *FolderList) CancelInput() {
	fl.clearInput()
}

// StartRename はリネーム入力モードを開始する。
func (fl *FolderList) StartRename() tea.Cmd {
	name := fl.SelectedName()
	fl.renameMode = true
	fl.renameName = name
	fl.lineInput.SetValue(name)

	return fl.blink.Reset()
}

// CommitRename はリネーム入力を確定する。値が空または変更なしなら破棄する。
func (fl *FolderList) CommitRename() tea.Cmd {
	val := fl.lineInput.Value()
	oldName := fl.renameName
	fl.clearRename()

	if val != "" && val != oldName {
		return folderRenameMsg{OldName: oldName, NewName: val}.Cmd()
	}

	return nil
}

// CancelRename はリネーム入力を破棄する（Esc用）。
func (fl *FolderList) CancelRename() {
	fl.clearRename()
}

// HandleHoverLocal はローカル座標でホバーを処理する。
func (fl *FolderList) HandleHoverLocal(x, y int) {
	if x < fl.width && y == 0 {
		fl.setHeaderHover(x, y)
	} else {
		fl.clearHeaderHover()
	}

	if fl.menuOpen {
		menuTopY := folderListHeaderLines
		menuHeight := fl.MenuHeight()
		menuX := fl.MenuLeftX()
		menuWidth := fl.PopupMenu.Width()

		if y >= menuTopY && y < menuTopY+menuHeight && x >= menuX && x < menuX+menuWidth {
			fl.PopupMenu.SetHoverByPos(x-menuX, y-menuTopY)
		} else {
			fl.PopupMenu.SetHoverByPos(-1, -1)
		}
	}
}

// Update はメッセージに応じてフォルダ一覧の状態を更新する。
func (fl *FolderList) Update(msg tea.Msg) (FolderList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		cmd := fl.handleClickLocal(msg.X, msg.Y)

		return *fl, cmd
	case tea.KeyPressMsg:
		return fl.handleKeyMsg(msg)
	}

	return *fl, nil
}

// SelectIndex はインデックスを指定して選択する。
func (fl *FolderList) SelectIndex(idx int) tea.Cmd {
	if idx < 0 || idx >= len(fl.folders) {
		return nil
	}

	prev := fl.selected
	fl.selected = idx

	if prev != fl.selected {
		return FolderListSelect.Cmd()
	}

	return nil
}

var folderListStyle = lipgloss.NewStyle().
	BorderRight(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("8"))

// View はフォルダ一覧の描画内容を返す。
func (fl *FolderList) View(focused bool, hoverSeparator bool) string {
	contentWidth := max(fl.width-folderListBorderWidth, 0)

	var b strings.Builder

	// ヘッダー: 閉じるボタン + タイトル + ボタン
	fl.renderHeader(&b, contentWidth)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", contentWidth))
	b.WriteString("\n")

	// フォルダ一覧
	for i, folder := range fl.folders {
		if fl.renameMode && i == fl.selected {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(fl.lineInput.View(fl.blink.Visible())))
		} else {
			fl.renderFolder(&b, folder, i == fl.selected, contentWidth)
		}

		b.WriteString("\n\n")
	}

	// インライン入力
	if fl.inputMode {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(fl.lineInput.View(fl.blink.Visible())))
		b.WriteString("\n\n")
	}

	// 残りの高さを埋める
	usedLines := folderListHeaderLines + len(fl.folders)*folderListItemHeight
	if fl.inputMode {
		usedLines += folderListItemHeight
	}

	for i := usedLines; i < fl.height; i++ {
		b.WriteString("\n")
	}

	style := folderListStyle

	if focused {
		style = style.BorderForeground(lipgloss.Color("4"))
	}

	if hoverSeparator {
		style = style.BorderStyle(lipgloss.ThickBorder())
	}

	return style.Width(fl.width).Height(fl.height).Render(b.String())
}

func (fl *FolderList) handleClickLocal(x, y int) tea.Cmd {
	// moreメニューが開いている場合
	if fl.menuOpen {
		return fl.handleMenuClick(x, y)
	}

	// ヘッダーのボタンクリック判定
	if y < folderListHeaderLines {
		hit := fl.hitTestHeader(x, y)

		switch hit {
		case headerHitClose:
			return FolderListClose.Cmd()
		case headerHitAdd:
			return FolderListStartInput.Cmd()
		}

		return nil
	}

	// リスト項目クリック
	idx := fl.HitTest(x, y)
	if idx >= 0 {
		return fl.SelectIndex(idx)
	}

	return nil
}

// hitTestHeader はヘッダー領域のクリック判定を行う.
// 戻り値: headerHitClose, headerHitAdd, "" (該当なし).
func (fl *FolderList) hitTestHeader(x, y int) string {
	if y != 0 {
		return ""
	}

	contentWidth := max(fl.width-folderListBorderWidth, 0)

	// ✕ ボタン (左端)
	if x <= headerCloseBtnWidth {
		return headerHitClose
	}

	// + ボタン (右端)
	if x == contentWidth-headerAddBtnOffset {
		return headerHitAdd
	}

	return ""
}

func (fl *FolderList) setHeaderHover(x, y int) {
	fl.hoverClose = false
	fl.hoverAdd = false

	if y != 0 {
		return
	}

	hit := fl.hitTestHeader(x, y)

	switch hit {
	case headerHitClose:
		fl.hoverClose = true
	case headerHitAdd:
		fl.hoverAdd = true
	}
}

func (fl *FolderList) clearHeaderHover() {
	fl.hoverClose = false
	fl.hoverAdd = false
}

var folderCountStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

func (fl *FolderList) renderFolder(b *strings.Builder, folder Folder, selected bool, contentWidth int) {
	name := " " + folder.Name
	count := fmt.Sprintf("%d ", folder.Count)
	name = truncateForCount(name, count, contentWidth)
	padding := max(contentWidth-lipgloss.Width(name)-lipgloss.Width(count), 0)

	if selected {
		baseStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("4")).
			Bold(true)
		nameStr := baseStyle.Foreground(lipgloss.Color("15")).Render(name)
		pad := baseStyle.Render(strings.Repeat(" ", padding))
		countStr := baseStyle.Foreground(lipgloss.Color("7")).Render(count)
		b.WriteString(nameStr + pad + countStr)
	} else {
		nameStr := lipgloss.NewStyle().Render(name)
		pad := strings.Repeat(" ", padding)
		countStr := folderCountStyle.Render(count)
		b.WriteString(nameStr + pad + countStr)
	}
}

// truncateForCount はフォルダ名をカウント文字列と合わせて contentWidth に収まるように省略する。
func truncateForCount(name, count string, contentWidth int) string {
	nameW := lipgloss.Width(name)
	countW := lipgloss.Width(count)
	maxNameW := contentWidth - countW

	if nameW <= maxNameW {
		return name
	}

	runes := []rune(name)

	const ellipsis = "…"

	for i := len(runes) - 1; i >= 0; i-- {
		candidate := string(runes[:i]) + ellipsis
		if lipgloss.Width(candidate) <= maxNameW {
			return candidate
		}
	}

	return ellipsis
}

func (fl *FolderList) renderHeader(b *strings.Builder, contentWidth int) {
	closeBtn := " "
	if fl.hoverClose {
		closeBtn += buttonHoverStyle.Render("✕")
	} else {
		closeBtn += buttonStyle.Render("✕")
	}

	title := " Folders"

	addBtn := " "
	if fl.hoverAdd {
		addBtn += buttonHoverStyle.Render("+")
	} else {
		addBtn += buttonStyle.Render("+")
	}

	headerLeft := closeBtn + title
	headerRight := addBtn + " "
	headerLeftWidth := lipgloss.Width(headerLeft)
	headerRightWidth := lipgloss.Width(headerRight)
	padding := max(contentWidth-headerLeftWidth-headerRightWidth, 0)

	headerStr := headerLeft + strings.Repeat(" ", padding) + headerRight
	b.WriteString(lipgloss.NewStyle().Bold(true).Width(contentWidth).Render(headerStr))
}

func (fl *FolderList) clearInput() {
	fl.inputMode = false
	fl.lineInput.Reset()
	fl.blink.Stop()
}

func (fl *FolderList) clearRename() {
	fl.renameMode = false
	fl.renameName = ""
	fl.lineInput.Reset()
	fl.blink.Stop()
}

func (fl *FolderList) handleMenuClick(x, y int) tea.Cmd {
	menuTopY := folderListHeaderLines
	menuHeight := fl.MenuHeight()
	menuX := fl.MenuLeftX()
	menuWidth := fl.PopupMenu.Width()

	if y >= menuTopY && y < menuTopY+menuHeight && x >= menuX && x < menuX+menuWidth {
		relX := x - menuX
		relY := y - menuTopY

		idx, hit := fl.PopupMenu.HandleClick(relX, relY)
		fl.CloseMenu()

		if hit {
			return folderMenuActionMsg{idx: idx}.Cmd()
		}

		return nil
	}

	fl.CloseMenu()

	return nil
}

func (fl *FolderList) handleKeyNav(keyMsg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	if keyMsg.Mod&tea.ModCtrl != 0 {
		return fl.handleCtrlKeyNav(keyMsg)
	}

	switch keyMsg.Code {
	case tea.KeyUp, 'k':
		return fl.moveUp()
	case tea.KeyDown, 'j':
		return fl.moveDown()
	case 'm':
		return *fl, FolderListMenu.Cmd()
	case tea.KeyEnter, tea.KeyTab:
		return *fl, FolderListFocusNext.Cmd()
	case 'q':
		return *fl, FolderListQuit.Cmd()
	case '?':
		return *fl, FolderListHelp.Cmd()
	}

	return *fl, nil
}

func (fl *FolderList) handleCtrlKeyNav(keyMsg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	switch keyMsg.Code {
	case 'p':
		return fl.moveUp()
	case 'n':
		return fl.moveDown()
	case 'b':
		return *fl, FolderListClose.Cmd()
	}

	return *fl, nil
}

func (fl *FolderList) moveUp() (FolderList, tea.Cmd) {
	if fl.selected > 0 {
		fl.selected--

		return *fl, FolderListSelect.Cmd()
	}

	return *fl, nil
}

func (fl *FolderList) moveDown() (FolderList, tea.Cmd) {
	if fl.selected < len(fl.folders)-1 {
		fl.selected++

		return *fl, FolderListSelect.Cmd()
	}

	return *fl, nil
}

func (fl *FolderList) updateRename(keyMsg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	result := fl.lineInput.HandleKey(keyMsg)

	switch result {
	case tui.LineInputNone:
		// blink reset は model_update 側で行う
	case tui.LineInputSubmit:
		return *fl, fl.CommitRename()
	case tui.LineInputCancel:
		fl.CancelRename()

		return *fl, nil
	}

	return *fl, nil
}

func (fl *FolderList) updateInput(keyMsg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	result := fl.lineInput.HandleKey(keyMsg)

	switch result {
	case tui.LineInputNone:
		// blink reset は model_update 側で行う
	case tui.LineInputSubmit:
		return *fl, fl.CommitInput()
	case tui.LineInputCancel:
		fl.CancelInput()

		return *fl, nil
	}

	return *fl, nil
}

func (fl *FolderList) handleKeyMsg(msg tea.KeyPressMsg) (FolderList, tea.Cmd) {
	// メニュー表示中
	if fl.menuOpen {
		if msg.Code == tea.KeyEscape {
			fl.CloseMenu()

			return *fl, nil
		}

		fl.CloseMenu()
	}

	// リネーム入力モード
	if fl.renameMode {
		return fl.updateRename(msg)
	}

	// インライン入力モード
	if fl.inputMode {
		return fl.updateInput(msg)
	}

	return fl.handleKeyNav(msg)
}
