package ui

import (
	"time"

	"github.com/bluegreenhq/tnotes/internal/note"
)

const (
	sidebarHeaderLines = 2 // タイトル + 区切り線
	sidebarBorderWidth = 2 // 左右ボーダー分
)

// Sidebar はノートリストの状態を表す。
type Sidebar struct {
	notes     []note.Note
	selected  int
	width     int
	height    int
	offset    int
	title     string
	sectioned bool
}

// NewSidebar は新しい Sidebar を生成する。
func NewSidebar(notes []note.Note, width, height int) Sidebar {
	return Sidebar{
		notes:     notes,
		selected:  0,
		width:     width,
		height:    height,
		offset:    0,
		title:     "Notes",
		sectioned: true,
	}
}

// Reset はサイドバーの状態をリセットし、先頭を選択する。
func (s *Sidebar) Reset(title string, sectioned bool, notes []note.Note, now time.Time) {
	s.SetTitle(title)
	s.SetSectioned(sectioned)
	s.SetNotes(notes, now)

	if len(notes) > 0 {
		s.SelectIndex(0, now)
	}
}

// SetNotes はノート一覧を更新する。
func (s *Sidebar) SetNotes(notes []note.Note, now time.Time) {
	s.notes = notes
	if s.selected >= len(notes) {
		s.selected = max(len(notes)-1, 0)
	}

	s.clampOffset(now)
}

// SetSize はサイズを更新する。
func (s *Sidebar) SetSize(width, height int, now time.Time) {
	s.width = width
	s.height = height
	s.clampOffset(now)
}

// SelectedIndex は選択中のインデックスを返す。
func (s *Sidebar) SelectedIndex() int { return s.selected }

// SelectIndex はインデックスを指定して選択する。
func (s *Sidebar) SelectIndex(idx int, now time.Time) {
	if idx >= 0 && idx < len(s.notes) {
		s.selected = idx
		s.clampOffset(now)
	}
}

// SelectedNote は選択中のノートを返す。
func (s *Sidebar) SelectedNote() (note.Note, bool) {
	if len(s.notes) == 0 || s.selected >= len(s.notes) {
		return note.ZeroNote(), false
	}

	return s.notes[s.selected], true
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

func (s *Sidebar) visibleRows() int {
	return s.height - sidebarHeaderLines
}

func (s *Sidebar) buildRows(now time.Time) []sidebarRow {
	if !s.sectioned {
		return s.buildFlatRows()
	}

	sections := GroupNotesBySection(s.notes, now)
	if len(sections) == 0 {
		return nil
	}

	noteIndexMap := make(map[note.NoteID]int, len(s.notes))
	for i, n := range s.notes {
		noteIndexMap[n.ID] = i
	}

	var rows []sidebarRow
	for _, sec := range sections {
		rows = append(rows, newHeaderRow(sec.Label))
		for _, n := range sec.Notes {
			rows = append(rows, newNoteRow(n, noteIndexMap[n.ID]))
		}
	}

	return rows
}

func (s *Sidebar) buildFlatRows() []sidebarRow {
	rows := make([]sidebarRow, len(s.notes))
	for i, n := range s.notes {
		rows[i] = newNoteRow(n, i)
	}

	return rows
}

func (s *Sidebar) clampOffset(now time.Time) {
	rows := s.buildRows(now)
	if len(rows) == 0 {
		s.offset = 0

		return
	}

	// 選択中のノートに対応するrowインデックスを見つける
	targetRow := 0

	for i, row := range rows {
		if !row.isHeader && row.noteIndex == s.selected {
			targetRow = i

			break
		}
	}

	vis := s.visibleRows()

	// 選択中のアイテムの全行が見えるようにオフセットを調整
	itemEnd := targetRow + itemHeight
	if itemEnd > s.offset+vis {
		s.offset = itemEnd - vis
	}

	if targetRow < s.offset {
		s.offset = targetRow
	}

	if s.offset < 0 {
		s.offset = 0
	}

	if s.offset >= len(rows) {
		s.offset = len(rows) - 1
	}
}
