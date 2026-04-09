package ui

import (
	"time"

	"github.com/bluegreenhq/tnotes/internal/note"
)

const (
	sidebarHeaderLines = 2 // タイトル + 区切り線
	sidebarBorderWidth = 2 // 左右ボーダー分
	sectionLinePadding = 2 // セクション罫線の左右余白
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

// SetNotes はノート一覧を更新する。
func (s *Sidebar) SetNotes(notes []note.Note, now time.Time) {
	s.notes = notes
	if s.selected >= len(notes) {
		s.selected = max(len(notes)-1, 0)
	}

	s.clampOffset(now)
}

// SelectedIndex は選択中のインデックスを返す。
func (s *Sidebar) SelectedIndex() int { return s.selected }

// SelectedNote は選択中のノートを返す。
func (s *Sidebar) SelectedNote() (note.Note, bool) {
	if len(s.notes) == 0 || s.selected >= len(s.notes) {
		return note.ZeroNote(), false
	}

	return s.notes[s.selected], true
}

func (s *Sidebar) visibleLines() int {
	return s.height - sidebarHeaderLines
}

// rowHeight は sidebarRow 1件の表示行数を返す。
func rowHeight(r sidebarRow) int {
	if r.isHeader {
		return sectionHeaderHeight
	}

	return itemHeight
}

// rowsHeightSum は rows[from:to] の合計行数を返す。
func rowsHeightSum(rows []sidebarRow, from, to int) int {
	total := 0
	for i := from; i < to; i++ {
		total += rowHeight(rows[i])
	}

	return total
}

// visibleEndRow は offset から visLines 行に収まる最後のrowインデックス（排他）を返す。
func visibleEndRow(rows []sidebarRow, offset, visLines int) int {
	used := 0

	for i := offset; i < len(rows); i++ {
		h := rowHeight(rows[i])
		if used+h > visLines {
			return i
		}

		used += h
	}

	return len(rows)
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

	targetRow := findSelectedRow(rows, s.selected)
	vis := s.visibleLines()

	// 選択中のアイテムの全行が見えるようにオフセットを調整
	endLines := rowsHeightSum(rows, s.offset, targetRow+1)
	if endLines > vis {
		s.offset = targetRow
		for s.offset > 0 && rowsHeightSum(rows, s.offset-1, targetRow+1) <= vis {
			s.offset--
		}
	}

	if targetRow < s.offset {
		s.offset = targetRow
	}

	s.offset = max(s.offset, 0)
	s.offset = min(s.offset, len(rows)-1)
}

func findSelectedRow(rows []sidebarRow, selected int) int {
	for i, row := range rows {
		if !row.isHeader && row.noteIndex == selected {
			return i
		}
	}

	return 0
}
