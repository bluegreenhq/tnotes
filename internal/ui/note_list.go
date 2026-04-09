package ui

import (
	"time"

	"github.com/bluegreenhq/tnotes/internal/note"
)

const (
	noteListHeaderLines = 2 // タイトル + 区切り線
	noteListBorderWidth = 2 // 左右ボーダー分
	sectionLinePadding = 2 // セクション罫線の左右余白
)

// NoteList はノートリストの状態を表す。
type NoteList struct {
	notes     []note.Note
	selected  int
	width     int
	height    int
	offset    int
	title     string
	sectioned bool
}

// NewNoteList は新しい NoteList を生成する。
func NewNoteList(notes []note.Note, width, height int) NoteList {
	return NoteList{
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
func (s *NoteList) SetNotes(notes []note.Note, now time.Time) {
	s.notes = notes
	if s.selected >= len(notes) {
		s.selected = max(len(notes)-1, 0)
	}

	s.clampOffset(now)
}

// SelectedIndex は選択中のインデックスを返す。
func (s *NoteList) SelectedIndex() int { return s.selected }

// SelectedNote は選択中のノートを返す。
func (s *NoteList) SelectedNote() (note.Note, bool) {
	if len(s.notes) == 0 || s.selected >= len(s.notes) {
		return note.ZeroNote(), false
	}

	return s.notes[s.selected], true
}

func (s *NoteList) visibleLines() int {
	return s.height - noteListHeaderLines
}

// rowHeight は noteListRow 1件の表示行数を返す。
func rowHeight(r noteListRow) int {
	if r.isHeader {
		return sectionHeaderHeight
	}

	return itemHeight
}

// rowsHeightSum は rows[from:to] の合計行数を返す。
func rowsHeightSum(rows []noteListRow, from, to int) int {
	total := 0
	for i := from; i < to; i++ {
		total += rowHeight(rows[i])
	}

	return total
}

// visibleEndRow は offset から visLines 行に収まる最後のrowインデックス（排他）を返す。
func visibleEndRow(rows []noteListRow, offset, visLines int) int {
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

func (s *NoteList) buildRows(now time.Time) []noteListRow {
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

	var rows []noteListRow
	for _, sec := range sections {
		rows = append(rows, newHeaderRow(sec.Label))
		for _, n := range sec.Notes {
			rows = append(rows, newNoteRow(n, noteIndexMap[n.ID]))
		}
	}

	return rows
}

func (s *NoteList) buildFlatRows() []noteListRow {
	rows := make([]noteListRow, len(s.notes))
	for i, n := range s.notes {
		rows[i] = newNoteRow(n, i)
	}

	return rows
}

func (s *NoteList) clampOffset(now time.Time) {
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

func findSelectedRow(rows []noteListRow, selected int) int {
	for i, row := range rows {
		if !row.isHeader && row.noteIndex == selected {
			return i
		}
	}

	return 0
}
