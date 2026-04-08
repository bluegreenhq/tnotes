package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/app"
)

// Update はメッセージに応じて状態を更新する。
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:cyclop,funlen // bubbletea Update
	now := time.Now()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		editorWidth := msg.Width - sidebarWidthPx
		m.Sidebar.SetSize(sidebarWidthPx, msg.Height-1, now)
		m.Editor.SetSize(editorWidth, msg.Height-1)

		return m, nil

	case tea.KeyPressMsg:
		return m, m.handleKey(msg, now)

	case tea.MouseClickMsg:
		return m, m.handleClick(msg, now)

	case tea.MouseMotionMsg:
		mouse := msg.Mouse()
		// ドラッグ中の選択範囲更新
		if m.Focus == FocusEditor && m.Editor.Selecting() {
			edX := mouse.X - sidebarWidthPx
			m.Editor.UpdateDragSelection(edX, mouse.Y)

			return m, nil
		}

		m.handleMouseMotion(msg)

		return m, nil

	case tea.MouseReleaseMsg:
		if m.Editor.Selecting() {
			m.Editor.StopDragSelection()
		}

		return m, nil

	case tea.MouseWheelMsg:
		m.handleWheel(msg, now)

		return m, nil

	case tea.MouseMsg:
		m.handleMouseMotion(msg)

		return m, nil

	case SidebarMsg:
		return m, m.handleSidebarMsg(msg, now)

	case FooterMsg:
		return m, m.handleFooterMsg(msg, now)

	case clearInfoMsg:
		if msg.id == m.infoMsgID {
			m.infoMsg = ""
		}

		return m, nil
	}

	if m.Focus == FocusEditor {
		_, cmd := m.Editor.Update(msg, now)

		return m, cmd
	}

	return m, nil
}

// --- キー入力 ---

func (m *Model) handleKey(msg tea.KeyPressMsg, now time.Time) tea.Cmd { //nolint:cyclop // キーバインド分岐
	m.errMsg = ""
	m.infoMsg = ""

	switch {
	case msg.Code == 'q' && m.Focus == FocusSidebar:
		m.syncEditorToNote(now)

		return tea.Quit
	case msg.Code == tea.KeyTab && m.Focus == FocusSidebar:
		return m.focusEditor()
	case msg.Code == tea.KeyTab && m.Focus == FocusEditor:
		return m.blurEditor(now)
	case msg.Code == tea.KeyEscape && m.Focus == FocusEditor:
		if m.Editor.HasSelection() {
			m.Editor.ClearSelection()

			return nil
		}

		return m.blurEditor(now)
	}

	switch m.Focus {
	case FocusSidebar:
		_, cmd := m.Sidebar.Update(msg, now, m.App.TrashMode)

		return m.processSidebarCmd(cmd, now)
	case FocusEditor:
		_, cmd := m.Editor.Update(msg, now)

		return cmd
	}

	return nil
}

// --- マウス入力 ---

func (m *Model) handleClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	m.errMsg = ""

	if msg.Button != tea.MouseLeft {
		return nil
	}

	footerY := m.height - 1

	switch {
	case msg.Y == footerY:
		return m.handleFooterClick(msg.X, now)
	case msg.X < sidebarWidthPx:
		return m.handleSidebarClick(msg, now)
	default:
		return m.handleEditorClick(msg)
	}
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

func (m *Model) handleSidebarClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	idx := m.Sidebar.HitTest(msg.X, msg.Y, now)
	if idx >= 0 {
		if !m.App.TrashMode {
			m.syncEditorToNote(now)
		}

		m.Sidebar.SelectIndex(idx, now)
		m.loadSelectedNote()
	}

	m.Focus = FocusSidebar
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

	edX := msg.X - sidebarWidthPx
	m.Editor.StartDragSelection(edX, msg.Y)

	return cmd
}

func (m *Model) handleWheel(msg tea.MouseWheelMsg, now time.Time) {
	mouse := msg.Mouse()

	const scrollLines = 1

	if mouse.X < sidebarWidthPx {
		switch mouse.Button {
		case tea.MouseWheelUp:
			m.Sidebar.ScrollUp(scrollLines, now)
		case tea.MouseWheelDown:
			m.Sidebar.ScrollDown(scrollLines, now)
		}

		return
	}

	switch mouse.Button {
	case tea.MouseWheelUp:
		m.Editor.ScrollUp(scrollLines)
	case tea.MouseWheelDown:
		m.Editor.ScrollDown(scrollLines)
	}
}

func (m *Model) handleMouseMotion(msg tea.MouseMsg) {
	mouse := msg.Mouse()

	footerY := m.height - 1
	if mouse.Y == footerY {
		m.rebuildFooterButtons()
		m.Footer.SetHover(m.Footer.HitTest(mouse.X))
	} else {
		m.Footer.SetHover(HoverNone)
	}
}

// --- メッセージディスパッチ ---

// processSidebarCmd は Sidebar.Update が返した tea.Cmd を即座に処理する。
// SidebarMsg の場合はアクションを実行し、それ以外はそのまま返す。
func (m *Model) processSidebarCmd(cmd tea.Cmd, now time.Time) tea.Cmd {
	if cmd == nil {
		return nil
	}

	msg := cmd()

	sMsg, ok := msg.(SidebarMsg)
	if !ok {
		return cmd
	}

	return m.handleSidebarMsg(sMsg, now)
}

func (m *Model) handleSidebarMsg(msg SidebarMsg, now time.Time) tea.Cmd { //nolint:cyclop // switch dispatch
	switch msg {
	case SidebarSelect:
		m.loadSelectedNote()

		return nil
	case SidebarCreate:
		return m.createNote(now)
	case SidebarTrash:
		return m.trashNote(now)
	case SidebarEnterTrash:
		return m.enterTrashMode(now)
	case SidebarExitTrash:
		return m.exitTrashMode(now)
	case SidebarRestore:
		return m.restoreNote(now)
	case SidebarUndo:
		return m.undoNote(now)
	case SidebarRedo:
		return m.redoNote(now)
	case SidebarEdit:
		return m.focusEditor()
	case SidebarQuit:
		m.syncEditorToNote(now)

		return tea.Quit
	}

	return nil
}

func (m *Model) handleFooterMsg(msg FooterMsg, now time.Time) tea.Cmd {
	switch msg {
	case FooterNew:
		return m.createNote(now)
	case FooterRestore:
		return m.restoreNote(now)
	case FooterTrashToggle:
		return m.toggleTrashMode(now)
	case FooterQuit:
		m.syncEditorToNote(now)

		return tea.Quit
	case FooterCopy:
		return m.copySelection()
	case FooterCut:
		return m.cutSelection()
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
	m.Focus = FocusSidebar
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

	_, ok := m.Sidebar.SelectedNote()
	if !ok {
		return nil
	}

	m.syncEditorToNote(now)

	idx := m.Sidebar.SelectedIndex()

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
	m.Focus = FocusSidebar
	m.Sidebar.Reset("Trash", false, m.App.TrashNotes, now)

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
	m.Sidebar.Reset("Notes", true, m.App.Notes, now)

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

	_, ok := m.Sidebar.SelectedNote()
	if !ok {
		return nil
	}

	idx := m.Sidebar.SelectedIndex()

	result, err := m.App.RestoreNote(idx)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.Editor.SetReadOnly(false)
	m.Sidebar.SetTitle("Notes")
	m.Sidebar.SetSectioned(true)

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

func (m *Model) toggleTrashMode(now time.Time) tea.Cmd {
	if m.App.TrashMode {
		return m.exitTrashMode(now)
	}

	return m.enterTrashMode(now)
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

// --- ヘルパー ---

func (m *Model) loadSelectedNote() {
	n, ok := m.Sidebar.SelectedNote()
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
	m.Sidebar.SetNotes(m.App.Notes, now)
	m.Sidebar.SelectIndex(newIdx, now)
}

// applyNoteResult は NoteResult をUI状態に反映する。
func (m *Model) applyNoteResult(r app.NoteResult, now time.Time) tea.Cmd {
	m.Sidebar.SetNotes(r.Notes, now)

	if r.SelectIdx >= 0 {
		m.Sidebar.SelectIndex(r.SelectIdx, now)
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
