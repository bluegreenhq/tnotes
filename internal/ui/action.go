package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// ModelAction はペインからモデルへの「やってほしいこと」を表す関数。
// pane の UpdatePane が直接返し、Model が即時実行する。
type ModelAction func(m *Model, ctx ActionContext) tea.Cmd

// ActionContext は ModelAction に渡す共通コンテキスト。
type ActionContext struct {
	Now time.Time
}

// --- 共通アクション（複数のペイン／オーバーレイから emit される） ---

// quit は編集中ノートを保存してから tea.Quit を返す。
func quit(m *Model, ctx ActionContext) tea.Cmd {
	m.syncEditorToNote(ctx.Now)

	return tea.Quit
}

// openHelp はヘルプオーバーレイを開く。
func openHelp(m *Model, _ ActionContext) tea.Cmd {
	m.openHelp()

	return nil
}

// toggleFolderList はフォルダペインの表示状態をトグルする。
func toggleFolderList(m *Model, ctx ActionContext) tea.Cmd {
	return m.toggleFolderList(ctx.Now)
}

// reportResult は操作結果（エラー or 情報メッセージ）を Model に伝える ModelAction を返す。
func reportResult(err error, info string) ModelAction {
	return func(m *Model, _ ActionContext) tea.Cmd {
		if err != nil {
			m.errMsg = err.Error()

			return nil
		}

		return m.setInfoMsg(info)
	}
}

// trashSelectedNote は選択ノートをゴミ箱に移動する。
// NoteList と EditorHeader の両方から emit される。
func trashSelectedNote(m *Model, ctx ActionContext) tea.Cmd {
	m.syncEditorToNote(ctx.Now)
	result, err := m.NoteList.TrashSelected()

	return m.applyNoteAction(result, err, ctx.Now)
}

// duplicateSelectedNote は選択ノートを複製する。
// NoteList と EditorHeader の両方から emit される。
func duplicateSelectedNote(m *Model, ctx ActionContext) tea.Cmd {
	m.syncEditorToNote(ctx.Now)
	result, err := m.NoteList.DuplicateSelected()

	return m.applyNoteAction(result, err, ctx.Now)
}

// copyNoteToClipboard は選択ノート内容をクリップボードにコピーする。
// NoteList と EditorHeader の両方から emit される。
func copyNoteToClipboard(m *Model, ctx ActionContext) tea.Cmd {
	return m.applyAction(m.Editor.CopyToClipboard(), ctx.Now)
}

// createNote は新規ノートを作成する。
// NoteList と EditorHeader の両方から emit される。
func createNote(m *Model, ctx ActionContext) tea.Cmd {
	return m.createNote(ctx.Now)
}
