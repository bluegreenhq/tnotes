package ui

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
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
	app        *app.App
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
func NewFolderList(a *app.App, width, height int) FolderList {
	return FolderList{
		app: a,
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

// IsTrash は現在 Trash フォルダを選択しているかを返す。
func (fl *FolderList) IsTrash() bool {
	return fl.SelectedKind() == FolderTrash
}

// CurrentFolderName は現在選択中のフォルダ名を返す。
// Trash なら note.TrashDir、未選択なら app.DefaultFolder を返す。
func (fl *FolderList) CurrentFolderName() string {
	if fl.IsTrash() {
		return note.TrashDir
	}

	name := fl.SelectedName()
	if name == "" {
		return app.DefaultFolder
	}

	return name
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

// UpdateCounts は各フォルダのノート件数を App から取得して更新する。
func (fl *FolderList) UpdateCounts() {
	notesCount := len(fl.app.ListByFolder(app.DefaultFolder))

	for i := range fl.folders {
		switch fl.folders[i].Kind {
		case FolderNotes:
			fl.folders[i].Count = notesCount
		case FolderTrash:
			fl.folders[i].Count = len(fl.app.ListTrashNotes())
		case FolderUser:
			count, err := fl.app.FolderNoteCount(fl.folders[i].Name)
			if err == nil {
				fl.folders[i].Count = count
			}
		}
	}
}

// RefreshFromApp は App からフォルダ一覧を再取得して表示を更新する。
func (fl *FolderList) RefreshFromApp() error {
	folders, err := fl.app.ListFolders()
	if err != nil {
		return err
	}

	notesCount := len(fl.app.ListByFolder(app.DefaultFolder))

	folderCounts := make(map[string]int, len(folders))
	for _, name := range folders {
		count, err := fl.app.FolderNoteCount(name)
		if err != nil {
			continue
		}

		folderCounts[name] = count
	}

	fl.SetFolders(folders, notesCount, len(fl.app.ListTrashNotes()), folderCounts)
	_ = fl.SelectIndex(0)

	return nil
}

// CreateFolder はフォルダを作成し、一覧を更新する。
func (fl *FolderList) CreateFolder(name string) error {
	err := fl.app.CreateFolder(name)
	if err != nil {
		return err
	}

	_ = fl.RefreshFromApp()

	return nil
}

// RenameFolder はフォルダをリネームし、一覧を更新する。
func (fl *FolderList) RenameFolder(oldName, newName string) error {
	err := fl.app.RenameFolder(oldName, newName)
	if err != nil {
		return err
	}

	_ = fl.RefreshFromApp()

	return nil
}

// DeleteFolder はフォルダを削除し、一覧を更新する。
// 戻り値はゴミ箱に移動したノート件数。
func (fl *FolderList) DeleteFolder(name string) (int, error) {
	deleted, err := fl.app.DeleteFolder(name)
	if err != nil {
		return 0, err
	}

	_ = fl.RefreshFromApp()

	return deleted, nil
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

	if val == "" {
		return nil
	}

	err := fl.CreateFolder(val)
	if err != nil {
		return actionCmd(resultAction{Err: err, Info: ""})
	}

	return actionCmd(resultAction{Err: nil, Info: "Created: " + val})
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

	if val == "" || val == oldName {
		return nil
	}

	err := fl.RenameFolder(oldName, val)
	if err != nil {
		return actionCmd(resultAction{Err: err, Info: ""})
	}

	return actionCmd(resultAction{Err: nil, Info: "Renamed: " + oldName + " → " + val})
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

// UpdatePane は PaneComponent インターフェース実装。
func (fl *FolderList) UpdatePane(msg tea.Msg, _ PaneContext) tea.Cmd {
	var cmd tea.Cmd

	*fl, cmd = fl.Update(msg)

	return cmd
}

// HandleBlinkMsg は BlinkHandler インターフェース実装。
func (fl *FolderList) HandleBlinkMsg(msg tui.CursorBlinkMsg) tea.Cmd {
	return fl.blink.HandleMsg(msg)
}

// Update はメッセージに応じてフォルダ一覧の状態を更新する。
func (fl *FolderList) Update(msg tea.Msg) (FolderList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseRight {
			return fl.handleRightClick(msg)
		}

		cmd := fl.handleClickLocal(msg.X, msg.Y)

		return *fl, cmd
	case tea.MouseMotionMsg:
		fl.HandleHoverLocal(msg.Mouse().X, msg.Mouse().Y)

		return *fl, nil
	case tea.KeyPressMsg:
		return fl.handleKeyMsg(msg)
	case tea.MouseMsg:
		fl.HandleHoverLocal(msg.Mouse().X, msg.Mouse().Y)

		return *fl, nil
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
		return actionCmd(folderSelectAction{})
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

// ClearHover はフォルダ一覧の全 hover 状態をクリアする。
func (fl *FolderList) ClearHover() {
	fl.clearHeaderHover()
}

// TryDeleteFolder はフォルダ削除を試行する。
// 空フォルダなら即時削除して resultAction を返す。
// ノートが存在する場合は openConfirmDeleteFolderAction を返して Model にダイアログ表示を委ねる。
func (fl *FolderList) TryDeleteFolder(name string) tea.Cmd {
	count, err := fl.app.FolderNoteCount(name)
	if err != nil {
		return actionCmd(resultAction{Err: err, Info: ""})
	}

	if count > 0 {
		return actionCmd(openConfirmDeleteFolderAction{Name: name, NoteCount: count})
	}

	// 空フォルダは即時削除
	_, err = fl.DeleteFolder(name)
	if err != nil {
		return actionCmd(resultAction{Err: err, Info: ""})
	}

	return actionCmd(resultAction{Err: nil, Info: "Deleted: " + name})
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
			return actionCmd(toggleFolderListAction{})
		case headerHitAdd:
			return actionCmd(folderStartInputAction{})
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

func (fl *FolderList) handleRightClick(msg tea.MouseClickMsg) (FolderList, tea.Cmd) {
	idx := fl.HitTest(msg.X, msg.Y)
	if idx >= 0 {
		fl.SelectIndex(idx)
	}

	if !fl.IsUserFolder() {
		return *fl, nil
	}

	fl.OpenMenu()

	return *fl, actionCmd(folderRightClickAction{
		AnchorX: msg.X,
		AnchorY: msg.Y,
	})
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

	for i := range slices.Backward(runes) {
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
			return actionCmd(folderMenuItemAction{Idx: idx})
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
		return *fl, actionCmd(folderOpenMenuAction{})
	case tea.KeyEnter, tea.KeyTab:
		return *fl, actionCmd(folderFocusNextAction{})
	case 'q':
		return *fl, actionCmd(quitAction{})
	case '?':
		return *fl, actionCmd(openHelpAction{})
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
		return *fl, actionCmd(toggleFolderListAction{})
	}

	return *fl, nil
}

func (fl *FolderList) moveUp() (FolderList, tea.Cmd) {
	if fl.selected > 0 {
		fl.selected--

		return *fl, actionCmd(folderSelectAction{})
	}

	return *fl, nil
}

func (fl *FolderList) moveDown() (FolderList, tea.Cmd) {
	if fl.selected < len(fl.folders)-1 {
		fl.selected++

		return *fl, actionCmd(folderSelectAction{})
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

// --- FolderList → Model アクション ---

// folderSelectAction はフォルダ選択変更を Model に通知する。
type folderSelectAction struct{}

// Apply は選択フォルダに応じて表示を切り替える。
func (folderSelectAction) Apply(m *Model, ctx ActionContext) tea.Cmd {
	return m.handleFolderSelect(ctx.Now)
}

// folderFocusNextAction はノート一覧へのフォーカス移動を要求する。
type folderFocusNextAction struct{}

// Apply はフォーカスを NoteList に移す。
func (folderFocusNextAction) Apply(m *Model, _ ActionContext) tea.Cmd {
	m.Focus = FocusNoteList

	return nil
}

// folderOpenMenuAction はフォルダ moreメニュー（fixed popup）の表示を要求する。
type folderOpenMenuAction struct{}

// Apply はユーザー定義フォルダの場合のみメニューを開く。
func (folderOpenMenuAction) Apply(m *Model, _ ActionContext) tea.Cmd {
	if m.FolderList.IsUserFolder() {
		m.FolderList.OpenMenu()
		m.openFixedPopup(m.FolderList.PopupMenu, m.folderListMenuOrigin, PopupKindFolderList)
	}

	return nil
}

// folderStartInputAction はフォルダ新規作成インライン入力の開始を要求する。
type folderStartInputAction struct{}

// Apply は入力モードに切り替え、FolderList にフォーカスを移す。
func (folderStartInputAction) Apply(m *Model, _ ActionContext) tea.Cmd {
	blinkCmd := m.FolderList.StartInput()
	m.Focus = FocusFolderList

	return blinkCmd
}

// folderRightClickAction はフォルダ一覧での右クリックによるアンカー付きポップアップ表示を要求する。
type folderRightClickAction struct {
	AnchorX int
	AnchorY int
}

// Apply はアンカー付きポップアップメニューを開く。
func (a folderRightClickAction) Apply(m *Model, _ ActionContext) tea.Cmd {
	m.openAnchoredPopup(m.FolderList.PopupMenu, a.AnchorX, a.AnchorY, PopupKindFolderList)

	return nil
}

// folderMenuItemAction はフォルダメニュー項目の選択を Model に通知する。
type folderMenuItemAction struct {
	Idx int
}

// Apply は選択されたメニュー項目に応じた処理を実行する。
func (a folderMenuItemAction) Apply(m *Model, ctx ActionContext) tea.Cmd {
	return m.handleFolderMenuAction(a.Idx, ctx.Now)
}

// openConfirmDeleteFolderAction はフォルダ削除確認ダイアログの表示を要求する。
type openConfirmDeleteFolderAction struct {
	Name      string
	NoteCount int
}

// Apply は確認ダイアログをオーバーレイとして開く。
func (a openConfirmDeleteFolderAction) Apply(m *Model, _ ActionContext) tea.Cmd {
	m.openConfirmDeleteFolder(a.Name, a.NoteCount)

	return nil
}
