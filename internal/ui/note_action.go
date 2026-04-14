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

// execNoteAction は sync → App呼び出し → エラー処理 → applyNoteResult の共通パターンを実行する。
func (m *Model) execNoteAction(now time.Time, sync bool, fn func() (app.NoteResult, error)) tea.Cmd {
	if sync {
		m.syncEditorToNote(now)
	}

	result, err := fn()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	return m.applyNoteResult(result, now)
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
	if len(m.currentFolderNotes()) == 0 {
		return nil
	}

	selected, ok := m.NoteList.SelectedNote()
	if !ok {
		return nil
	}

	return m.execNoteAction(now, true, func() (app.NoteResult, error) {
		return m.App.TrashNote(selected.ID)
	})
}

func (m *Model) duplicateNote(now time.Time) tea.Cmd {
	if len(m.currentFolderNotes()) == 0 {
		return nil
	}

	selected, ok := m.NoteList.SelectedNote()
	if !ok {
		return nil
	}

	return m.execNoteAction(now, true, func() (app.NoteResult, error) {
		return m.App.DuplicateNote(selected.ID)
	})
}

func (m *Model) enterTrashMode(now time.Time) tea.Cmd {
	m.syncEditorToNote(now)

	err := m.App.RefreshTrashNotes()
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

	m.NoteList.Reset("Trash", false, m.App.ListTrashNotes(), now)

	// フォルダ選択を同期
	m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderTrash))

	if len(m.App.ListTrashNotes()) > 0 {
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}

	return nil
}

func (m *Model) exitTrashMode(now time.Time) tea.Cmd { //nolint:unparam // 他アクションメソッドとシグネチャを統一
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

func (m *Model) undoNote(now time.Time) tea.Cmd {
	if m.Editor.Header.TrashMode() {
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
	if m.Editor.Header.TrashMode() {
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

	currentFolder := m.findNoteFolder(id)

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

	// 右クリックメニュー経由の場合、アンカーを復元して同じ位置にサブメニューを表示
	if a := m.popup.TakeLastAnchor(); a != nil {
		m.popup.SetAnchor(a.x, a.y)
	}

	return nil
}

// findNoteFolder は指定IDのノートが属するフォルダ名を返す。
// 通常ノートとゴミ箱ノートの両方を検索する。
func (m *Model) findNoteFolder(id note.NoteID) string {
	for _, n := range m.App.Notes {
		if n.ID == id {
			parts := strings.SplitN(n.Path, string(filepath.Separator), 2) //nolint:mnd // folder/rest
			if len(parts) > 0 {
				return parts[0]
			}

			return ""
		}
	}

	return ""
}

func (m *Model) handleNoteMove(msg noteMoveMsg, now time.Time) tea.Cmd {
	id := m.Editor.NoteID()
	if id == "" {
		return nil
	}

	wasTrash := m.Editor.Header.TrashMode()

	if !wasTrash {
		m.syncEditorToNote(now)
	}

	err := m.App.MoveNoteToFolder(id, msg.DestFolder)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	// Trash から移動した場合は ReadOnly を解除
	if wasTrash {
		m.Editor.SetReadOnly(false)
		m.Editor.Header.SetTrashMode(false)
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

func (m *Model) toggleFolderList(now time.Time) tea.Cmd {
	m.FolderList.ToggleVisible()

	if m.FolderList.Visible() {
		m.Focus = FocusFolderList
		m.refreshFolderList()

		// 現在のTrash表示状態をフォルダ選択に反映
		if m.Editor.Header.TrashMode() {
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
		m.Editor.SetReadOnly(false)
		m.Editor.Header.SetTrashMode(false)

		notes := m.App.ListByFolder(app.DefaultFolder)
		m.NoteList.Reset(app.DefaultFolder, true, notes, now)

		if len(notes) > 0 {
			m.loadSelectedNote()
		} else {
			m.Editor.Clear()
		}
	case FolderTrash:
		return m.enterTrashMode(now)
	case FolderUser:
		m.Editor.SetReadOnly(false)
		m.Editor.Header.SetTrashMode(false)

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

func (m *Model) handleFolderMenuAction(idx int, now time.Time) tea.Cmd {
	const (
		menuRename = 0
		menuDelete = 1
	)

	switch idx {
	case menuRename:
		return m.FolderList.StartRename()
	case menuDelete:
		name := m.FolderList.SelectedName()

		return m.handleFolderDelete(folderDeleteMsg{Name: name}, now)
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
		dialog.SetScreenSize(m.layout.width, m.layout.BodyHeight())
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
	return m.applyConfirmResult(m.confirmDialog.HandleClickAbs(msg.X, msg.Y))
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

	m.FolderList.SetFolders(folders, notesCount, len(m.App.ListTrashNotes()), folderCounts)
	_ = m.FolderList.SelectIndex(0)
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
	switch m.FolderList.SelectedKind() {
	case FolderNotes:
		return m.App.ListByFolder(app.DefaultFolder)
	case FolderUser:
		return m.App.ListByFolder(m.FolderList.SelectedName())
	case FolderTrash:
		return m.App.ListTrashNotes()
	}

	return m.App.ListByFolder(app.DefaultFolder)
}

func (m *Model) syncEditorToNote(now time.Time) {
	saved := false

	if m.Editor.Dirty() {
		_, err := m.App.SaveNote(m.Editor.NoteID(), m.Editor.Value(), now)
		if err != nil {
			m.errMsg = err.Error()
		}

		m.Editor.MarkClean()

		saved = true
	}

	if m.App.DiscardIfEmpty(m.Editor.NoteID()) {
		m.refreshNoteListKeepSelection(now)

		return
	}

	if saved {
		m.refreshNoteListKeepSelection(now)
	}
}

// applyNoteResult は NoteResult をUI状態に反映する。
func (m *Model) applyNoteResult(r app.NoteResult, now time.Time) tea.Cmd {
	notes := m.currentFolderNotes()
	selectIdx := m.resolveSelectIdx(r, notes)

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

// resolveSelectIdx は NoteResult からUI上の選択インデックスを決定する。
func (m *Model) resolveSelectIdx(r app.NoteResult, notes []note.Note) int {
	// Note.ID による検索
	if r.Note.ID != "" {
		for i, n := range notes {
			if n.ID == r.Note.ID {
				return i
			}
		}
	}

	// SelectIdx によるフォールバック
	if r.SelectIdx >= 0 && r.SelectIdx < len(notes) {
		return r.SelectIdx
	}

	// ノートが残っていれば現在の選択位置を維持
	if len(notes) > 0 {
		return min(m.NoteList.SelectedIndex(), len(notes)-1)
	}

	return -1
}

func (m *Model) setInfoMsg(msg string) tea.Cmd {
	m.infoMsgID++
	m.infoMsg = msg

	id := m.infoMsgID

	return tea.Tick(infoMsgDuration, func(_ time.Time) tea.Msg {
		return clearInfoMsg{id: id}
	})
}
