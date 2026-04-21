package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/bluegreenhq/tnotes/internal/note"
)

const (
	noteListHeaderLines = 2 // タイトル + 区切り線
	noteListBorderWidth = 1 // 右ボーダー分
	sectionLinePadding  = 2 // セクション罫線の左右余白
)

// NoteList はノートリストの状態を表す。
type NoteList struct {
	notes          []note.Note
	selected       int
	width          int
	height         int
	offset         int
	title          string
	sectioned      bool
	hoverFolderBtn bool
	dirtyNoteID    note.NoteID
	searchQuery    string
}

// NewNoteList は新しい NoteList を生成する。
func NewNoteList(notes []note.Note, width, height int) NoteList {
	return NoteList{
		notes:          notes,
		selected:       0,
		width:          width,
		height:         height,
		offset:         0,
		title:          "Notes",
		sectioned:      true,
		hoverFolderBtn: false,
		dirtyNoteID:    "",
		searchQuery:    "",
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

// SelectedY は選択中ノートの画面上Y座標を返す。
// now はセクション表示の計算に必要。
func (s *NoteList) SelectedY(now time.Time) int {
	rows := s.buildRows(now)
	targetRow := findSelectedRow(rows, s.selected)
	y := noteListHeaderLines

	for i := s.offset; i < len(rows) && i < targetRow; i++ {
		y += rowHeight(rows[i])
	}

	return y
}

// SetDirtyNoteID は未保存状態のノートIDを設定する。未保存がなければ空文字。
func (s *NoteList) SetDirtyNoteID(id note.NoteID) { s.dirtyNoteID = id }

// SetHoverFolderBtn はフォルダボタンのホバー状態を設定する。
func (s *NoteList) SetHoverFolderBtn(v bool) { s.hoverFolderBtn = v }

// SetSearchQuery は検索クエリを設定する。
func (s *NoteList) SetSearchQuery(q string) { s.searchQuery = q }

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

// Update はメッセージに応じてノート一覧の状態を更新する。
// trashMode はゴミ箱モードかどうかを示す。
// ナビゲーション（カーソル移動）は自身で処理し、
// ノート操作（作成、削除等）は tea.Cmd で NoteListMsg を返して Model に委譲する。
func (s *NoteList) Update(msg tea.Msg, now time.Time, trashMode bool) (NoteList, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return *s, nil
	}

	// ナビゲーション共通キー（通常/ゴミ箱モード共通）
	if cmd, handled := s.handleNavKey(keyMsg); handled {
		return *s, cmd
	}

	if keyMsg.Mod&tea.ModCtrl != 0 {
		return s.handleCtrlKey(keyMsg, now)
	}

	if trashMode {
		return s.handleTrashModeKey(keyMsg, now)
	}

	return s.handleNormalKey(keyMsg, now)
}

var noteListStyle = lipgloss.NewStyle().
	BorderRight(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("8"))

// View はノート一覧の描画内容を返す。
func (s *NoteList) View(focused bool, hoverSeparator bool, now time.Time, folderVisible bool) string {
	contentWidth := max(s.width-noteListBorderWidth, 0)

	var b strings.Builder

	s.writeHeader(&b, contentWidth, folderVisible)

	rows := s.buildRows(now)
	visEnd := visibleEndRow(rows, s.offset, s.visibleLines())

	usedLines := 2

	for i := s.offset; i < visEnd; i++ {
		row := rows[i]
		if row.isHeader {
			b.WriteString(sectionHeaderStyle.Width(contentWidth).Render(" " + row.label))
			b.WriteString("\n")

			line := " " + strings.Repeat("─", max(contentWidth-sectionLinePadding, 0))
			b.WriteString(sectionHeaderStyle.Width(contentWidth).Render(line))
			b.WriteString("\n")

			usedLines += sectionHeaderHeight

			continue
		}

		isDirty := s.dirtyNoteID != "" && row.note.ID == s.dirtyNoteID
		b.WriteString(renderItem(row.note, row.noteIndex == s.selected, isDirty, contentWidth, now, s.searchQuery))

		usedLines += itemHeight
	}

	for i := usedLines; i < s.height; i++ {
		b.WriteString("\n")
	}

	style := noteListStyle

	if focused {
		style = style.BorderForeground(lipgloss.Color("4"))
	}

	if hoverSeparator {
		style = style.BorderStyle(lipgloss.ThickBorder())
	}

	return style.Width(s.width).Height(s.height).Render(b.String())
}

func (s *NoteList) visibleLines() int {
	return s.height - noteListHeaderLines
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

func (s *NoteList) handleNavKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	switch {
	case msg.Code == 'q' && msg.Mod == 0:
		return NoteListQuit.Cmd(), true
	case msg.Code == tea.KeyTab:
		return NoteListEdit.Cmd(), true
	case msg.Code == tea.KeyEscape:
		return NoteListFocusPrev.Cmd(), true
	case msg.Code == 'b' && msg.Mod&tea.ModCtrl != 0:
		return NoteListToggleFolder.Cmd(), true
	case msg.Code == '?' && msg.Mod == 0:
		return NoteListHelp.Cmd(), true
	}

	return nil, false
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

func (s *NoteList) writeHeader(b *strings.Builder, contentWidth int, folderVisible bool) {
	var titleName string
	if folderVisible {
		titleName = " " + s.title
	} else {
		folderBtn := "≡"
		if s.hoverFolderBtn {
			folderBtn = buttonHoverStyle.Render(folderBtn)
		} else {
			folderBtn = buttonStyle.Render(folderBtn)
		}

		titleName = " " + folderBtn + " " + s.title
	}

	count := fmt.Sprintf("%d ", len(s.notes))
	titleName = truncateForCount(titleName, count, contentWidth)
	padding := max(contentWidth-lipgloss.Width(titleName)-lipgloss.Width(count), 0)

	titleStr := lipgloss.NewStyle().Bold(true).Render(titleName) +
		strings.Repeat(" ", padding) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(count)

	b.WriteString(titleStr)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", contentWidth))
	b.WriteString("\n")
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

func findSelectedRow(rows []noteListRow, selected int) int {
	for i, row := range rows {
		if !row.isHeader && row.noteIndex == selected {
			return i
		}
	}

	return 0
}
