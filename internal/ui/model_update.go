package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
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
		return m, m.handleRelease()
	case tea.MouseWheelMsg:
		return m, m.handleWheel(msg, now)
	case tea.MouseMsg:
		return m, m.handleHover(msg)
	case FolderListMsg:
		return m, m.processFolderListMsg(msg, now)
	case folderCreateMsg:
		return m, m.handleFolderCreate(msg)
	case folderRenameMsg:
		return m, m.handleFolderRename(msg)
	case noteMoveMsg:
		return m, m.handleNoteMove(msg, now)
	case folderDeleteMsg:
		return m, m.handleFolderDelete(msg, now)
	case NoteListMsg:
		return m, m.processNoteListMsg(msg, now)
	case EditorMsg:
		return m, m.processEditorMsg(msg, now)
	case editorOpenURLMsg:
		return m, openURLInBrowser(msg.URL)
	case EditorHeaderMsg:
		return m, m.processEditorHeaderMsg(msg, now)
	case FooterMsg:
		return m, m.processFooterMsg(msg, now)
	case tea.FocusMsg:
		return m, m.handleFocusRestore()
	case tea.BlurMsg: // 他アプリへ切り替え時に編集中の内容を保存
		m.syncEditorToNote(now)

		return m, nil
	case cursorBlinkMsg:
		switch msg.owner {
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
		return m, m.handleDefault(msg, now)
	}
}

func (m *Model) handleKey(msg tea.KeyPressMsg, now time.Time) tea.Cmd { //nolint:cyclop,gocyclo,funlen // key dispatch
	// メニューが開いている場合（右クリック or キーボード起動）
	if menu, kind := m.popup.Active(); menu != nil {
		return m.popup.HandleKey(msg, menu, kind, now)
	}

	// 確認ダイアログ表示中
	if m.confirmDialog != nil {
		return m.handleConfirmDialogKey(msg)
	}

	// ヘルプオーバーレイ表示中
	if m.helpOverlay != nil {
		if msg.Code == 'q' && msg.Mod&tea.ModCtrl != 0 {
			m.helpOverlay = nil
			m.syncEditorToNote(now)

			return tea.Quit
		}

		if m.helpOverlay.Update(msg) == HelpClose {
			m.helpOverlay = nil
		}

		return nil
	}

	m.errMsg = ""
	m.infoMsg = ""
	m.Footer.CloseMenu()
	m.Editor.Header.CloseMenu()
	m.Editor.Header.CloseMoveMenu()

	switch {
	case msg.Code == 'q' && msg.Mod&tea.ModCtrl != 0:
		m.syncEditorToNote(now)

		return tea.Quit
	case msg.Code == 'q' && (m.Focus == FocusNoteList || m.Focus == FocusFolderList):
		m.syncEditorToNote(now)

		return tea.Quit
	case msg.Code == tea.KeyTab && m.Focus == FocusNoteList:
		return m.focusEditor()
	case msg.Code == tea.KeyTab && m.Focus == FocusFolderList:
		m.Focus = FocusNoteList

		return nil
	case msg.Code == tea.KeyEscape && m.Focus == FocusNoteList && m.FolderList.Visible():
		m.Focus = FocusFolderList

		return nil
	case msg.Code == 'b' && msg.Mod&tea.ModCtrl != 0 &&
		m.Focus != FocusEditor &&
		!m.FolderList.InputMode() && !m.FolderList.RenameMode():
		return m.toggleFolderList(now)
	case msg.Code == 'f' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		m.Editor.Header.SetSearchFocused(true)
		m.Focus = FocusEditor

		return m.Editor.Header.searchBlink.Reset()
	case msg.Code == '/' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		m.helpOverlay = NewHelpOverlay(m.Focus)

		return nil
	case msg.Code == '?' && msg.Mod == 0 && m.Focus != FocusEditor:
		m.helpOverlay = NewHelpOverlay(m.Focus)

		return nil
	}

	switch m.Focus {
	case FocusFolderList:
		return m.handleFolderListKey(msg, now)
	case FocusNoteList:
		_, cmd := m.NoteList.Update(msg, now, m.Editor.Header.TrashMode())

		return m.processNoteListCmd(cmd, now)
	case FocusEditor:
		// 検索フィールドにフォーカスがある場合
		if m.Editor.Header.SearchFocused() {
			handled, _ := m.Editor.Header.HandleSearchKey(msg)
			if handled {
				if !m.Editor.Header.SearchFocused() {
					// Esc で検索フォーカスを外した → ノート一覧にフォーカス
					m.Editor.Header.searchBlink.Stop()
					m.Focus = FocusNoteList

					return nil
				}

				blinkCmd := m.Editor.Header.searchBlink.Reset()

				return tea.Batch(m.scheduleSearchDebounce(), blinkCmd)
			}
		}

		_, cmd := m.Editor.Update(msg, now)
		editorCmd := m.processEditorCmd(cmd, now)
		blinkCmd := m.Editor.resetBlink()

		return tea.Batch(editorCmd, blinkCmd)
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

func (m *Model) handleDefault(msg tea.Msg, now time.Time) tea.Cmd {
	if m.Focus == FocusEditor {
		_, cmd := m.Editor.Update(msg, now)

		return cmd
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

// processFooterMenuAction はフッターメニューのインデックスからアクションを実行する。
func (m *Model) processFooterMenuAction(idx int, now time.Time) tea.Cmd {
	cmd := m.Footer.ExecuteMenuAction(idx)
	if cmd == nil {
		return nil
	}

	msg, ok := cmd().(FooterMsg)
	if !ok {
		return cmd
	}

	return m.processFooterMsg(msg, now)
}

func (m *Model) handleClickInner(msg tea.MouseClickMsg, now time.Time) tea.Cmd { //nolint:cyclop,funlen // mouse dispatch
	// 右クリックメニュー（アンカー付き）が開いている場合
	if m.popup.HasAnchor() {
		return m.popup.HandleAnchoredClick(msg)
	}

	// ヘルプオーバーレイ表示中（✕ボタンのみで閉じる）
	if m.helpOverlay != nil {
		if m.helpCloseButtonHit(msg.X, msg.Y) {
			m.helpOverlay = nil
		}

		return nil
	}

	// 確認ダイアログ表示中
	if m.confirmDialog != nil {
		return m.handleConfirmDialogClick(msg)
	}

	// フッターメニューが開いている場合
	if m.Footer.MenuOpen() {
		return m.handleClickWithMenu(msg, now)
	}

	// 移動先メニューが開いている場合
	if m.Editor.Header.MoveMenuOpen() {
		edX := m.layout.EditorLocalX(msg.X)
		menuTopY := editorHeaderMenuTopY
		menuHeight := m.Editor.Header.MoveMenuHeight()
		menuX := m.Editor.Header.MoveMenuLeftX()
		menuWidth := m.Editor.Header.MoveMenu.Width()

		if msg.Y >= menuTopY && msg.Y < menuTopY+menuHeight && edX >= menuX && edX < menuX+menuWidth {
			relX := edX - menuX
			relY := msg.Y - menuTopY

			return m.Editor.Header.HandleMoveMenuClick(relX, relY)
		}

		m.Editor.Header.CloseMoveMenu()

		return nil
	}

	// エディタヘッダーメニューが開いている場合
	if m.Editor.IsHeaderMenuOpen() {
		edX := m.layout.EditorLocalX(msg.X)
		cmd := m.Editor.HandleClick(edX, msg.Y)

		return m.processEditorHeaderCmd(cmd, now)
	}

	// 検索フォーカス中にエディタヘッダー以外をクリックしたらフォーカス解除
	if m.Editor.Header.SearchFocused() {
		isEditorHeader := msg.Y == 0 && msg.X >= m.layout.EditorStartX()

		if !isEditorHeader {
			m.Editor.Header.SetSearchFocused(false)
		}
	}

	zone := m.layout.HitTest(msg.X, msg.Y)

	switch zone { //nolint:exhaustive // EditorHeader/EditorBody は default で処理
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
		return m.handleFolderListClick(msg, now)
	case ZoneNoteList:
		return m.handleNoteListClick(msg, now)
	default:
		return m.handleEditorClick(msg)
	}
}

func (m *Model) handleClickWithMenu(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	menuHeight := m.Footer.MenuHeight()
	bodyLines := m.layout.BodyHeight()
	menuTopY := bodyLines - menuHeight

	// メニュー領域内のクリック
	if msg.Y >= menuTopY && msg.Y < menuTopY+menuHeight {
		relX := msg.X - 1        // 先頭スペース分を引く
		relY := msg.Y - menuTopY // PopupMenu 座標 (0=上枠, 1=項目1, ...)

		cmd := m.Footer.HandleMenuClick(relX, relY)
		if cmd == nil {
			return nil
		}

		fMsg, ok := cmd().(FooterMsg)
		if !ok {
			return cmd
		}

		return m.processFooterMsg(fMsg, now)
	}

	// メニュー外クリック → メニューを閉じるだけ
	m.Footer.CloseMenu()

	return nil
}

func (m *Model) handleFooterClick(x int, now time.Time) tea.Cmd {
	m.rebuildFooterButtons()

	cmd := m.Footer.HandleClick(x)
	if cmd == nil {
		return nil
	}

	msg := cmd()

	fMsg, ok := msg.(FooterMsg)
	if !ok {
		return cmd
	}

	return m.processFooterMsg(fMsg, now)
}

func (m *Model) handleNoteListClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	// NoteList のトグルボタン（≡）クリック判定
	nlOffset := m.layout.NoteListOffset()
	if !m.FolderList.Visible() && msg.Y == 0 && msg.X >= nlOffset+1 && msg.X <= nlOffset+2 {
		return m.toggleFolderList(now)
	}

	relX := m.layout.NoteListLocalX(msg.X)
	idx := m.NoteList.HitTest(relX, msg.Y, now)

	if idx >= 0 {
		if !m.isTrashFolder() {
			m.syncEditorToNote(now)
		}

		m.NoteList.SelectIndex(idx, now)
		m.loadSelectedNote()
	}

	m.Focus = FocusNoteList
	m.Editor.Blur()

	return nil
}

func (m *Model) handleEditorClick(msg tea.MouseClickMsg) tea.Cmd {
	edX := m.layout.EditorLocalX(msg.X)

	// ヘッダー行のクリック
	if msg.Y == 0 {
		cmd := m.Editor.HandleClick(edX, 0)

		if m.Editor.Header.SearchFocused() {
			m.Focus = FocusEditor

			blinkCmd := m.Editor.Header.searchBlink.Reset()

			return tea.Batch(m.processEditorHeaderCmd(cmd, time.Now()), blinkCmd)
		}

		return m.processEditorHeaderCmd(cmd, time.Now())
	}

	if m.isTrashFolder() || m.Editor.NoteID() == "" {
		return nil
	}

	m.Focus = FocusEditor
	cmd := m.Editor.Focus()

	// textarea 領域はヘッダー分だけ Y を補正
	m.Editor.HandleTextAreaClick(edX, msg.Y-editorHeaderHeight, time.Now())

	return cmd
}

func (m *Model) handleFolderListKey(msg tea.KeyPressMsg, now time.Time) tea.Cmd {
	_, cmd := m.FolderList.Update(msg)
	folderCmd := m.processFolderListCmd(cmd, now)

	if m.FolderList.InputMode() || m.FolderList.RenameMode() {
		blinkCmd := m.FolderList.blink.Reset()

		return tea.Batch(folderCmd, blinkCmd)
	}

	return folderCmd
}

func (m *Model) handleFolderListClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	cmd := m.FolderList.HandleClickLocal(msg.X, msg.Y)
	m.Focus = FocusFolderList

	return m.processFolderListCmd(cmd, now)
}

func (m *Model) handleWheel(msg tea.MouseWheelMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()

	const scrollLines = 1

	noteListStart := m.layout.NoteListOffset()
	noteListEnd := m.layout.EditorStartX()

	if mouse.X < noteListStart {
		// フォルダ一覧領域 — スクロール不要（項目が少ないため）
		return nil
	}

	if mouse.X < noteListEnd {
		switch mouse.Button {
		case tea.MouseWheelUp:
			m.NoteList.ScrollUp(scrollLines, now)
		case tea.MouseWheelDown:
			m.NoteList.ScrollDown(scrollLines, now)
		}

		return nil
	}

	switch mouse.Button {
	case tea.MouseWheelUp:
		m.Editor.ScrollUp(scrollLines)
	case tea.MouseWheelDown:
		m.Editor.ScrollDown(scrollLines)
	}

	return nil
}

func (m *Model) handleDrag(msg tea.MouseMotionMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()

	if m.helpOverlay != nil {
		m.helpOverlay.SetCloseHover(m.helpCloseButtonHit(mouse.X, mouse.Y))

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
		edX := m.layout.EditorLocalX(mouse.X)
		m.Editor.UpdateDragSelection(edX, mouse.Y-editorHeaderHeight)

		return nil
	}

	if m.popup.HasAnchor() {
		m.popup.HandleHover(mouse)
	} else {
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

	originX, originY := m.confirmDialogOrigin()
	relX := mouse.X - originX
	relY := mouse.Y - originY

	m.confirmDialog.HandleMotion(relX, relY)
}

func (m *Model) updateNoteListFolderBtnHover(mouse tea.Mouse) {
	if m.FolderList.Visible() {
		m.NoteList.SetHoverFolderBtn(false)

		return
	}

	offset := m.layout.NoteListOffset()
	m.NoteList.SetHoverFolderBtn(mouse.Y == 0 && mouse.X == offset+1)
}

func (m *Model) handleRelease() tea.Cmd {
	if m.resizingFolder {
		m.resizingFolder = false

		return nil
	}

	if m.resizing {
		m.resizing = false

		return nil
	}

	if m.Editor.Selecting() {
		m.Editor.StopDragSelection()
	}

	return nil
}

func (m *Model) handleHover(msg tea.MouseMsg) tea.Cmd {
	mouse := msg.Mouse()
	m.hoverSeparator = m.layout.IsOnSeparator(mouse.X)
	m.hoverFolderSep = m.layout.folderVisible && m.layout.IsOnFolderSeparator(mouse.X)

	if m.helpOverlay != nil {
		m.helpOverlay.SetCloseHover(m.helpCloseButtonHit(mouse.X, mouse.Y))

		return nil
	}

	if m.popup.HasAnchor() {
		m.popup.HandleHover(mouse)
	} else {
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

	if m.Editor.Header.MoveMenuOpen() {
		edX := mouse.X - editorStartX
		menuTopY := editorHeaderMenuTopY
		menuHeight := m.Editor.Header.MoveMenuHeight()
		menuX := m.Editor.Header.MoveMenuLeftX()
		menuWidth := m.Editor.Header.MoveMenu.Width()

		if mouse.Y >= menuTopY && mouse.Y < menuTopY+menuHeight && edX >= menuX && edX < menuX+menuWidth {
			m.Editor.Header.SetMoveMenuHover(edX-menuX, mouse.Y-menuTopY)
		} else {
			m.Editor.Header.SetMoveMenuHover(-1, -1)
		}
	}
}

func (m *Model) updateFooterHover(mouse tea.Mouse) {
	footerLabelY := m.layout.FooterLabelY()

	if m.Footer.MenuOpen() {
		menuHeight := m.Footer.MenuHeight()
		bodyLines := m.layout.BodyHeight()
		menuTopY := bodyLines - menuHeight

		if mouse.Y >= menuTopY && mouse.Y < menuTopY+menuHeight {
			relX := mouse.X - 1
			relY := mouse.Y - menuTopY
			m.Footer.SetMenuHover(relX, relY)
			m.Footer.SetHover(HoverNone)

			return
		}
	}

	if mouse.Y == footerLabelY {
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
		return m.processFolderListMsg(msg, now)
	case folderMenuActionMsg:
		return m.handleFolderMenuAction(msg.idx)
	default:
		return cmd
	}
}

func (m *Model) processFolderListMsg(msg FolderListMsg, now time.Time) tea.Cmd {
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

	return m.processNoteListMsg(msg, now)
}

func (m *Model) processNoteListMsg(msg NoteListMsg, now time.Time) tea.Cmd { //nolint:cyclop // メッセ���ジ振り分けのため許容
	switch msg {
	case NoteListSelect:
		m.loadSelectedNote()

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
		m.Editor.Header.OpenMenu()
		menuW := m.Editor.Header.PopupMenu.Width()
		x := m.layout.EditorStartX() - menuW
		y := m.NoteList.SelectedY(now)
		m.popup.SetAnchor(x, y)

		return nil
	case NoteListQuit:
		m.syncEditorToNote(now)

		return tea.Quit
	}

	return nil
}

func (m *Model) processEditorCmd(cmd tea.Cmd, now time.Time) tea.Cmd {
	if cmd == nil {
		return nil
	}

	msg, ok := cmd().(EditorMsg)
	if !ok {
		return cmd
	}

	return m.processEditorMsg(msg, now)
}

func (m *Model) processEditorMsg(msg EditorMsg, now time.Time) tea.Cmd {
	switch msg {
	case EditorBlur:
		return m.blurEditor(now)
	case EditorSave:
		m.syncEditorToNote(now)

		return m.setInfoMsg("Saved")
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

	return m.processEditorHeaderMsg(msg, now)
}

func (m *Model) processEditorHeaderMsg(msg EditorHeaderMsg, now time.Time) tea.Cmd {
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

func (m *Model) processFooterMsg(msg FooterMsg, now time.Time) tea.Cmd {
	switch msg {
	case FooterQuit:
		m.syncEditorToNote(now)

		return tea.Quit
	case FooterMore:
		return nil
	case FooterHelp:
		m.helpOverlay = NewHelpOverlay(m.Focus)

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
