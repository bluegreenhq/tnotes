package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"github.com/bluegreenhq/tnotes/internal/app"
)

// --- イベントハンドラ ---

// Update はメッセージに応じて状態を更新する。
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:cyclop // type switch dispatch
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
		return m, m.handleFolderListMsg(msg, now)
	case NoteListMsg:
		return m, m.handleNoteListMsg(msg, now)
	case EditorMsg:
		return m, m.handleEditorMsg(msg, now)
	case FooterMsg:
		return m, m.handleFooterMsg(msg, now)
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
	m.errMsg = ""
	m.infoMsg = ""
	m.Footer.CloseMenu()

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
	m.NoteList.SetNotes(m.App.Notes, time.Now())

	if len(m.App.Notes) > 0 {
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}

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

func (m *Model) handleClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd { //nolint:cyclop // mouse dispatch
	m.errMsg = ""

	if msg.Button != tea.MouseLeft {
		return nil
	}

	footerLabelY := m.height - footerLineCount + 1 // フッター3行の中央（ラベル行）

	// メニューが開いている場合
	if m.Footer.MenuOpen() {
		return m.handleClickWithMenu(msg, now)
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

		return m.handleFooterMsg(fMsg, now)
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

	return m.handleFooterMsg(fMsg, now)
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
	if m.App.TrashMode || m.Editor.NoteID() == "" {
		return nil
	}

	m.Focus = FocusEditor
	cmd := m.Editor.Focus()
	m.Editor.ClearSelection()

	edX := msg.X - m.noteListOffset() - m.noteListWidth
	m.Editor.StartDragSelection(edX, msg.Y)

	return cmd
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
		newWidth := max(mouse.X, minFolderListWidth)
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
		m.Editor.UpdateDragSelection(edX, mouse.Y)

		return nil
	}

	m.updateFooterHover(mouse)

	return nil
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
	m.updateFooterHover(mouse)

	return nil
}

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
		m.FolderList.UpdateCounts(len(m.App.Notes), len(m.App.TrashNotes))

		// 現在のTrashMode状態をフォルダ選択に反映
		if m.App.TrashMode {
			m.FolderList.SelectIndex(1) // Trash
		} else {
			m.FolderList.SelectIndex(0) // Notes
		}
	} else if m.Focus == FocusFolderList {
		m.Focus = FocusNoteList
	}

	m.recalcLayout(now)

	return nil
}

func (m *Model) recalcLayout(now time.Time) {
	bodyHeight := m.height - footerLineCount

	if m.FolderList.Visible() {
		m.FolderList.SetSize(m.folderListWidth, bodyHeight)
		m.NoteList.SetSize(m.noteListWidth, bodyHeight, now)
		m.Editor.SetSize(m.width-m.noteListWidth-m.folderListWidth, bodyHeight)
	} else {
		m.NoteList.SetSize(m.noteListWidth, bodyHeight, now)
		m.Editor.SetSize(m.width-m.noteListWidth, bodyHeight)
	}
}

func (m *Model) processFolderListCmd(cmd tea.Cmd, now time.Time) tea.Cmd {
	if cmd == nil {
		return nil
	}

	msg := cmd()

	flMsg, ok := msg.(FolderListMsg)
	if !ok {
		return cmd
	}

	return m.handleFolderListMsg(flMsg, now)
}

func (m *Model) handleFolderListMsg(msg FolderListMsg, now time.Time) tea.Cmd {
	switch msg {
	case FolderListSelect:
		return m.handleFolderSelect(now)
	case FolderListFocusNext:
		m.Focus = FocusNoteList

		return nil
	}

	return nil
}

func (m *Model) handleFolderSelect(now time.Time) tea.Cmd {
	switch m.FolderList.SelectedKind() {
	case FolderNotes:
		if m.App.TrashMode {
			return m.exitTrashMode(now)
		}
	case FolderTrash:
		if !m.App.TrashMode {
			return m.enterTrashMode(now)
		}
	}

	return nil
}

func (m *Model) handleFolderListClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	// ヘッダーの閉じるボタン（✕）クリック判定
	if msg.Y < folderListHeaderLines {
		if msg.X >= 1 && msg.X <= 2 {
			return m.toggleFolderList(now)
		}

		return nil
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

// --- メッセージディスパッチ ---

// processNoteListCmd は NoteList.Update が返した tea.Cmd を即座に処理する。
// NoteListMsg の場合はアクションを実行し、それ以外はそのまま返す。
func (m *Model) processNoteListCmd(cmd tea.Cmd, now time.Time) tea.Cmd {
	if cmd == nil {
		return nil
	}

	msg := cmd()

	nMsg, ok := msg.(NoteListMsg)
	if !ok {
		return cmd
	}

	return m.handleNoteListMsg(nMsg, now)
}

func (m *Model) handleNoteListMsg(msg NoteListMsg, now time.Time) tea.Cmd {
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

	msg := cmd()

	eMsg, ok := msg.(EditorMsg)
	if !ok {
		return cmd
	}

	return m.handleEditorMsg(eMsg, now)
}

func (m *Model) handleEditorMsg(msg EditorMsg, now time.Time) tea.Cmd {
	switch msg {
	case EditorBlur:
		return m.blurEditor(now)
	case EditorSave:
		m.syncEditorToNote(now)

		return m.setInfoMsg("Saved")
	}

	return nil
}

func (m *Model) handleFooterMsg(msg FooterMsg, now time.Time) tea.Cmd {
	switch msg {
	case FooterNew:
		return m.createNote(now)
	case FooterRestore:
		return m.restoreNote(now)
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

	result, err := m.App.CreateNote(now)
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

	if m.Focus != FocusFolderList {
		m.Focus = FocusNoteList
	}

	m.NoteList.Reset("Trash", false, m.App.TrashNotes, now)

	// フォルダ選択を同期
	m.FolderList.SelectIndex(1) // Trash

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
	m.NoteList.Reset("Notes", true, m.App.Notes, now)

	// フォルダ選択を同期
	m.FolderList.SelectIndex(0) // Notes

	if len(m.App.Notes) > 0 {
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
	m.NoteList.SetTitle("Notes")
	m.NoteList.SetSectioned(true)

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

// --- ヘルパー ---

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

func (m *Model) syncEditorToNote(now time.Time) {
	if !m.Editor.Dirty() {
		return
	}

	newIdx, err := m.App.SaveNote(m.Editor.NoteID(), m.Editor.Value(), now)
	if err != nil {
		m.errMsg = err.Error()
	}

	m.Editor.MarkClean()
	m.NoteList.SetNotes(m.App.Notes, now)
	m.NoteList.SelectIndex(newIdx, now)
}

// applyNoteResult は NoteResult をUI状態に反映する。
func (m *Model) applyNoteResult(r app.NoteResult, now time.Time) tea.Cmd {
	m.NoteList.SetNotes(r.Notes, now)

	if r.SelectIdx >= 0 {
		m.NoteList.SelectIndex(r.SelectIdx, now)
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
