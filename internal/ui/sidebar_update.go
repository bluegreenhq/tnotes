package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/note"
)

// Reset はサイドバーの状態をリセットし、先頭を選択する。
func (s *Sidebar) Reset(title string, sectioned bool, notes []note.Note, now time.Time) {
	s.SetTitle(title)
	s.SetSectioned(sectioned)
	s.SetNotes(notes, now)

	if len(notes) > 0 {
		s.SelectIndex(0, now)
	}
}

// SetSize はサイズを更新する。
func (s *Sidebar) SetSize(width, height int, now time.Time) {
	s.width = width
	s.height = height
	s.clampOffset(now)
}

// SelectIndex はインデックスを指定して選択する。
func (s *Sidebar) SelectIndex(idx int, now time.Time) {
	if idx >= 0 && idx < len(s.notes) {
		s.selected = idx
		s.clampOffset(now)
	}
}

// MoveUp は選択を1つ上に移動する。
func (s *Sidebar) MoveUp(now time.Time) {
	if s.selected > 0 {
		s.selected--
		s.clampOffset(now)
	}
}

// MoveDown は選択を1つ下に移動する。
func (s *Sidebar) MoveDown(now time.Time) {
	if s.selected < len(s.notes)-1 {
		s.selected++
		s.clampOffset(now)
	}
}

// SetTitle はサイドバーのヘッダータイトルを設定する。
func (s *Sidebar) SetTitle(t string) { s.title = t }

// SetSectioned はセクション分け表示を切り替える。
func (s *Sidebar) SetSectioned(v bool) { s.sectioned = v }

// HitTest は座標からクリックされたノートのインデックスを返す。該当なしは -1。
func (s *Sidebar) HitTest(x, y int, now time.Time) int {
	if x < 0 || x >= s.width {
		return -1
	}

	contentY := y - sidebarHeaderLines
	if contentY < 0 {
		return -1
	}

	rows := s.buildRows(now)

	currentLine := 0

	for i := s.offset; i < len(rows); i++ {
		row := rows[i]

		var rowHeight int
		if row.isHeader {
			rowHeight = sectionHeaderHeight
		} else {
			rowHeight = itemHeight
		}

		if contentY >= currentLine && contentY < currentLine+rowHeight {
			if row.isHeader {
				return -1
			}

			return row.noteIndex
		}

		currentLine += rowHeight
	}

	return -1
}

// ScrollUp は表示オフセットを n 行上にスクロールする。選択は変更しない。
func (s *Sidebar) ScrollUp(n int, now time.Time) {
	s.offset = max(s.offset-n, 0)
	s.clampScrollOffset(now)
}

// ScrollDown は表示オフセットを n 行下にスクロールする。選択は変更しない。
func (s *Sidebar) ScrollDown(n int, now time.Time) {
	s.offset += n
	s.clampScrollOffset(now)
}

func (s *Sidebar) clampScrollOffset(now time.Time) {
	rows := s.buildRows(now)
	if len(rows) == 0 {
		s.offset = 0

		return
	}

	vis := s.visibleLines()

	// 末尾が見える最大オフセットを求める
	maxOffset := len(rows)
	for maxOffset > 0 && rowsHeightSum(rows, maxOffset-1, len(rows)) <= vis {
		maxOffset--
	}

	if s.offset > maxOffset {
		s.offset = maxOffset
	}

	if s.offset < 0 {
		s.offset = 0
	}
}

// Update はメッセージに応じてサイドバーの状態を更新する。
// trashMode はゴミ箱モードかどうかを示す。
// ナビゲーション（カーソル移動）は自身で処理し、
// ノート操作（作成、削除等）は tea.Cmd で SidebarMsg を返して Model に委譲する。
func (s *Sidebar) Update(msg tea.Msg, now time.Time, trashMode bool) (Sidebar, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return *s, nil
	}

	if keyMsg.Mod&tea.ModCtrl != 0 {
		return s.handleCtrlKey(keyMsg, now)
	}

	if trashMode {
		return s.handleTrashModeKey(keyMsg, now)
	}

	return s.handleNormalKey(keyMsg, now)
}

func (s *Sidebar) handleNormalKey(msg tea.KeyPressMsg, now time.Time) (Sidebar, tea.Cmd) {
	switch msg.Code {
	case 'n':
		return *s, sidebarCmd(SidebarCreate)
	case 'd', tea.KeyDelete, tea.KeyBackspace:
		return *s, sidebarCmd(SidebarTrash)
	case 'g':
		return *s, sidebarCmd(SidebarEnterTrash)
	case tea.KeyUp, 'k':
		return s.moveUpCmd(now)
	case tea.KeyDown, 'j':
		return s.moveDownCmd(now)
	case tea.KeyEnter:
		return *s, sidebarCmd(SidebarEdit)
	}

	return *s, nil
}

func (s *Sidebar) handleTrashModeKey(msg tea.KeyPressMsg, now time.Time) (Sidebar, tea.Cmd) {
	switch msg.Code {
	case 'g':
		return *s, sidebarCmd(SidebarExitTrash)
	case 'r':
		return *s, sidebarCmd(SidebarRestore)
	case tea.KeyUp, 'k':
		return s.moveUpCmd(now)
	case tea.KeyDown, 'j':
		return s.moveDownCmd(now)
	}

	return *s, nil
}

func (s *Sidebar) handleCtrlKey(msg tea.KeyPressMsg, now time.Time) (Sidebar, tea.Cmd) {
	switch msg.Code {
	case 'z':
		if msg.Mod&tea.ModShift != 0 {
			return *s, sidebarCmd(SidebarRedo)
		}

		return *s, sidebarCmd(SidebarUndo)
	case 'n':
		return s.moveDownCmd(now)
	case 'p':
		return s.moveUpCmd(now)
	}

	return *s, nil
}

func (s *Sidebar) moveUpCmd(now time.Time) (Sidebar, tea.Cmd) {
	prev := s.selected
	s.MoveUp(now)

	if s.selected != prev {
		return *s, sidebarCmd(SidebarSelect)
	}

	return *s, nil
}

func (s *Sidebar) moveDownCmd(now time.Time) (Sidebar, tea.Cmd) {
	prev := s.selected
	s.MoveDown(now)

	if s.selected != prev {
		return *s, sidebarCmd(SidebarSelect)
	}

	return *s, nil
}

func sidebarCmd(msg SidebarMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}
