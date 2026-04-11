package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

const blinkInterval = 500 * time.Millisecond

// cursorBlinkOwner は cursorBlink の所有者を識別する。
type cursorBlinkOwner int

const (
	blinkOwnerEditor cursorBlinkOwner = iota
	blinkOwnerFolderList
)

// cursorBlinkMsg はカーソルの点滅状態を切り替えるメッセージ。
type cursorBlinkMsg struct {
	owner cursorBlinkOwner
	tag   int
}

// cursorBlink はカーソルの点滅状態を管理する。
type cursorBlink struct {
	visible bool
	tag     int
	owner   cursorBlinkOwner
}

// newCursorBlink は新しい cursorBlink を生成する。
func newCursorBlink(owner cursorBlinkOwner) cursorBlink {
	return cursorBlink{visible: true, tag: 0, owner: owner}
}

// Visible はカーソルが表示状態かを返す。
func (cb *cursorBlink) Visible() bool { return cb.visible }

// Reset はカーソルを表示状態にリセットし、新しい blink タイマーを開始する。
func (cb *cursorBlink) Reset() tea.Cmd {
	cb.visible = true
	cb.tag++
	tag := cb.tag
	owner := cb.owner

	return tea.Tick(blinkInterval, func(_ time.Time) tea.Msg {
		return cursorBlinkMsg{owner: owner, tag: tag}
	})
}

// Stop はカーソルを表示状態にし、blink を停止する。
func (cb *cursorBlink) Stop() {
	cb.visible = true
	cb.tag++
}

// HandleMsg は cursorBlinkMsg を処理して blink 状態を切り替える。
// owner または tag が一致しない場合は無視して nil を返す。
func (cb *cursorBlink) HandleMsg(msg cursorBlinkMsg) tea.Cmd {
	if msg.owner != cb.owner || msg.tag != cb.tag {
		return nil
	}

	cb.visible = !cb.visible
	tag := cb.tag
	owner := cb.owner

	return tea.Tick(blinkInterval, func(_ time.Time) tea.Msg {
		return cursorBlinkMsg{owner: owner, tag: tag}
	})
}
