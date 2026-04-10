package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

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
		cmd := m.Editor.HandleBlinkMsg(msg)

		return m, cmd
	case clearInfoMsg:
		return m, m.handleClearInfo(msg)
	default:
		return m, m.handleDefault(msg, now)
	}
}

func (m *Model) handleKey(msg tea.KeyPressMsg, now time.Time) tea.Cmd { //nolint:cyclop // key dispatch
	// 確認ダイアログ表示中
	if m.confirmDialog != nil {
		return m.handleConfirmDialogKey(msg)
	}

	m.errMsg = ""
	m.infoMsg = ""
	m.Footer.CloseMenu()
	m.Editor.Header.CloseMenu()
	m.Editor.Header.CloseMoveMenu()

	switch {
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
	case msg.Code == 'b' && msg.Mod&tea.ModCtrl != 0 && m.Focus != FocusEditor:
		return m.toggleFolderList(now)
	}

	switch m.Focus {
	case FocusFolderList:
		_, cmd := m.FolderList.Update(msg)

		return m.processFolderListCmd(cmd, now)
	case FocusNoteList:
		_, cmd := m.NoteList.Update(msg, now, m.App.TrashMode)

		return m.processNoteListCmd(cmd, now)
	case FocusEditor:
		_, cmd := m.Editor.Update(msg, now)
		editorCmd := m.processEditorCmd(cmd, now)
		blinkCmd := m.Editor.resetBlink()

		return tea.Batch(editorCmd, blinkCmd)
	}

	return nil
}

func (m *Model) handleResize(msg tea.WindowSizeMsg, now time.Time) tea.Cmd {
	m.width = msg.Width
	m.height = msg.Height
	m.noteListWidth = max(m.noteListWidth, minNoteListWidth)
	m.noteListWidth = min(m.noteListWidth, m.maxNoteListWidth())

	if m.FolderList.Visible() {
		maxNoteWidth := m.width - m.folderListWidth - minEditorWidth
		m.noteListWidth = min(m.noteListWidth, maxNoteWidth)
		m.noteListWidth = max(m.noteListWidth, minNoteListWidth)

		maxFolderWidth := m.width - m.noteListWidth - minEditorWidth
		m.folderListWidth = max(m.folderListWidth, minFolderListWidth)
		m.folderListWidth = min(m.folderListWidth, maxFolderWidth)
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

func (m *Model) handleClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd { //nolint:cyclop,funlen // mouse dispatch
	m.errMsg = ""

	if msg.Button != tea.MouseLeft {
		return nil
	}

	// 確認ダイアログ表示中
	if m.confirmDialog != nil {
		return m.handleConfirmDialogClick(msg)
	}

	footerLabelY := m.height - footerLineCount + 1 // フッター3行の中央（ラベル行）

	// フッターメニューが開いている場合
	if m.Footer.MenuOpen() {
		return m.handleClickWithMenu(msg, now)
	}

	// 移動先メニューが開いている場合
	if m.Editor.Header.MoveMenuOpen() {
		edX := msg.X - m.noteListOffset() - m.noteListWidth
		menuTopY := editorHeaderMenuTopY
		menuHeight := m.Editor.Header.MoveMenuHeight()
		menuWidth := m.Editor.Header.MoveMenu.Width()
		menuX := m.Editor.Header.Width() - moreButtonOffset + 1 - menuWidth

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
		edX := msg.X - m.noteListOffset() - m.noteListWidth
		cmd := m.Editor.HandleClick(edX, msg.Y)

		return m.processEditorHeaderCmd(cmd, now)
	}

	switch {
	case msg.Y == footerLabelY:
		return m.handleFooterClick(msg.X, now)
	case msg.Y >= m.height-footerLineCount:
		return nil // フッターの罫線行
	case m.FolderList.Visible() && m.isOnFolderSeparator(msg.X):
		m.resizingFolder = true

		return nil
	case m.isOnSeparator(msg.X):
		m.resizing = true

		return nil
	case m.FolderList.Visible() && msg.X < m.folderListWidth:
		return m.handleFolderListClick(msg, now)
	case msg.X < m.noteListOffset()+m.noteListWidth:
		return m.handleNoteListClick(msg, now)
	default:
		return m.handleEditorClick(msg)
	}
}

func (m *Model) handleClickWithMenu(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	menuHeight := m.Footer.MenuHeight()
	bodyLines := m.height - footerLineCount
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
	if !m.FolderList.Visible() && msg.Y == 0 && msg.X >= m.noteListOffset()+1 && msg.X <= m.noteListOffset()+2 {
		return m.toggleFolderList(now)
	}

	relX := msg.X - m.noteListOffset()
	idx := m.NoteList.HitTest(relX, msg.Y, now)

	if idx >= 0 {
		if !m.App.TrashMode {
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
	edX := msg.X - m.noteListOffset() - m.noteListWidth

	// ヘッダー行のクリック
	if msg.Y == 0 {
		cmd := m.Editor.HandleClick(edX, 0)

		return m.processEditorHeaderCmd(cmd, time.Now())
	}

	if m.App.TrashMode || m.Editor.NoteID() == "" {
		return nil
	}

	m.Focus = FocusEditor
	cmd := m.Editor.Focus()
	m.Editor.ClearSelection()

	// textarea 領域はヘッダー分だけ Y を補正
	m.Editor.StartDragSelection(edX, msg.Y-editorHeaderHeight)

	return cmd
}

func (m *Model) handleFolderListClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	// moreメニューが開いている場合
	if m.FolderList.MenuOpen() {
		return m.handleFolderMenuClick(msg)
	}

	// ヘッダーのボタンクリック判定
	if msg.Y < folderListHeaderLines {
		return m.handleFolderHeaderClick(msg, now)
	}

	idx := m.FolderList.HitTest(msg.X, msg.Y)
	if idx >= 0 {
		cmd := m.FolderList.SelectIndex(idx)
		m.Focus = FocusFolderList

		return m.processFolderListCmd(cmd, now)
	}

	m.Focus = FocusFolderList

	return nil
}

func (m *Model) handleFolderMenuClick(msg tea.MouseClickMsg) tea.Cmd {
	menuTopY := folderListHeaderLines
	menuHeight := m.FolderList.MenuHeight()
	menuWidth := m.FolderList.PopupMenu.Width()
	menuX := m.FolderList.Width() - folderListBorderWidth - menuWidth

	if msg.Y >= menuTopY && msg.Y < menuTopY+menuHeight && msg.X >= menuX && msg.X < menuX+menuWidth {
		relX := msg.X - menuX
		relY := msg.Y - menuTopY

		idx, hit := m.FolderList.PopupMenu.HandleClick(relX, relY)
		m.FolderList.CloseMenu()

		if hit {
			return m.handleFolderMenuAction(idx)
		}

		return nil
	}

	m.FolderList.CloseMenu()

	return nil
}

func (m *Model) handleFolderHeaderClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	hit := m.FolderList.HitTestHeader(msg.X, msg.Y)

	switch hit {
	case headerHitClose:
		return m.toggleFolderList(now)
	case headerHitAdd:
		m.FolderList.StartInput()
		m.Focus = FocusFolderList

		return nil
	case headerHitMore:
		if m.FolderList.IsUserFolder() {
			if m.FolderList.MenuOpen() {
				m.FolderList.CloseMenu()
			} else {
				m.FolderList.OpenMenu()
			}
		}

		return nil
	}

	return nil
}

func (m *Model) handleWheel(msg tea.MouseWheelMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()

	const scrollLines = 1

	noteListStart := m.noteListOffset()
	noteListEnd := noteListStart + m.noteListWidth

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

	m.hoverSeparator = m.resizing || m.isOnSeparator(mouse.X)
	m.hoverFolderSep = m.resizingFolder || (m.FolderList.Visible() && m.isOnFolderSeparator(mouse.X))

	if m.resizingFolder {
		maxFolderWidth := m.width - m.noteListWidth - minEditorWidth
		newWidth := max(mouse.X, minFolderListWidth)
		newWidth = min(newWidth, maxFolderWidth)
		m.folderListWidth = newWidth
		m.recalcLayout(now)

		return nil
	}

	if m.resizing {
		newWidth := max(mouse.X-m.noteListOffset(), minNoteListWidth)
		newWidth = min(newWidth, m.maxNoteListWidth())
		m.noteListWidth = newWidth
		m.recalcLayout(now)

		return nil
	}

	if m.Focus == FocusEditor && m.Editor.Selecting() {
		edX := mouse.X - m.noteListOffset() - m.noteListWidth
		m.Editor.UpdateDragSelection(edX, mouse.Y-editorHeaderHeight)

		return nil
	}

	m.updateConfirmDialogHover(mouse)
	m.updateNoteListFolderBtnHover(mouse)
	m.updateFolderListHeaderHover(mouse)
	m.updateEditorHeaderHover(mouse)
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

	offset := m.noteListOffset()
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
	m.hoverSeparator = m.isOnSeparator(mouse.X)
	m.hoverFolderSep = m.FolderList.Visible() && m.isOnFolderSeparator(mouse.X)
	m.updateConfirmDialogHover(mouse)
	m.updateEditorHeaderHover(mouse)
	m.updateFooterHover(mouse)

	return nil
}

func (m *Model) updateFolderListHeaderHover(mouse tea.Mouse) {
	if !m.FolderList.Visible() {
		return
	}

	if mouse.X < m.folderListWidth && mouse.Y == 0 {
		m.FolderList.SetHeaderHover(mouse.X, mouse.Y)
	} else {
		m.FolderList.ClearHeaderHover()
	}

	if m.FolderList.MenuOpen() {
		menuTopY := folderListHeaderLines
		menuHeight := m.FolderList.MenuHeight()
		menuWidth := m.FolderList.PopupMenu.Width()
		menuX := m.FolderList.Width() - folderListBorderWidth - menuWidth

		if mouse.Y >= menuTopY && mouse.Y < menuTopY+menuHeight && mouse.X >= menuX && mouse.X < menuX+menuWidth {
			m.FolderList.PopupMenu.SetHoverByPos(mouse.X-menuX, mouse.Y-menuTopY)
		} else {
			m.FolderList.PopupMenu.SetHoverByPos(-1, -1)
		}
	}
}

func (m *Model) updateEditorHeaderHover(mouse tea.Mouse) {
	editorStartX := m.noteListOffset() + m.noteListWidth
	if mouse.X >= editorStartX {
		m.Editor.HandleHover(mouse.X-editorStartX, mouse.Y)
	} else {
		m.Editor.Header.ClearHover()
	}

	if m.Editor.Header.MoveMenuOpen() {
		edX := mouse.X - editorStartX
		menuTopY := editorHeaderMenuTopY
		menuHeight := m.Editor.Header.MoveMenuHeight()
		menuWidth := m.Editor.Header.MoveMenu.Width()
		menuX := m.Editor.Header.Width() - moreButtonOffset + 1 - menuWidth

		if mouse.Y >= menuTopY && mouse.Y < menuTopY+menuHeight && edX >= menuX && edX < menuX+menuWidth {
			m.Editor.Header.SetMoveMenuHover(edX-menuX, mouse.Y-menuTopY)
		} else {
			m.Editor.Header.SetMoveMenuHover(-1, -1)
		}
	}
}

func (m *Model) updateFooterHover(mouse tea.Mouse) {
	footerLabelY := m.height - footerLineCount + 1

	if m.Footer.MenuOpen() {
		menuHeight := m.Footer.MenuHeight()
		bodyLines := m.height - footerLineCount
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

	msg, ok := cmd().(FolderListMsg)
	if !ok {
		return cmd
	}

	return m.processFolderListMsg(msg, now)
}

func (m *Model) processFolderListMsg(msg FolderListMsg, now time.Time) tea.Cmd {
	switch msg {
	case FolderListSelect:
		return m.handleFolderSelect(now)
	case FolderListFocusNext:
		m.Focus = FocusNoteList

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

	return m.processNoteListMsg(msg, now)
}

func (m *Model) processNoteListMsg(msg NoteListMsg, now time.Time) tea.Cmd {
	switch msg {
	case NoteListSelect:
		m.loadSelectedNote()

		return nil
	case NoteListCreate:
		return m.createNote(now)
	case NoteListTrash:
		return m.trashNote(now)
	case NoteListRestore:
		return m.restoreNote(now)
	case NoteListUndo:
		return m.undoNote(now)
	case NoteListRedo:
		return m.redoNote(now)
	case NoteListEdit:
		return m.focusEditor()
	case NoteListCopy:
		return m.copyNote()
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
	case EditorHeaderRestore:
		return m.restoreNote(now)
	case EditorHeaderPin:
		return m.pinNote()
	case EditorHeaderUnpin:
		return m.unpinNote()
	case EditorHeaderMove:
		return m.openMoveMenu()
	}

	return nil
}

func (m *Model) processFooterMsg(msg FooterMsg, now time.Time) tea.Cmd {
	switch msg {
	case FooterQuit:
		m.syncEditorToNote(now)

		return tea.Quit
	case FooterCopy:
		return m.copySelection()
	case FooterCut:
		return m.cutSelection()
	case FooterMore:
		return nil
	}

	return nil
}

// --- アクション ---

func (m *Model) focusEditor() tea.Cmd {
	if m.App.TrashMode {
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

func (m *Model) createNote(now time.Time) tea.Cmd {
	m.syncEditorToNote(now)

	folder := ""
	if m.FolderList.Visible() && m.FolderList.SelectedKind() == FolderUser {
		folder = m.FolderList.SelectedName()
	}

	result, err := m.App.CreateNote(now, folder)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	infoCmd := m.applyNoteResult(result, now)
	m.Editor.LoadNote(result.Note)
	m.Focus = FocusEditor

	return tea.Batch(m.Editor.Focus(), infoCmd)
}

func (m *Model) trashNote(now time.Time) tea.Cmd {
	if len(m.App.Notes) == 0 {
		return nil
	}

	_, ok := m.NoteList.SelectedNote()
	if !ok {
		return nil
	}

	m.syncEditorToNote(now)

	idx := m.NoteList.SelectedIndex()

	result, err := m.App.TrashNote(idx)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	return m.applyNoteResult(result, now)
}

func (m *Model) enterTrashMode(now time.Time) tea.Cmd {
	m.syncEditorToNote(now)

	err := m.App.EnterTrashMode()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.Editor.Blur()
	m.Editor.SetReadOnly(true)
	m.Editor.Header.SetTrashMode(true)

	if m.Focus != FocusFolderList {
		m.Focus = FocusNoteList
	}

	m.NoteList.Reset("Trash", false, m.App.TrashNotes, now)

	// フォルダ選択を同期
	m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderTrash))

	if len(m.App.TrashNotes) > 0 {
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}

	return nil
}

func (m *Model) exitTrashMode(now time.Time) tea.Cmd {
	m.App.ExitTrashMode()
	m.Editor.SetReadOnly(false)
	m.Editor.Header.SetTrashMode(false)

	// フォルダ選択を同期
	m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderNotes))

	notes := m.currentFolderNotes()
	m.NoteList.Reset(app.DefaultFolder, true, notes, now)

	if len(notes) > 0 {
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}

	return nil
}

func (m *Model) restoreNote(now time.Time) tea.Cmd {
	if !m.App.TrashMode {
		return nil
	}

	if len(m.App.TrashNotes) == 0 {
		return nil
	}

	_, ok := m.NoteList.SelectedNote()
	if !ok {
		return nil
	}

	idx := m.NoteList.SelectedIndex()

	result, err := m.App.RestoreNote(idx)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.Editor.SetReadOnly(false)
	m.Editor.Header.SetTrashMode(false)
	m.NoteList.SetTitle(app.DefaultFolder)
	m.NoteList.SetSectioned(true)
	m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderNotes))

	return m.applyNoteResult(result, now)
}

func (m *Model) undoNote(now time.Time) tea.Cmd {
	if m.App.TrashMode {
		m.exitTrashMode(now)
	}

	result, err := m.App.UndoNote()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	if result.SelectIdx < 0 && result.InfoHint == "" {
		return nil
	}

	return m.applyNoteResult(result, now)
}

func (m *Model) redoNote(now time.Time) tea.Cmd {
	if m.App.TrashMode {
		m.exitTrashMode(now)
	}

	result, err := m.App.RedoNote()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	if result.SelectIdx < 0 && result.InfoHint == "" {
		return nil
	}

	return m.applyNoteResult(result, now)
}

func (m *Model) copySelection() tea.Cmd {
	if m.Editor.HasSelection() {
		_ = m.Editor.CopySelection()
	}

	return nil
}

func (m *Model) cutSelection() tea.Cmd {
	if m.Editor.HasSelection() {
		_ = m.Editor.CutSelection()
	}

	return nil
}

func (m *Model) copyNote() tea.Cmd {
	content := m.Editor.Value()
	if content == "" {
		return nil
	}

	err := clipboard.WriteAll(content)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	return m.setInfoMsg("Copied")
}

func (m *Model) pinNote() tea.Cmd {
	id := m.Editor.NoteID()
	if id == "" {
		return nil
	}

	err := m.App.PinNote(id)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.Editor.Header.SetPinned(true)
	m.refreshNoteListKeepSelection(time.Now())

	return m.setInfoMsg("Pinned")
}

func (m *Model) unpinNote() tea.Cmd {
	id := m.Editor.NoteID()
	if id == "" {
		return nil
	}

	err := m.App.UnpinNote(id)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.Editor.Header.SetPinned(false)
	m.refreshNoteListKeepSelection(time.Now())

	return m.setInfoMsg("Unpinned")
}

func (m *Model) openMoveMenu() tea.Cmd {
	id := m.Editor.NoteID()
	if id == "" {
		return nil
	}

	// 現在のフォルダを特定
	currentFolder := ""

	for _, n := range m.App.Notes {
		if n.ID == id {
			parts := strings.SplitN(n.Path, string(filepath.Separator), 2) //nolint:mnd // folder/rest
			if len(parts) > 0 {
				currentFolder = parts[0]
			}

			break
		}
	}

	// 移動先候補: Notes + ユーザーフォルダから現在のフォルダを除外
	folders, err := m.App.ListFolders()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	candidates := make([]string, 0, len(folders)+1)
	if currentFolder != app.DefaultFolder {
		candidates = append(candidates, app.DefaultFolder)
	}

	for _, f := range folders {
		if f != currentFolder {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return m.setInfoMsg("No folders to move to")
	}

	m.Editor.Header.OpenMoveMenu(candidates)

	return nil
}

func (m *Model) handleNoteMove(msg noteMoveMsg, now time.Time) tea.Cmd {
	id := m.Editor.NoteID()
	if id == "" {
		return nil
	}

	m.syncEditorToNote(now)

	err := m.App.MoveNoteToFolder(id, msg.DestFolder)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	// 移動先フォルダに切り替え
	if m.FolderList.Visible() {
		m.refreshFolderList()
		m.FolderList.SelectIndex(m.FolderList.IndexByName(msg.DestFolder))

		notes := m.App.ListByFolder(msg.DestFolder)
		sectioned := msg.DestFolder == app.DefaultFolder
		m.NoteList.Reset(msg.DestFolder, sectioned, notes, now)

		// 移動したノートを選択
		for i, n := range notes {
			if n.ID == id {
				m.NoteList.SelectIndex(i, now)

				break
			}
		}

		m.loadSelectedNote()
	} else {
		m.refreshNoteListKeepSelection(now)
	}

	return m.setInfoMsg("Moved to " + msg.DestFolder)
}

// --- ヘルパー ---

func (m *Model) isOnSeparator(x int) bool {
	sepX := m.noteListOffset() + m.noteListWidth

	return x >= sepX-1 && x <= sepX
}

func (m *Model) isOnFolderSeparator(x int) bool {
	return x >= m.folderListWidth-1 && x <= m.folderListWidth
}

func (m *Model) noteListOffset() int {
	if m.FolderList.Visible() {
		return m.folderListWidth
	}

	return 0
}

func (m *Model) toggleFolderList(now time.Time) tea.Cmd {
	m.FolderList.ToggleVisible()

	if m.FolderList.Visible() {
		m.Focus = FocusFolderList
		m.refreshFolderList()

		// 現在のTrashMode状態をフォルダ選択に反映
		if m.App.TrashMode {
			m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderTrash))
		} else {
			m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderNotes))
		}
	} else if m.Focus == FocusFolderList {
		m.Focus = FocusNoteList
	}

	m.recalcLayout(now)

	return nil
}

func (m *Model) handleFolderSelect(now time.Time) tea.Cmd {
	switch m.FolderList.SelectedKind() {
	case FolderNotes:
		if m.App.TrashMode {
			return m.exitTrashMode(now)
		}

		notes := m.App.ListByFolder(app.DefaultFolder)
		m.NoteList.Reset(app.DefaultFolder, true, notes, now)

		if len(notes) > 0 {
			m.loadSelectedNote()
		} else {
			m.Editor.Clear()
		}
	case FolderTrash:
		if !m.App.TrashMode {
			return m.enterTrashMode(now)
		}
	case FolderUser:
		if m.App.TrashMode {
			m.App.ExitTrashMode()
			m.Editor.SetReadOnly(false)
			m.Editor.Header.SetTrashMode(false)
		}

		name := m.FolderList.SelectedName()
		notes := m.App.ListByFolder(name)
		m.NoteList.Reset(name, false, notes, now)

		if len(notes) > 0 {
			m.loadSelectedNote()
		} else {
			m.Editor.Clear()
		}
	}

	return nil
}

func (m *Model) handleFolderMenuAction(idx int) tea.Cmd {
	const (
		menuRename = 0
		menuDelete = 1
	)

	switch idx {
	case menuRename:
		m.FolderList.StartRename()

		return nil
	case menuDelete:
		name := m.FolderList.SelectedName()

		return func() tea.Msg {
			return folderDeleteMsg{Name: name}
		}
	}

	return nil
}

func (m *Model) handleFolderRename(msg folderRenameMsg) tea.Cmd {
	err := m.App.RenameFolder(msg.OldName, msg.NewName)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.refreshFolderList()

	return m.setInfoMsg("Renamed: " + msg.OldName + " → " + msg.NewName)
}

func (m *Model) handleFolderCreate(msg folderCreateMsg) tea.Cmd {
	err := m.App.CreateFolder(msg.Name)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.refreshFolderList()

	return m.setInfoMsg("Created: " + msg.Name)
}

func (m *Model) handleFolderDelete(msg folderDeleteMsg, _ time.Time) tea.Cmd {
	count, err := m.App.FolderNoteCount(msg.Name)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	if count > 0 {
		m.confirmDeleteFolder = msg.Name
		detail := fmt.Sprintf("%d note(s) will be moved to Trash.", count)
		dialog := NewConfirmDialog(fmt.Sprintf("Delete %q?", msg.Name), detail)
		m.confirmDialog = &dialog

		return nil
	}

	// 空フォルダは即時削除
	_, err = m.App.DeleteFolder(msg.Name)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.refreshFolderList()

	return m.setInfoMsg("Deleted: " + msg.Name)
}

func (m *Model) handleConfirmDialogKey(msg tea.KeyPressMsg) tea.Cmd {
	return m.applyConfirmResult(m.confirmDialog.Update(msg))
}

func (m *Model) handleConfirmDialogClick(msg tea.MouseClickMsg) tea.Cmd {
	originX, originY := m.confirmDialogOrigin()
	relX := msg.X - originX
	relY := msg.Y - originY

	return m.applyConfirmResult(m.confirmDialog.HandleClick(relX, relY))
}

func (m *Model) applyConfirmResult(result ConfirmResult) tea.Cmd {
	switch result {
	case ConfirmYes:
		name := m.confirmDeleteFolder
		m.confirmDialog = nil
		m.confirmDeleteFolder = ""

		deleted, err := m.App.DeleteFolder(name)
		if err != nil {
			m.errMsg = err.Error()

			return nil
		}

		m.refreshFolderList()

		return m.setInfoMsg("Deleted: " + name + " (" + strconv.Itoa(deleted) + " note(s) trashed)")
	case ConfirmNo:
		m.confirmDialog = nil
		m.confirmDeleteFolder = ""

		return nil
	case ConfirmContinue:
		return nil
	}

	return nil
}

func (m *Model) refreshFolderList() {
	folders, err := m.App.ListFolders()
	if err != nil {
		m.errMsg = err.Error()

		return
	}

	notesCount := len(m.App.ListByFolder(app.DefaultFolder))

	folderCounts := make(map[string]int, len(folders))
	for _, name := range folders {
		count, err := m.App.FolderNoteCount(name)
		if err != nil {
			continue
		}

		folderCounts[name] = count
	}

	m.FolderList.SetFolders(folders, notesCount, len(m.App.TrashNotes), folderCounts)
	_ = m.FolderList.SelectIndex(0)
}

func (m *Model) recalcLayout(now time.Time) {
	bodyHeight := m.height - footerLineCount

	if m.FolderList.Visible() {
		folderW := max(m.folderListWidth, 0)
		noteW := max(m.noteListWidth, 0)
		editorW := max(m.width-noteW-folderW, minEditorWidth)

		m.FolderList.SetSize(folderW, bodyHeight)
		m.NoteList.SetSize(noteW, bodyHeight, now)
		m.Editor.SetSize(editorW, bodyHeight)
	} else {
		noteW := max(m.noteListWidth, 0)
		editorW := max(m.width-noteW, minEditorWidth)

		m.NoteList.SetSize(noteW, bodyHeight, now)
		m.Editor.SetSize(editorW, bodyHeight)
	}
}

func (m *Model) loadSelectedNote() {
	n, ok := m.NoteList.SelectedNote()
	if !ok {
		m.Editor.Clear()

		return
	}

	n, err := m.App.LoadNote(n)
	if err != nil {
		m.errMsg = err.Error()
	}

	m.Editor.LoadNote(n)
}

// refreshNoteListKeepSelection はNoteListを現在のフォルダに応じたノート一覧で更新し、選択を維持する。
func (m *Model) refreshNoteListKeepSelection(now time.Time) {
	notes := m.currentFolderNotes()

	// 現在選択中のノートIDを記憶
	selectedID := m.Editor.NoteID()
	selectIdx := 0

	for i, n := range notes {
		if n.ID == selectedID {
			selectIdx = i

			break
		}
	}

	m.NoteList.SetNotes(notes, now)
	m.NoteList.SelectIndex(selectIdx, now)
}

// currentFolderNotes は現在のフォルダビューに応じたノート一覧を返す。
func (m *Model) currentFolderNotes() []note.Note {
	if !m.FolderList.Visible() {
		return m.App.Notes
	}

	switch m.FolderList.SelectedKind() {
	case FolderNotes:
		return m.App.ListByFolder(app.DefaultFolder)
	case FolderUser:
		return m.App.ListByFolder(m.FolderList.SelectedName())
	case FolderTrash:
		return m.App.TrashNotes
	}

	return m.App.Notes
}

func (m *Model) syncEditorToNote(now time.Time) {
	if !m.Editor.Dirty() {
		return
	}

	_, err := m.App.SaveNote(m.Editor.NoteID(), m.Editor.Value(), now)
	if err != nil {
		m.errMsg = err.Error()
	}

	m.Editor.MarkClean()
	m.refreshNoteListKeepSelection(now)
}

// applyNoteResult は NoteResult をUI状態に反映する。
func (m *Model) applyNoteResult(r app.NoteResult, now time.Time) tea.Cmd {
	notes := r.Notes
	selectIdx := r.SelectIdx

	if m.FolderList.Visible() && m.FolderList.SelectedKind() != FolderTrash {
		notes = m.currentFolderNotes()

		// フィルタ後のリストでの選択インデックスを再計算
		selectIdx = -1

		if r.Note.ID != "" {
			for i, n := range notes {
				if n.ID == r.Note.ID {
					selectIdx = i

					break
				}
			}
		}
	}

	m.NoteList.SetNotes(notes, now)

	if selectIdx >= 0 {
		m.NoteList.SelectIndex(selectIdx, now)
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}

	if r.InfoHint != "" {
		return m.setInfoMsg(r.InfoHint)
	}

	return nil
}

func (m *Model) setInfoMsg(msg string) tea.Cmd {
	m.infoMsgID++
	m.infoMsg = msg

	id := m.infoMsgID

	return tea.Tick(infoMsgDuration, func(_ time.Time) tea.Msg {
		return clearInfoMsg{id: id}
	})
}
