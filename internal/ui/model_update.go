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

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.handleResize(msg, now)
	case tea.KeyPressMsg:
		cmd = m.handleKey(msg, now)
	case tea.MouseClickMsg:
		cmd = m.handleClick(msg, now)
	case tea.MouseMotionMsg:
		cmd = m.handleDrag(msg, now)
	case tea.MouseReleaseMsg:
		cmd = m.handleRelease(msg, now)
	case tea.MouseWheelMsg:
		cmd = m.routeMouse(msg, now)
	case tea.MouseMsg:
		cmd = m.handleHover(msg, now)
	case tea.FocusMsg:
		cmd = m.handleFocusRestore()
	case tea.BlurMsg: // 他アプリへ切り替え時に編集中の内容を保存
		m.syncEditorToNote(now)
	case tui.CursorBlinkMsg:
		switch msg.Owner {
		case blinkOwnerEditor:
			cmd = m.Editor.HandleBlinkMsg(msg)
		case blinkOwnerFolderList:
			cmd = m.FolderList.blink.HandleMsg(msg)
		case blinkOwnerSearch:
			cmd = m.Editor.Header.searchBlink.HandleMsg(msg)
		}
	case searchDebounceMsg:
		if msg.id != m.searchDebounceID {
			return m, nil // 古いタイマーは無視
		}

		m.applySearchFilter(msg.query, now)
	case searchClearedMsg:
		m.applySearchFilter("", now)
	case clearInfoMsg:
		cmd = m.handleClearInfo(msg)
	default:
		cmd = m.routeFocus(msg, now)
	}

	m.syncViewState()

	return m, cmd
}

// syncViewState は View() に必要な派生状態を同期する。
// Update の最後に呼ばれ、View() の純粋性を保証する。
func (m *Model) syncViewState() {
	m.updateFolderCounts()

	if m.Editor.Dirty() {
		m.NoteList.SetDirtyNoteID(m.Editor.NoteID())
	} else {
		m.NoteList.SetDirtyNoteID("")
	}

	m.rebuildFooterButtons()
}

func (m *Model) updateFolderCounts() {
	notesCount := len(m.App.ListByFolder(app.DefaultFolder))

	for i := range m.FolderList.folders {
		switch m.FolderList.folders[i].Kind {
		case FolderNotes:
			m.FolderList.folders[i].Count = notesCount
		case FolderTrash:
			m.FolderList.folders[i].Count = len(m.App.ListTrashNotes())
		case FolderUser:
			count, err := m.App.FolderNoteCount(m.FolderList.folders[i].Name)
			if err == nil {
				m.FolderList.folders[i].Count = count
			}
		}
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

	if m.helpOverlay != nil {
		m.helpOverlay.SetScreenSize(m.layout.width, m.layout.BodyHeight())
	}

	if m.confirmDialog != nil {
		m.confirmDialog.SetScreenSize(m.layout.width, m.layout.BodyHeight())
	}

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
		m.popup.CloseAll()

		return m.routeMouse(msg, now)
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
	m.dragTarget = m.zoneToDragTarget(zone)

	switch zone { //nolint:exhaustive // コンポーネントゾーンは routeMouse で処理
	case ZoneFooterLabel:
		return m.handleFooterClick(msg.X, now)
	case ZoneFooterBorder:
		return nil // フッターの罫線行
	case ZoneFolderSeparator:
		return nil
	case ZoneNoteSeparator:
		return nil
	case ZoneFolderList:
		m.Focus = FocusFolderList

		return m.routeMouse(msg, now)
	default:
		return m.routeMouse(msg, now)
	}
}

// zoneToDragTarget は HitZone から dragTarget を算出する。
func (m *Model) zoneToDragTarget(zone HitZone) dragTarget {
	switch zone { //nolint:exhaustive // フッター等は dragTarget に対応しない
	case ZoneFolderSeparator:
		return dragFolderSeparator
	case ZoneNoteSeparator:
		return dragNoteSeparator
	case ZoneFolderList:
		return dragFolderList
	case ZoneNoteList:
		return dragNoteList
	case ZoneEditorHeader, ZoneEditorBody:
		return dragEditor
	default:
		return dragNone
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

// routeDragTarget はキャプチャ中の dragTarget に応じたコンポーネントに msg を委譲する。
func (m *Model) routeDragTarget(msg mouseMsg, now time.Time) tea.Cmd {
	var cmd tea.Cmd

	switch m.dragTarget { //nolint:exhaustive // セパレーター・None は呼び出し元で処理済み
	case dragFolderList:
		m.FolderList, cmd = m.FolderList.Update(msg)

		return m.processFolderListCmd(cmd, now)
	case dragNoteList:
		m.NoteList, cmd = m.NoteList.Update(msg, now, m.Editor.Header.TrashMode())

		return m.processNoteListCmd(cmd, now)
	case dragEditor:
		m.Editor, cmd = m.Editor.Update(msg, now)

		return m.processEditorCmd(cmd, now)
	}

	return nil
}

// calcHoverTarget はマウス座標から hover 対象コンポーネントを算出する。
func (m *Model) calcHoverTarget(x, y int) hoverClearer { //nolint:ireturn // 複数の具象型を返すため
	if y >= m.layout.FooterLabelY() {
		return &m.Footer
	}

	noteListStart := m.layout.NoteListOffset()
	noteListEnd := m.layout.EditorStartX()

	switch {
	case m.layout.folderVisible && x < noteListStart:
		return &m.FolderList
	case x < noteListEnd:
		return &m.NoteList
	default:
		return &m.Editor
	}
}

// updateHoverTarget は hover 対象が変わった場合に旧対象の hover をクリアする。
func (m *Model) updateHoverTarget(target hoverClearer) {
	if m.lastHovered == target {
		return
	}

	if m.lastHovered != nil {
		m.lastHovered.ClearHover()
	}

	m.lastHovered = target
}

// routeMouse はマウス位置に応じたコンポーネントに msg を委譲する。
func (m *Model) routeMouse(msg mouseMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()

	// フッター領域
	if mouse.Y >= m.layout.FooterLabelY() {
		m.rebuildFooterButtons()
		m.Footer.SetHover(m.Footer.HitTest(mouse.X))

		return nil
	}

	var cmd tea.Cmd

	x := mouse.X
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

func (m *Model) handleDrag(msg tea.MouseMotionMsg, now time.Time) tea.Cmd { //nolint:cyclop // dragTarget dispatch
	mouse := msg.Mouse()

	if m.helpOverlay != nil {
		m.helpOverlay.SetCloseHover(m.helpOverlay.CloseButtonHit(mouse.X, mouse.Y))

		return nil
	}

	m.hoverSeparator = m.dragTarget == dragNoteSeparator || m.layout.IsOnSeparator(mouse.X)
	m.hoverFolderSep = m.dragTarget == dragFolderSeparator ||
		(m.layout.folderVisible && m.layout.IsOnFolderSeparator(mouse.X))

	switch m.dragTarget {
	case dragFolderSeparator:
		maxFolderWidth := m.layout.width - m.layout.noteListWidth - minEditorWidth
		newWidth := max(mouse.X, minFolderListWidth)
		newWidth = min(newWidth, maxFolderWidth)
		m.layout.folderListWidth = newWidth
		m.recalcLayout(now)

		return nil
	case dragNoteSeparator:
		newWidth := max(mouse.X-m.layout.NoteListOffset(), minNoteListWidth)
		newWidth = min(newWidth, m.layout.MaxNoteListWidth())
		m.layout.noteListWidth = newWidth
		m.recalcLayout(now)

		return nil
	case dragFolderList, dragNoteList, dragEditor:
		return m.routeDragTarget(msg, now)
	case dragNone:
		if m.confirmDialog != nil {
			m.confirmDialog.HandleMotionAbs(mouse.X, mouse.Y)
		}

		newTarget := m.calcHoverTarget(mouse.X, mouse.Y)
		m.updateHoverTarget(newTarget)

		if !m.popup.HandleHover(mouse) {
			return m.routeMouse(msg, now)
		}

		return nil
	}

	return nil
}

func (m *Model) handleRelease(msg tea.MouseReleaseMsg, now time.Time) tea.Cmd {
	defer func() { m.dragTarget = dragNone }()

	switch m.dragTarget {
	case dragFolderSeparator, dragNoteSeparator:
		return nil
	case dragFolderList, dragNoteList, dragEditor:
		return m.routeDragTarget(msg, now)
	case dragNone:
		return nil
	}

	return nil
}

func (m *Model) handleHover(msg tea.MouseMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()
	m.hoverSeparator = m.layout.IsOnSeparator(mouse.X)
	m.hoverFolderSep = m.layout.folderVisible && m.layout.IsOnFolderSeparator(mouse.X)

	if m.helpOverlay != nil {
		m.helpOverlay.SetCloseHover(m.helpOverlay.CloseButtonHit(mouse.X, mouse.Y))

		return nil
	}

	if m.confirmDialog != nil {
		m.confirmDialog.HandleMotionAbs(mouse.X, mouse.Y)
	}

	newTarget := m.calcHoverTarget(mouse.X, mouse.Y)
	m.updateHoverTarget(newTarget)

	if !m.popup.HandleHover(mouse) {
		return m.routeMouse(msg, now)
	}

	return nil
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
	case FolderListRightClickMsg:
		m.popup.SetAnchor(msg.AnchorX, msg.AnchorY)

		return nil
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

	rawMsg := cmd()

	switch msg := rawMsg.(type) {
	case NoteListMsg:
		return m.dispatchNoteListMsg(msg, now)
	case NoteListRightClickMsg:
		return m.handleNoteListRightClick(msg, now)
	default:
		return cmd
	}
}

func (m *Model) handleNoteListRightClick(msg NoteListRightClickMsg, now time.Time) tea.Cmd {
	if !m.isTrashFolder() {
		m.syncEditorToNote(now)
	}

	m.loadSelectedNote()
	m.Editor.Header.OpenMenu()
	m.popup.SetAnchor(msg.AnchorX, msg.AnchorY)

	return nil
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
	case EditorRightClickMsg:
		m.popup.SetAnchor(msg.AnchorX, msg.AnchorY)

		return nil
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
