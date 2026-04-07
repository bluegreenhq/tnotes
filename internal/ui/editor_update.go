package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/x/ansi"
	"github.com/cockroachdb/errors"
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/utils"
)

// --- Update ---

// Update はメッセージに応じて状態を更新する。
func (e *Editor) Update(msg tea.Msg, now time.Time) (Editor, tea.Cmd) {
	if e.readOnly {
		return *e, nil
	}

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		return e.handleKey(msg, now)
	}

	prevText := e.textarea.Value()
	prevLine := e.textarea.Line()
	prevCol := e.textarea.Column()

	var cmd tea.Cmd

	e.textarea, cmd = e.textarea.Update(msg)

	newText := e.textarea.Value()
	if newText != prevText {
		e.saveSnapshotBefore(prevText, prevLine, prevCol, false, now)
	}

	return *e, cmd
}

func (e *Editor) handleKey(msg tea.KeyPressMsg, now time.Time) (Editor, tea.Cmd) { //nolint:cyclop // キーバインド分岐
	// Ctrl+Z: Undo
	if msg.Code == 'z' && msg.Mod == tea.ModCtrl {
		e.Undo()

		return *e, nil
	}

	// Ctrl+Shift+Z: Redo
	if msg.Code == 'z' && msg.Mod == (tea.ModCtrl|tea.ModShift) {
		e.Redo()

		return *e, nil
	}

	// Ctrl+A: 全選択
	if msg.Code == 'a' && msg.Mod == tea.ModCtrl {
		e.SelectAll()

		return *e, nil
	}

	isArrow := msg.Code == tea.KeyLeft || msg.Code == tea.KeyRight ||
		msg.Code == tea.KeyUp || msg.Code == tea.KeyDown ||
		msg.Code == tea.KeyHome || msg.Code == tea.KeyEnd

	if isArrow && msg.Mod == tea.ModShift {
		cmd := e.handleShiftArrow(msg)

		return *e, cmd
	}

	prevText := e.textarea.Value()
	prevLine := e.textarea.Line()
	prevCol := e.textarea.Column()

	if e.HasSelection() {
		switch {
		case isArrow:
			e.ClearSelection()
		case msg.Code == tea.KeyBackspace || msg.Code == tea.KeyDelete:
			e.saveSnapshotBefore(prevText, prevLine, prevCol, true, now)
			e.DeleteSelection()

			return *e, nil
		case msg.Text != "":
			e.DeleteSelection()
		}
	}

	forceSnapshot := msg.Code == tea.KeyEnter || msg.Code == tea.KeyBackspace || msg.Code == tea.KeyDelete

	var cmd tea.Cmd

	e.textarea, cmd = e.textarea.Update(msg)

	newText := e.textarea.Value()
	if newText != prevText {
		e.saveSnapshotBefore(prevText, prevLine, prevCol, forceSnapshot, now)
	}

	return *e, cmd
}

func (e *Editor) handleShiftArrow(msg tea.KeyPressMsg) tea.Cmd {
	if !e.HasSelection() {
		anchor := SelectionAnchor{Line: e.textarea.Line(), Column: e.textarea.Column()}
		e.selStart = &anchor
	}

	plainMsg := tea.KeyPressMsg{Code: msg.Code, Mod: 0}

	var cmd tea.Cmd

	e.textarea, cmd = e.textarea.Update(plainMsg)

	newPos := SelectionAnchor{Line: e.textarea.Line(), Column: e.textarea.Column()}
	e.selEnd = &newPos

	return cmd
}

// --- 選択 ---

// HasSelection は選択範囲があるかを返す。
func (e *Editor) HasSelection() bool {
	return e.selStart != nil && e.selEnd != nil && *e.selStart != *e.selEnd
}

// SetSelection は選択範囲を設定する。
func (e *Editor) SetSelection(start, end SelectionAnchor) {
	e.selStart = &start
	e.selEnd = &end
}

// ClearSelection は選択範囲を解除する。
func (e *Editor) ClearSelection() {
	e.selStart = nil
	e.selEnd = nil
}

// SelectAll はテキスト全体を選択する。
func (e *Editor) SelectAll() {
	lines := strings.Split(e.textarea.Value(), "\n")

	lastLine := len(lines) - 1
	if lastLine < 0 {
		return
	}

	start := SelectionAnchor{Line: 0, Column: 0}
	end := SelectionAnchor{Line: lastLine, Column: len([]rune(lines[lastLine]))}
	e.SetSelection(start, end)
}

// NormalizedSelection は開始 < 終了に正規化した選択範囲を返す。
// HasSelection() == false の場合の動作は未定義。
func (e *Editor) NormalizedSelection() (SelectionAnchor, SelectionAnchor) {
	s, end := *e.selStart, *e.selEnd
	if selBefore(end, s) {
		s, end = end, s
	}

	return s, end
}

// SelectedText は選択範囲のテキストを返す。選択なしの場合は空文字列を返す。
func (e *Editor) SelectedText() string {
	if !e.HasSelection() {
		return ""
	}

	start, end := e.NormalizedSelection()
	lines := strings.Split(e.textarea.Value(), "\n")

	if start.Line == end.Line {
		line := lines[start.Line]
		runes := []rune(line)
		from := utils.ClampInt(start.Column, 0, len(runes))
		to := utils.ClampInt(end.Column, 0, len(runes))

		return string(runes[from:to])
	}

	var b strings.Builder

	startRunes := []rune(lines[start.Line])
	from := utils.ClampInt(start.Column, 0, len(startRunes))
	b.WriteString(string(startRunes[from:]))

	for i := start.Line + 1; i < end.Line; i++ {
		b.WriteString("\n")
		b.WriteString(lines[i])
	}

	b.WriteString("\n")

	endRunes := []rune(lines[end.Line])
	to := utils.ClampInt(end.Column, 0, len(endRunes))
	b.WriteString(string(endRunes[:to]))

	return b.String()
}

// CopySelection は選択範囲のテキストをシステムクリップボードにコピーする。
func (e *Editor) CopySelection() error {
	text := e.SelectedText()
	if text == "" {
		return nil
	}

	err := clipboard.WriteAll(text)
	if err != nil {
		return errors.WithStack(err)
	}

	e.ClearSelection()

	return nil
}

// CutSelection は選択範囲のテキストをシステムクリップボードにコピーし、テキストを削除する。
func (e *Editor) CutSelection() error {
	text := e.SelectedText()
	if text == "" {
		return nil
	}

	err := clipboard.WriteAll(text)
	if err != nil {
		return errors.WithStack(err)
	}

	e.DeleteSelection()

	return nil
}

// DeleteSelection は選択範囲のテキストを削除する。選択なしの場合は何もしない。
func (e *Editor) DeleteSelection() {
	if !e.HasSelection() {
		return
	}

	start, end := e.NormalizedSelection()
	lines := strings.Split(e.textarea.Value(), "\n")

	startRunes := []rune(lines[start.Line])
	endRunes := []rune(lines[end.Line])

	from := utils.ClampInt(start.Column, 0, len(startRunes))
	to := utils.ClampInt(end.Column, 0, len(endRunes))

	var merged strings.Builder
	merged.WriteString(string(startRunes[:from]))
	merged.WriteString(string(endRunes[to:]))

	result := make([]string, 0, start.Line+1+(len(lines)-end.Line-1))
	result = append(result, lines[:start.Line]...)
	result = append(result, merged.String())
	result = append(result, lines[end.Line+1:]...)

	e.textarea.SetValue(strings.Join(result, "\n"))

	// カーソルを選択開始位置に移動
	e.textarea.MoveToBegin()

	for range start.Line {
		e.textarea.CursorDown()
	}

	e.textarea.SetCursorColumn(start.Column)

	e.ClearSelection()
}

// Selecting はドラッグ中かを返す。
func (e *Editor) Selecting() bool { return e.selecting }

// StartDragSelection はドラッグ選択を開始する。
func (e *Editor) StartDragSelection(x, y int) {
	pos := e.positionFromMouse(x, y)
	e.moveCursorTo(pos)
	e.selStart = &pos
	e.selEnd = &pos
	e.selecting = true
}

// StopDragSelection はドラッグ選択を終了する。
func (e *Editor) StopDragSelection() { e.selecting = false }

// UpdateDragSelection はドラッグ中の選択範囲を更新する。
func (e *Editor) UpdateDragSelection(x, y int) {
	if e.selStart == nil {
		return
	}

	pos := e.positionFromMouse(x, y)
	e.selEnd = &pos
}

func (e *Editor) positionFromMouse(x, y int) SelectionAnchor {
	col := max(x-1, 0) // padding分を差し引き

	line := y + e.textarea.ScrollYOffset()

	lineCount := e.textarea.LineCount()
	if lineCount == 0 {
		return SelectionAnchor{Line: 0, Column: 0}
	}

	if line >= lineCount {
		line = lineCount - 1
	}

	lines := strings.Split(e.textarea.Value(), "\n")
	if line < len(lines) {
		col = cellToRuneIndex([]rune(lines[line]), col)
	}

	return SelectionAnchor{Line: line, Column: col}
}

func (e *Editor) moveCursorTo(pos SelectionAnchor) {
	e.textarea.MoveToBegin()

	for range pos.Line {
		e.textarea.CursorDown()
	}

	e.textarea.SetCursorColumn(pos.Column)
}

func (e *Editor) applySelectionHighlight(raw string) string {
	selectionStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15"))

	if !e.HasSelection() {
		return raw
	}

	start, end := e.NormalizedSelection()
	scrollOffset := e.textarea.ScrollYOffset()

	contentLines := strings.Split(e.textarea.Value(), "\n")
	viewLines := strings.Split(raw, "\n")

	for i, line := range viewLines {
		logicalLine := i + scrollOffset

		if logicalLine < start.Line || logicalLine > end.Line {
			continue
		}

		// Determine rune-based column range for this line
		var runes []rune
		if logicalLine < len(contentLines) {
			runes = []rune(contentLines[logicalLine])
		} else {
			runes = []rune(ansi.Strip(line))
		}

		var colStart, colEnd int
		if logicalLine == start.Line {
			colStart = start.Column
		} else {
			colStart = 0
		}

		if logicalLine == end.Line {
			colEnd = end.Column
		} else {
			colEnd = len(runes)
		}

		colStart = utils.ClampInt(colStart, 0, len(runes))
		colEnd = utils.ClampInt(colEnd, 0, len(runes))

		if colStart >= colEnd {
			continue
		}

		// ルーンインデックスをセル幅位置に変換し、ビュー行からハイライト部分を切り出す
		cellStart := runewidth.StringWidth(string(runes[:colStart]))
		cellEnd := runewidth.StringWidth(string(runes[:colEnd]))
		totalWidth := ansi.StringWidth(line)

		before := ansi.Cut(line, 0, cellStart)
		middle := ansi.Cut(line, cellStart, cellEnd)
		after := ansi.Cut(line, cellEnd, totalWidth)

		// NOTE: bubbletea のレンダラーが全角文字境界での SGR 変更を正しく
		// 再描画しない既知の問題があり、ハイライト末尾が全角文字の場合に
		// 右半分のセルが更新されないことがある。
		viewLines[i] = before + selectionStyle.Render(ansi.Strip(middle)) + after
	}

	return strings.Join(viewLines, "\n")
}

// cellToRuneIndex はセル幅の位置をルーンインデックスに変換する。
// 全角文字（幅2セル）を考慮して正確なルーン位置を返す。
func cellToRuneIndex(runes []rune, cellCol int) int {
	w := 0

	for i, r := range runes {
		rw := runewidth.RuneWidth(r)
		if w+rw > cellCol {
			return i
		}

		w += rw
		if w >= cellCol {
			return i + 1
		}
	}

	return len(runes)
}

// --- Undo/Redo ---

// SaveSnapshot はエディタのスナップショットをデバウンス付きで保存する。
func (e *Editor) SaveSnapshot(now time.Time) {
	snap := EditorSnapshot{
		Text:       e.textarea.Value(),
		CursorLine: e.textarea.Line(),
		CursorCol:  e.textarea.Column(),
	}
	e.UndoMgr.MaybeSave(snap, now)
}

// ForceSaveSnapshot はデバウンスなしでスナップショットを保存する。
func (e *Editor) ForceSaveSnapshot(now time.Time) {
	snap := EditorSnapshot{
		Text:       e.textarea.Value(),
		CursorLine: e.textarea.Line(),
		CursorCol:  e.textarea.Column(),
	}
	e.UndoMgr.ForceSave(snap, now)
}

// Undo はエディタの状態を1つ前に戻す。
func (e *Editor) Undo() {
	snap := e.UndoMgr.PopUndo()
	if snap == nil {
		return
	}

	current := EditorSnapshot{
		Text:       e.textarea.Value(),
		CursorLine: e.textarea.Line(),
		CursorCol:  e.textarea.Column(),
	}
	e.UndoMgr.PushRedo(current)
	e.restoreSnapshot(snap)
}

// Redo はエディタの状態を1つ先に進める。
func (e *Editor) Redo() {
	snap := e.UndoMgr.PopRedo()
	if snap == nil {
		return
	}

	current := EditorSnapshot{
		Text:       e.textarea.Value(),
		CursorLine: e.textarea.Line(),
		CursorCol:  e.textarea.Column(),
	}
	e.UndoMgr.PushUndo(current)
	e.restoreSnapshot(snap)
}

func (e *Editor) saveSnapshotBefore(prevText string, prevLine, prevCol int, force bool, now time.Time) {
	snap := EditorSnapshot{
		Text:       prevText,
		CursorLine: prevLine,
		CursorCol:  prevCol,
	}

	if force {
		e.UndoMgr.ForceSave(snap, now)
	} else {
		e.UndoMgr.MaybeSave(snap, now)
	}
}

func (e *Editor) restoreSnapshot(snap *EditorSnapshot) {
	e.textarea.SetValue(snap.Text)
	e.textarea.MoveToBegin()

	for range snap.CursorLine {
		e.textarea.CursorDown()
	}

	e.textarea.SetCursorColumn(snap.CursorCol)
	e.ClearSelection()
}
