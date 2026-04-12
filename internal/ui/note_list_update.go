package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/bluegreenhq/tnotes/internal/note"
)

// Reset はノート一覧の状態をリセットし、先頭を選択する。
func (s *NoteList) Reset(title string, sectioned bool, notes []note.Note, now time.Time) {
	s.SetTitle(title)
	s.SetSectioned(sectioned)
	s.SetNotes(notes, now)

	if len(notes) > 0 {
		s.SelectIndex(0, now)
	}
}

// SetSize はサイズを更新する。
func (s *NoteList) SetSize(width, height int, now time.Time) {
	s.width = width
	s.height = height
	s.clampOffset(now)
}

// SelectIndex はインデックスを指定して選択する。
func (s *NoteList) SelectIndex(idx int, now time.Time) {
	if idx >= 0 && idx < len(s.notes) {
		s.selected = idx
		s.clampOffset(now)
	}
}

// MoveUp は選択を1つ上に移動する。
func (s *NoteList) MoveUp(now time.Time) {
	idx := s.adjacentNoteIndex(now, -1)
	if idx >= 0 {
		s.selected = idx
		s.clampOffset(now)
	}
}

// MoveDown は選択を1つ下に移動する。
func (s *NoteList) MoveDown(now time.Time) {
	idx := s.adjacentNoteIndex(now, 1)
	if idx >= 0 {
		s.selected = idx
		s.clampOffset(now)
	}
}

// adjacentNoteIndex はセクション表示の行順序で前後のノートインデックスを返す。
// direction は -1（上）または 1（下）。該当なしは -1 を返す。
func (s *NoteList) adjacentNoteIndex(now time.Time, direction int) int {
	rows := s.buildRows(now)
	cur := findSelectedRow(rows, s.selected)

	for i := cur + direction; i >= 0 && i < len(rows); i += direction {
		if !rows[i].isHeader {
			return rows[i].noteIndex
		}
	}

	return -1
}

// SetTitle はノート一覧のヘッダータイトルを設定する。
func (s *NoteList) SetTitle(t string) { s.title = t }

// SetSectioned はセクション分け表示を切り替える。
func (s *NoteList) SetSectioned(v bool) { s.sectioned = v }

// HitTest は座標からクリックされたノートのインデックスを返す。該当なしは -1。
func (s *NoteList) HitTest(x, y int, now time.Time) int {
	if x < 0 || x >= s.width {
		return -1
	}

	contentY := y - noteListHeaderLines
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
func (s *NoteList) ScrollUp(n int, now time.Time) {
	s.offset = max(s.offset-n, 0)
	s.clampScrollOffset(now)
}

// ScrollDown は表示オフセットを n 行下にスクロールする。選択は変更しない。
func (s *NoteList) ScrollDown(n int, now time.Time) {
	s.offset += n
	s.clampScrollOffset(now)
}

func (s *NoteList) clampScrollOffset(now time.Time) {
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

// Update はメッセージに応じてノート一覧の状態を更新する。
// trashMode はゴミ箱モードかどうかを示す。
// ナビゲーション（カーソル移動）は自身で処理し、
// ノート操作（作成、削除等）は tea.Cmd で NoteListMsg を返して Model に委譲する。
func (s *NoteList) Update(msg tea.Msg, now time.Time, trashMode bool) (NoteList, tea.Cmd) {
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

func (s *NoteList) handleNormalKey(msg tea.KeyPressMsg, now time.Time) (NoteList, tea.Cmd) {
	switch msg.Code {
	case 'n':
		return *s, NoteListCreate.Cmd()
	case 'd', tea.KeyDelete, tea.KeyBackspace:
		return *s, NoteListTrash.Cmd()
	case 'm':
		return *s, NoteListMenu.Cmd()
	case tea.KeyUp, 'k':
		return s.moveUpCmd(now)
	case tea.KeyDown, 'j':
		return s.moveDownCmd(now)
	case tea.KeyEnter:
		return *s, NoteListEdit.Cmd()
	}

	return *s, nil
}

func (s *NoteList) handleTrashModeKey(msg tea.KeyPressMsg, now time.Time) (NoteList, tea.Cmd) {
	switch msg.Code {
	case 'm':
		return *s, NoteListMenu.Cmd()
	case tea.KeyUp, 'k':
		return s.moveUpCmd(now)
	case tea.KeyDown, 'j':
		return s.moveDownCmd(now)
	}

	return *s, nil
}

func (s *NoteList) handleCtrlKey(msg tea.KeyPressMsg, now time.Time) (NoteList, tea.Cmd) {
	switch msg.Code {
	case 'z':
		if msg.Mod&tea.ModShift != 0 {
			return *s, NoteListRedo.Cmd()
		}

		return *s, NoteListUndo.Cmd()
	case 'c':
		return *s, NoteListCopy.Cmd()
	case 'd':
		return *s, NoteListDuplicate.Cmd()
	case 'n':
		return s.moveDownCmd(now)
	case 'p':
		return s.moveUpCmd(now)
	}

	return *s, nil
}

func (s *NoteList) moveUpCmd(now time.Time) (NoteList, tea.Cmd) {
	prev := s.selected
	s.MoveUp(now)

	if s.selected != prev {
		return *s, NoteListSelect.Cmd()
	}

	return *s, nil
}

func (s *NoteList) moveDownCmd(now time.Time) (NoteList, tea.Cmd) {
	prev := s.selected
	s.MoveDown(now)

	if s.selected != prev {
		return *s, NoteListSelect.Cmd()
	}

	return *s, nil
}
