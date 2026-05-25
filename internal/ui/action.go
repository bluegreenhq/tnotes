package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// ModelAction はペインからモデルへの「やってほしいこと」を表す。
// pane が UpdatePane で返す tea.Cmd の搬送先として用いる。
type ModelAction interface {
	Apply(m *Model, ctx ActionContext) tea.Cmd
}

// ActionContext は Apply に渡す共通コンテキスト。
type ActionContext struct {
	Now time.Time
}

// actionMsg は ModelAction を tea.Cmd で運ぶための内部キャリア。
type actionMsg struct{ action ModelAction }

// actionCmd は ModelAction を tea.Cmd 化するヘルパー。
func actionCmd(a ModelAction) tea.Cmd {
	return func() tea.Msg { return actionMsg{action: a} }
}

// --- 共通アクション（複数のペイン／オーバーレイから emit される） ---

// quitAction はアプリケーション終了を Model に要求する。
type quitAction struct{}

// Apply は編集中ノートを保存してから tea.Quit を返す。
func (quitAction) Apply(m *Model, ctx ActionContext) tea.Cmd {
	m.syncEditorToNote(ctx.Now)

	return tea.Quit
}

// openHelpAction はショートカットヘルプ表示を Model に要求する。
type openHelpAction struct{}

// Apply はヘルプオーバーレイを開く。
func (openHelpAction) Apply(m *Model, _ ActionContext) tea.Cmd {
	m.openHelp()

	return nil
}

// toggleFolderListAction はフォルダ一覧の表示切り替えを Model に要求する。
type toggleFolderListAction struct{}

// Apply はフォルダペインの表示状態をトグルする。
func (toggleFolderListAction) Apply(m *Model, ctx ActionContext) tea.Cmd {
	return m.toggleFolderList(ctx.Now)
}

// resultAction はコンポーネント操作の結果（エラー or 情報メッセージ）を Model に伝える。
type resultAction struct {
	Err  error
	Info string
}

// Apply はエラーメッセージ／情報メッセージをセットする。
func (a resultAction) Apply(m *Model, _ ActionContext) tea.Cmd {
	if a.Err != nil {
		m.errMsg = a.Err.Error()

		return nil
	}

	return m.setInfoMsg(a.Info)
}
