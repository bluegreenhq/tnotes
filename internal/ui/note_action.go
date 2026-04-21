package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

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
	folder := ""
	if m.FolderList.Visible() && m.FolderList.SelectedKind() == FolderUser {
		folder = m.FolderList.SelectedName()
	}

	return m.execNoteAction(now, true, func() (app.NoteResult, error) {
		return m.App.CreateNote(now, folder)
	})
}

func (m *Model) trashNote(now time.Time) tea.Cmd {
	if len(m.NoteList.CurrentFolderNotes(m.FolderList.SelectedKind(), m.FolderList.SelectedName())) == 0 {
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
	if len(m.NoteList.CurrentFolderNotes(m.FolderList.SelectedKind(), m.FolderList.SelectedName())) == 0 {
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

// folderView はフォルダ切替時のUI状態をまとめた構造体。
type folderView struct {
	name      string      // 表示名
	notes     []note.Note // 表示するノート一覧
	sectioned bool        // Today/Yesterday 等のセクション分けを行うか
	readOnly  bool        // エディタを読み取り専用にするか
	trash     bool        // ゴミ箱モードか
	selectID  note.NoteID // 指定IDのノートを選択（空なら先頭）
}

// switchFolder は表示フォルダを切り替え、Editor・NoteList の状態を同期する。
func (m *Model) switchFolder(fv folderView, now time.Time) {
	m.Editor.SetReadOnly(fv.readOnly)
	m.Editor.Header.SetTrashMode(fv.trash)

	if fv.trash {
		m.Editor.Blur()

		if m.Focus != FocusFolderList {
			m.Focus = FocusNoteList
		}
	}

	m.NoteList.Reset(fv.name, fv.sectioned, fv.notes, now)

	if fv.selectID != "" {
		for i, n := range fv.notes {
			if n.ID == fv.selectID {
				m.NoteList.SelectIndex(i, now)

				break
			}
		}
	}

	if len(fv.notes) > 0 {
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}
}

func (m *Model) enterTrashMode(now time.Time) tea.Cmd {
	m.syncEditorToNote(now)

	err := m.App.RefreshTrashNotes()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderTrash))
	m.switchFolder(folderView{
		name:      "Trash",
		notes:     m.App.ListTrashNotes(),
		sectioned: false,
		readOnly:  true,
		trash:     true,
		selectID:  "",
	}, now)

	return nil
}

func (m *Model) exitTrashMode(now time.Time) tea.Cmd { //nolint:unparam // 他アクションメソッドとシグネチャを統一
	m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderNotes))
	m.switchFolder(folderView{
		name:      app.DefaultFolder,
		notes:     m.NoteList.CurrentFolderNotes(m.FolderList.SelectedKind(), m.FolderList.SelectedName()),
		sectioned: true,
		readOnly:  false,
		trash:     false,
		selectID:  "",
	}, now)

	return nil
}

func (m *Model) undoRedoNote(now time.Time, undo bool) tea.Cmd {
	if m.Editor.Header.TrashMode() {
		m.exitTrashMode(now)
	}

	var (
		result app.NoteResult
		err    error
	)

	if undo {
		result, err = m.App.UndoNote()
	} else {
		result, err = m.App.RedoNote()
	}

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
	err := m.Editor.CopyToClipboard()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	return m.setInfoMsg("Copied")
}

func (m *Model) setNotePin(pin bool) tea.Cmd {
	id := m.Editor.NoteID()
	if id == "" {
		return nil
	}

	var err error
	if pin {
		err = m.App.PinNote(id)
	} else {
		err = m.App.UnpinNote(id)
	}

	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.Editor.Header.SetPinned(pin)

	kind, name := m.FolderList.SelectedKind(), m.FolderList.SelectedName()
	m.NoteList.RefreshKeepSelection(kind, name, m.Editor.NoteID(), time.Now())

	msg := "Pinned"
	if !pin {
		msg = "Unpinned"
	}

	return m.setInfoMsg(msg)
}

func (m *Model) openMoveMenu() tea.Cmd {
	opened, err := m.Editor.BuildMoveMenu()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	if !opened {
		return m.setInfoMsg("No folders to move to")
	}

	// 右クリックメニュー経由の場合、アンカーを復元して同じ位置にサブメニューを表示
	if a := m.popup.TakeLastAnchor(); a != nil {
		m.popup.SetAnchor(a.x, a.y)
	}

	return nil
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

	// 移動先フォルダに切り替え
	if m.FolderList.Visible() {
		_ = m.FolderList.RefreshFromApp()
		m.FolderList.SelectIndex(m.FolderList.IndexByName(msg.DestFolder))
		m.switchFolder(folderView{
			name:      msg.DestFolder,
			notes:     m.App.ListByFolder(msg.DestFolder),
			sectioned: msg.DestFolder == app.DefaultFolder,
			readOnly:  false,
			trash:     false,
			selectID:  id,
		}, now)
	} else {
		m.NoteList.RefreshKeepSelection(m.FolderList.SelectedKind(), m.FolderList.SelectedName(), m.Editor.NoteID(), now)
	}

	return m.setInfoMsg("Moved to " + msg.DestFolder)
}

func (m *Model) toggleFolderList(now time.Time) tea.Cmd {
	m.FolderList.ToggleVisible()

	if m.FolderList.Visible() {
		m.Focus = FocusFolderList
		_ = m.FolderList.RefreshFromApp()

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
		m.switchFolder(folderView{
			name:      app.DefaultFolder,
			notes:     m.App.ListByFolder(app.DefaultFolder),
			sectioned: true,
			readOnly:  false,
			trash:     false,
			selectID:  "",
		}, now)
	case FolderTrash:
		return m.enterTrashMode(now)
	case FolderUser:
		name := m.FolderList.SelectedName()
		m.switchFolder(folderView{
			name:      name,
			notes:     m.App.ListByFolder(name),
			sectioned: false,
			readOnly:  false,
			trash:     false,
			selectID:  "",
		}, now)
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

		return m.processFolderListCmd(m.FolderList.TryDeleteFolder(name), now)
	}

	return nil
}

func (m *Model) handleFolderResult(msg folderResultMsg) tea.Cmd {
	if msg.Err != nil {
		m.errMsg = msg.Err.Error()

		return nil
	}

	return m.setInfoMsg(msg.Info)
}

func (m *Model) loadSelectedNote() {
	err := m.Editor.LoadSelected(&m.NoteList)
	if err != nil {
		m.errMsg = err.Error()
	}
}

func (m *Model) syncEditorToNote(now time.Time) {
	saved, err := m.Editor.Save(now)
	if err != nil {
		m.errMsg = err.Error()
	}

	if m.App.DiscardIfEmpty(m.Editor.NoteID()) || saved {
		m.NoteList.RefreshKeepSelection(m.FolderList.SelectedKind(), m.FolderList.SelectedName(), m.Editor.NoteID(), now)
		m.updateIndexModTime()
	}
}

// applyNoteResult は NoteResult をUI状態に反映する。
func (m *Model) applyNoteResult(r app.NoteResult, now time.Time) tea.Cmd {
	notes := m.NoteList.CurrentFolderNotes(m.FolderList.SelectedKind(), m.FolderList.SelectedName())
	selectIdx := m.NoteList.ResolveSelectIdx(r, notes)

	m.NoteList.SetNotes(notes, now)

	if selectIdx >= 0 {
		m.NoteList.SelectIndex(selectIdx, now)
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}

	if r.LoadNote {
		m.Editor.LoadNote(r.Note)
	}

	var cmds []tea.Cmd

	if r.FocusEditor {
		m.Focus = FocusEditor
		cmds = append(cmds, m.Editor.Focus())
	}

	if r.InfoHint != "" {
		cmds = append(cmds, m.setInfoMsg(r.InfoHint))
	}

	return tea.Batch(cmds...)
}

func (m *Model) setInfoMsg(msg string) tea.Cmd {
	m.infoMsgID++
	m.infoMsg = msg

	id := m.infoMsgID

	return tea.Tick(infoMsgDuration, func(_ time.Time) tea.Msg {
		return clearInfoMsg{id: id}
	})
}
