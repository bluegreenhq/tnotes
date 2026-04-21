package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

// blinkOwner はアプリ固有の CursorBlink 所有者定数。
const (
	blinkOwnerEditor = iota
	blinkOwnerFolderList
	blinkOwnerSearch
)

// --- イベントハンドラ ---

// Update はメッセージに応じて状態を更新する。
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:cyclop,funlen // type switch dispatch
	now := time.Now()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, m.handleResize(msg, now)
	case tea.KeyPressMsg:
		return m, m.handleKey(msg, now)
	case tea.MouseClickMsg:
		return m, m.handleClick(msg, now)
	case tea.MouseMotionMsg:
		return m, m.handleDrag(msg, now)
	case tea.MouseReleaseMsg:
		return m, m.handleRelease(msg, now)
	case tea.MouseWheelMsg:
		return m, m.routeMouse(msg, now)
	case tea.MouseMsg:
		return m, m.handleHover(msg)
	case tea.FocusMsg:
		return m, m.handleFocusRestore()
	case tea.BlurMsg: // 他アプリへ切り替え時に編集中の内容を保存
		m.syncEditorToNote(now)

		return m, nil
	case tui.CursorBlinkMsg:
		switch msg.Owner {
		case blinkOwnerEditor:
			return m, m.Editor.HandleBlinkMsg(msg)
		case blinkOwnerFolderList:
			return m, m.FolderList.blink.HandleMsg(msg)
		case blinkOwnerSearch:
			return m, m.Editor.Header.searchBlink.HandleMsg(msg)
		}

		return m, nil
	case searchDebounceMsg:
		if msg.id != m.searchDebounceID {
			return m, nil // 古いタイマーは無視
		}

		m.applySearchFilter(msg.query, now)

		return m, nil
	case searchClearedMsg:
		m.applySearchFilter("", now)

		return m, nil
	case clearInfoMsg:
		return m, m.handleClearInfo(msg)
	default:
		return m, m.routeFocus(msg, now)
	}
}

func (m *Model) handleKey(msg tea.KeyPressMsg, now time.Time) tea.Cmd {
	// モーダル状態の処理（優先度順）
	if cmd, handled := m.handleModalKey(msg, now); handled {
		return cmd
	}

	m.errMsg = ""
	m.infoMsg = ""
	m.Footer.CloseMenu()
	m.Editor.Header.CloseMenu()
	m.Editor.Header.CloseMoveMenu()

	// グローバルキー
	if cmd, handled := m.handleGlobalKey(msg, now); handled {
		return cmd
	}

	// フォーカス先に委譲
	cmd := m.routeFocus(msg, now)

	return tea.Batch(cmd, m.resetFocusBlink())
}

func (m *Model) handleModalKey(msg tea.KeyPressMsg, now time.Time) (tea.Cmd, bool) {
	// メニューが開いている場合（右クリック or キーボード起動）
	if menu, kind := m.popup.Active(); menu != nil {
		return m.popup.HandleKey(msg, menu, kind, now), true
	}

	// 確認ダイアログ表示中
	if m.confirmDialog != nil {
		return m.handleConfirmDialogKey(msg), true
	}

	// ヘルプオーバーレイ表示中
	if m.helpOverlay != nil {
		switch m.helpOverlay.Update(msg) {
		case HelpQuit:
			m.helpOverlay = nil
			m.syncEditorToNote(now)

			return tea.Quit, true
		case HelpClose:
			m.helpOverlay = nil
		case HelpContinue:
			// 継続中 — 何もしない
		}

		return nil, true
	}

	return nil, false
}

func (m *Model) handleGlobalKey(msg tea.KeyPressMsg, now time.Time) (tea.Cmd, bool) {
	switch {
	case msg.Code == 'q' && msg.Mod&tea.ModCtrl != 0:
		m.syncEditorToNote(now)

		return tea.Quit, true
	case msg.Code == 'f' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		m.Editor.Header.SetSearchFocused(true)
		m.Focus = FocusEditor

		return m.Editor.Header.searchBlink.Reset(), true
	case msg.Code == '/' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		m.openHelp()

		return nil, true
	}

	return nil, false
}

// routeFocus はフォーカス先のコンポーネントに msg を委譲する。
func (m *Model) routeFocus(msg tea.Msg, now time.Time) tea.Cmd {
	switch m.Focus {
	case FocusFolderList:
		_, cmd := m.FolderList.Update(msg)

		return m.processFolderListCmd(cmd, now)
	case FocusNoteList:
		_, cmd := m.NoteList.Update(msg, now, m.Editor.Header.TrashMode())

		return m.processNoteListCmd(cmd, now)
	case FocusEditor:
		_, cmd := m.Editor.Update(msg, now)

		return m.processEditorCmd(cmd, now)
	}

	return nil
}

// resetFocusBlink はキー入力後にフォーカス先の blink をリセットする。
func (m *Model) resetFocusBlink() tea.Cmd {
	switch m.Focus { //nolint:exhaustive // NoteList にはカーソル blink がない
	case FocusFolderList:
		if m.FolderList.InputMode() || m.FolderList.RenameMode() {
			return m.FolderList.blink.Reset()
		}
	case FocusEditor:
		if m.Editor.Header.SearchFocused() {
			return m.Editor.Header.ResetSearchBlink()
		}

		return m.Editor.resetBlink()
	}

	return nil
}

func (m *Model) handleResize(msg tea.WindowSizeMsg, now time.Time) tea.Cmd {
	m.layout.width = msg.Width
	m.layout.height = msg.Height
	m.layout.folderVisible = m.FolderList.Visible()
	m.layout.noteListWidth = max(m.layout.noteListWidth, minNoteListWidth)
	m.layout.noteListWidth = min(m.layout.noteListWidth, m.layout.MaxNoteListWidth())

	if m.layout.folderVisible {
		maxNoteWidth := m.layout.width - m.layout.folderListWidth - minEditorWidth
		m.layout.noteListWidth = min(m.layout.noteListWidth, maxNoteWidth)
		m.layout.noteListWidth = max(m.layout.noteListWidth, minNoteListWidth)

		maxFolderWidth := m.layout.width - m.layout.noteListWidth - minEditorWidth
		m.layout.folderListWidth = max(m.layout.folderListWidth, minFolderListWidth)
		m.layout.folderListWidth = min(m.layout.folderListWidth, maxFolderWidth)
	}

	m.recalcLayout(now)
	m.Footer.CloseMenu()

	return nil
}

func (m *Model) handleFocusRestore() tea.Cmd {
	changed, mt, err := m.App.RefreshNotes(m.indexModTime)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	if !changed {
		return nil
	}

	m.indexModTime = mt
	m.refreshNoteListKeepSelection(time.Now())
	m.loadSelectedNote()

	return nil
}

func (m *Model) handleClearInfo(msg clearInfoMsg) tea.Cmd {
	if msg.id == m.infoMsgID {
		m.infoMsg = ""
	}

	return nil
}

func (m *Model) handleClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	m.errMsg = ""

	if msg.Button == tea.MouseRight {
		return m.handleRightClick(msg, now)
	}

	if msg.Button != tea.MouseLeft {
		return nil
	}

	// インライン入力中はクリックで確定
	commitCmd := m.commitFolderLineInput()
	clickCmd := m.handleClickInner(msg, now)

	return tea.Batch(commitCmd, clickCmd)
}

func (m *Model) commitFolderLineInput() tea.Cmd {
	if m.FolderList.InputMode() {
		return m.FolderList.CommitInput()
	}

	if m.FolderList.RenameMode() {
		return m.FolderList.CommitRename()
	}

	return nil
}

func (m *Model) handleRightClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	// 既存メニューをすべて閉じる
	m.popup.CloseAll()

	switch {
	case m.layout.folderVisible && msg.X < m.layout.folderListWidth:
		return m.rightClickFolderList(msg)
	case msg.X < m.layout.EditorStartX():
		return m.rightClickNoteList(msg, now)
	default:
		return m.rightClickEditor(msg)
	}
}

func (m *Model) rightClickNoteList(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	// クリック位置のノートを選択
	relX := m.layout.NoteListLocalX(msg.X)
	idx := m.NoteList.HitTest(relX, msg.Y, now)

	if idx < 0 {
		return nil
	}

	if !m.isTrashFolder() {
		m.syncEditorToNote(now)
	}

	m.NoteList.SelectIndex(idx, now)
	m.loadSelectedNote()

	// 既存の EditorHeader メニューをそのまま開く
	m.Editor.Header.OpenMenu()
	m.popup.SetAnchor(msg.X, msg.Y)

	return nil
}

func (m *Model) rightClickFolderList(msg tea.MouseClickMsg) tea.Cmd {
	// クリック位置のフォルダを選択
	idx := m.FolderList.HitTest(msg.X, msg.Y)
	if idx >= 0 {
		m.FolderList.SelectIndex(idx)
	}

	if !m.FolderList.IsUserFolder() {
		return nil
	}

	// 既存の FolderList メニューをそのまま開く
	m.FolderList.OpenMenu()
	m.popup.SetAnchor(msg.X, msg.Y)

	return nil
}

func (m *Model) rightClickEditor(msg tea.MouseClickMsg) tea.Cmd {
	if m.Editor.ReadOnly() {
		return nil
	}

	m.Editor.OpenContextMenu()
	m.popup.SetAnchor(msg.X, msg.Y)

	return nil
}

func (m *Model) handleClickInner(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	// モーダル状態の処理（優先度順）
	if cmd, handled := m.handleModalClick(msg, now); handled {
		return cmd
	}

	// 検索フォーカス中にエディタヘッダー以外をクリックしたらフォーカス解除
	if m.Editor.Header.SearchFocused() {
		isEditorHeader := msg.Y == 0 && msg.X >= m.layout.EditorStartX()

		if !isEditorHeader {
			m.Editor.Header.SetSearchFocused(false)
		}
	}

	return m.handleZoneClick(msg, now)
}

func (m *Model) handleModalClick(msg tea.MouseClickMsg, now time.Time) (tea.Cmd, bool) {
	// 右クリックメニュー（アンカー付き）が開いている場合
	if m.popup.HasAnchor() {
		return m.popup.HandleAnchoredClick(msg), true
	}

	// ヘルプオーバーレイ表示中（✕ボタンのみで閉じる）
	if m.helpOverlay != nil {
		if m.helpOverlay.CloseButtonHit(msg.X, msg.Y) {
			m.helpOverlay = nil
		}

		return nil, true
	}

	// 確認ダイアログ表示中
	if m.confirmDialog != nil {
		return m.handleConfirmDialogClick(msg), true
	}

	// 固定位置メニューが開いている場合（Footer / EditorHeader / MoveMenu / FolderList）
	return m.popup.HandleFixedClick(msg, now)
}

func (m *Model) handleZoneClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	zone := m.layout.HitTest(msg.X, msg.Y)

	switch zone { //nolint:exhaustive // コンポーネントゾーンは routeMouse で処理
	case ZoneFooterLabel:
		return m.handleFooterClick(msg.X, now)
	case ZoneFooterBorder:
		return nil // フッターの罫線行
	case ZoneFolderSeparator:
		m.resizingFolder = true

		return nil
	case ZoneNoteSeparator:
		m.resizing = true

		return nil
	case ZoneFolderList:
		m.Focus = FocusFolderList

		return m.routeMouse(msg, now)
	default:
		return m.routeMouse(msg, now)
	}
}

func (m *Model) handleFooterClick(x int, now time.Time) tea.Cmd {
	m.rebuildFooterButtons()

	return m.processFooterCmd(m.Footer.HandleClick(x), now)
}

// mouseMsg はマウス位置を持つメッセージ。
type mouseMsg interface {
	Mouse() tea.Mouse
}

// routeMouse はマウス位置に応じたコンポーネントに msg を委譲する。
func (m *Model) routeMouse(msg mouseMsg, now time.Time) tea.Cmd {
	var cmd tea.Cmd

	x := msg.Mouse().X
	noteListStart := m.layout.NoteListOffset()
	noteListEnd := m.layout.EditorStartX()

	switch {
	case x < noteListStart:
		m.FolderList, cmd = m.FolderList.Update(msg)

		return m.processFolderListCmd(cmd, now)
	case x < noteListEnd:
		m.NoteList, cmd = m.NoteList.Update(msg, now, m.Editor.Header.TrashMode())

		return m.processNoteListCmd(cmd, now)
	default:
		m.Editor, cmd = m.Editor.Update(msg, now)

		return m.processEditorCmd(cmd, now)
	}
}

func (m *Model) handleDrag(msg tea.MouseMotionMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()

	if m.helpOverlay != nil {
		m.helpOverlay.SetCloseHover(m.helpOverlay.CloseButtonHit(mouse.X, mouse.Y))

		return nil
	}

	m.hoverSeparator = m.resizing || m.layout.IsOnSeparator(mouse.X)
	m.hoverFolderSep = m.resizingFolder || (m.layout.folderVisible && m.layout.IsOnFolderSeparator(mouse.X))

	if m.resizingFolder {
		maxFolderWidth := m.layout.width - m.layout.noteListWidth - minEditorWidth
		newWidth := max(mouse.X, minFolderListWidth)
		newWidth = min(newWidth, maxFolderWidth)
		m.layout.folderListWidth = newWidth
		m.recalcLayout(now)

		return nil
	}

	if m.resizing {
		newWidth := max(mouse.X-m.layout.NoteListOffset(), minNoteListWidth)
		newWidth = min(newWidth, m.layout.MaxNoteListWidth())
		m.layout.noteListWidth = newWidth
		m.recalcLayout(now)

		return nil
	}

	if m.Focus == FocusEditor && m.Editor.Selecting() {
		m.Editor, _ = m.Editor.Update(msg, now)

		return nil
	}

	if !m.popup.HandleHover(mouse) {
		m.updateFolderListHeaderHover(mouse)
		m.updateEditorHeaderHover(mouse)
	}

	m.updateConfirmDialogHover(mouse)
	m.updateNoteListFolderBtnHover(mouse)
	m.updateFooterHover(mouse)

	return nil
}

func (m *Model) updateConfirmDialogHover(mouse tea.Mouse) {
	if m.confirmDialog == nil {
		return
	}

	m.confirmDialog.HandleMotionAbs(mouse.X, mouse.Y)
}

func (m *Model) updateNoteListFolderBtnHover(mouse tea.Mouse) {
	if m.FolderList.Visible() {
		m.NoteList.SetHoverFolderBtn(false)

		return
	}

	offset := m.layout.NoteListOffset()
	m.NoteList.SetHoverFolderBtn(mouse.Y == 0 && mouse.X == offset+1)
}

func (m *Model) handleRelease(msg tea.MouseReleaseMsg, now time.Time) tea.Cmd {
	if m.resizingFolder {
		m.resizingFolder = false

		return nil
	}

	if m.resizing {
		m.resizing = false

		return nil
	}

	m.Editor, _ = m.Editor.Update(msg, now)

	return nil
}

func (m *Model) handleHover(msg tea.MouseMsg) tea.Cmd {
	mouse := msg.Mouse()
	m.hoverSeparator = m.layout.IsOnSeparator(mouse.X)
	m.hoverFolderSep = m.layout.folderVisible && m.layout.IsOnFolderSeparator(mouse.X)

	if m.helpOverlay != nil {
		m.helpOverlay.SetCloseHover(m.helpOverlay.CloseButtonHit(mouse.X, mouse.Y))

		return nil
	}

	if !m.popup.HandleHover(mouse) {
		m.updateFolderListHeaderHover(mouse)
		m.updateEditorHeaderHover(mouse)
	}

	m.updateConfirmDialogHover(mouse)
	m.updateNoteListFolderBtnHover(mouse)
	m.updateFooterHover(mouse)

	return nil
}

func (m *Model) updateFolderListHeaderHover(mouse tea.Mouse) {
	if !m.FolderList.Visible() {
		return
	}

	m.FolderList.HandleHoverLocal(mouse.X, mouse.Y)
}

func (m *Model) updateEditorHeaderHover(mouse tea.Mouse) {
	editorStartX := m.layout.EditorStartX()
	if mouse.X >= editorStartX {
		m.Editor.HandleHover(mouse.X-editorStartX, mouse.Y)
	} else {
		m.Editor.Header.ClearHover()
	}
}

func (m *Model) updateFooterHover(mouse tea.Mouse) {
	if mouse.Y == m.layout.FooterLabelY() {
		m.rebuildFooterButtons()
		m.Footer.SetHover(m.Footer.HitTest(mouse.X))
	} else {
		m.Footer.SetHover(HoverNone)
	}
}

func (m *Model) processFolderListCmd(cmd tea.Cmd, now time.Time) tea.Cmd {
	if cmd == nil {
		return nil
	}

	rawMsg := cmd()

	switch msg := rawMsg.(type) {
	case FolderListMsg:
		return m.handleFolderListMsg(msg, now)
	case folderMenuActionMsg:
		return m.handleFolderMenuAction(msg.idx, now)
	case folderCreateMsg:
		return m.handleFolderCreate(msg)
	case folderRenameMsg:
		return m.handleFolderRename(msg)
	default:
		return cmd
	}
}

func (m *Model) handleFolderListMsg(msg FolderListMsg, now time.Time) tea.Cmd {
	switch msg {
	case FolderListSelect:
		return m.handleFolderSelect(now)
	case FolderListFocusNext:
		m.Focus = FocusNoteList

		return nil
	case FolderListMenu:
		if m.FolderList.IsUserFolder() {
			m.FolderList.OpenMenu()
		}

		return nil
	case FolderListClose:
		return m.toggleFolderList(now)
	case FolderListStartInput:
		blinkCmd := m.FolderList.StartInput()
		m.Focus = FocusFolderList

		return blinkCmd
	case FolderListQuit:
		m.syncEditorToNote(now)

		return tea.Quit
	case FolderListHelp:
		m.openHelp()

		return nil
	}

	return nil
}

func (m *Model) processNoteListCmd(cmd tea.Cmd, now time.Time) tea.Cmd {
	if cmd == nil {
		return nil
	}

	msg, ok := cmd().(NoteListMsg)
	if !ok {
		return cmd
	}

	return m.dispatchNoteListMsg(msg, now)
}

func (m *Model) dispatchNoteListMsg(msg NoteListMsg, now time.Time) tea.Cmd { //nolint:cyclop // msg種別ごとの分岐
	switch msg {
	case NoteListSelect:
		m.loadSelectedNote()

		return nil
	case NoteListClickSelect:
		if !m.isTrashFolder() {
			m.syncEditorToNote(now)
		}

		m.loadSelectedNote()
		m.Focus = FocusNoteList
		m.Editor.Blur()

		return nil
	case NoteListCreate:
		return m.createNote(now)
	case NoteListTrash:
		return m.trashNote(now)
	case NoteListUndo:
		return m.undoNote(now)
	case NoteListRedo:
		return m.redoNote(now)
	case NoteListEdit:
		return m.focusEditor()
	case NoteListDuplicate:
		return m.duplicateNote(now)
	case NoteListCopy:
		return m.copyNote()
	case NoteListMenu:
		return m.openNoteListMenu(now)
	case NoteListQuit:
		m.syncEditorToNote(now)

		return tea.Quit
	case NoteListFocusPrev:
		if m.FolderList.Visible() {
			m.Focus = FocusFolderList
		}

		return nil
	case NoteListToggleFolder:
		return m.toggleFolderList(now)
	case NoteListHelp:
		m.openHelp()

		return nil
	}

	return nil
}

func (m *Model) openNoteListMenu(now time.Time) tea.Cmd {
	m.Editor.Header.OpenMenu()
	menuW := m.Editor.Header.PopupMenu.Width()
	x := m.layout.EditorStartX() - menuW
	y := m.NoteList.SelectedY(now)
	m.popup.SetAnchor(x, y)

	return nil
}

func (m *Model) processEditorCmd(cmd tea.Cmd, now time.Time) tea.Cmd { //nolint:cyclop // msg種別ごとの分岐
	if cmd == nil {
		return nil
	}

	rawMsg := cmd()

	switch msg := rawMsg.(type) {
	case EditorMsg:
		switch msg {
		case EditorBlur:
			return m.blurEditor(now)
		case EditorSave:
			m.syncEditorToNote(now)

			return m.setInfoMsg("Saved")
		case EditorSearchChanged:
			return m.scheduleSearchDebounce()
		case EditorSearchBlur:
			m.Focus = FocusNoteList

			return nil
		case EditorClickBody:
			m.Focus = FocusEditor

			return m.Editor.Focus()
		}
	case EditorHeaderMsg:
		return m.handleEditorHeaderMsg(msg, now)
	case editorOpenURLMsg:
		return openURLInBrowser(msg.URL)
	default:
		return cmd
	}

	return nil
}

func (m *Model) handleEditorHeaderMsg(msg EditorHeaderMsg, now time.Time) tea.Cmd {
	switch msg {
	case EditorHeaderNew:
		return m.createNote(now)
	case EditorHeaderTrash:
		return m.trashNote(now)
	case EditorHeaderCopy:
		return m.copyNote()
	case EditorHeaderPin:
		return m.pinNote()
	case EditorHeaderUnpin:
		return m.unpinNote()
	case EditorHeaderMove:
		return m.openMoveMenu()
	case EditorHeaderDuplicate:
		return m.duplicateNote(now)
	}

	return nil
}

func (m *Model) processEditorHeaderCmd(cmd tea.Cmd, now time.Time) tea.Cmd {
	if cmd == nil {
		return nil
	}

	msg, ok := cmd().(EditorHeaderMsg)
	if !ok {
		return cmd
	}

	return m.handleEditorHeaderMsg(msg, now)
}

func (m *Model) processFooterCmd(cmd tea.Cmd, now time.Time) tea.Cmd {
	if cmd == nil {
		return nil
	}

	msg, ok := cmd().(FooterMsg)
	if !ok {
		return cmd
	}

	switch msg {
	case FooterQuit:
		m.syncEditorToNote(now)

		return tea.Quit
	case FooterMore:
		return nil
	case FooterHelp:
		m.openHelp()

		return nil
	}

	return nil
}

// --- アクション ---

func (m *Model) focusEditor() tea.Cmd {
	if m.isTrashFolder() {
		return nil
	}

	if m.Editor.NoteID() == "" {
		return nil
	}

	m.Focus = FocusEditor

	return m.Editor.Focus()
}

func (m *Model) blurEditor(now time.Time) tea.Cmd {
	m.syncEditorToNote(now)
	m.Focus = FocusNoteList
	m.Editor.Blur()

	return nil
}

func (m *Model) recalcLayout(now time.Time) {
	m.layout.folderVisible = m.FolderList.Visible()
	bodyHeight := m.layout.BodyHeight()

	if m.layout.folderVisible {
		folderW := max(m.layout.folderListWidth, 0)
		m.FolderList.SetSize(folderW, bodyHeight)
	}

	noteW := max(m.layout.noteListWidth, 0)
	editorW := m.layout.EditorWidth()

	m.NoteList.SetSize(noteW, bodyHeight, now)
	m.Editor.SetSize(editorW, bodyHeight)
}

const searchDebounceDuration = 150 * time.Millisecond

func (m *Model) scheduleSearchDebounce() tea.Cmd {
	m.searchDebounceID++
	id := m.searchDebounceID
	query := m.Editor.Header.SearchQuery()

	return tea.Tick(searchDebounceDuration, func(_ time.Time) tea.Msg {
		return searchDebounceMsg{id: id, query: query}
	})
}

func (m *Model) applySearchFilter(query string, now time.Time) {
	folderName := m.currentFolderName()

	if query == "" {
		m.NoteList.SetNotes(m.App.ListByFolder(folderName), now)
	} else {
		results := m.App.SearchByFolder(folderName, query)
		m.NoteList.SetNotes(results, now)
	}

	m.NoteList.SetSearchQuery(query)
	m.Editor.SetSearchQuery(query)

	// 最初のノートを選択してエディタに読み込む
	if _, ok := m.NoteList.SelectedNote(); ok {
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}
}

func (m *Model) currentFolderName() string {
	if m.isTrashFolder() {
		return note.TrashDir
	}

	name := m.FolderList.SelectedName()
	if name == "" {
		return app.DefaultFolder
	}

	return name
}
