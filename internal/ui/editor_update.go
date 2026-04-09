package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/cockroachdb/errors"
	"github.com/mattn/go-runewidth"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/utils"
)

// LoadNote はノートをエディタに読み込む。
func (e *Editor) LoadNote(n note.Note) {
	e.noteID = n.ID
	e.original = n.Body
	e.textarea.SetValue(n.Body)
	e.ClearSelection()
	e.UndoMgr.Clear()
}

// SetValue はテキストエリアの値を設定する。
func (e *Editor) SetValue(s string) { e.textarea.SetValue(s) }

// MarkClean は現在の値を基準値として記録する。
func (e *Editor) MarkClean() { e.original = e.textarea.Value() }

// Focus はエディタにフォーカスを当てる。
func (e *Editor) Focus() tea.Cmd { return e.textarea.Focus() }

// Blur はエディタのフォーカスを外す。
func (e *Editor) Blur() { e.textarea.Blur() }

// SetSize はサイズを更新する。
func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height
	e.textarea.SetWidth(width - editorPadding)
	e.textarea.SetHeight(height)
}

// SetReadOnly は読み取り専用モードを設定する。
func (e *Editor) SetReadOnly(v bool) { e.readOnly = v }

// Clear はエディタをクリアする。
func (e *Editor) Clear() {
	e.noteID = ""
	e.original = ""
	e.textarea.SetValue("")
	e.ClearSelection()
}

// --- イベントハンドラ ---

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

	cmd := e.textarea.Update(msg)

	newText := e.textarea.Value()
	if newText != prevText {
		e.saveSnapshotBefore(prevText, prevLine, prevCol, false, now)
	}

	return *e, cmd
}

func (e *Editor) handleKey(msg tea.KeyPressMsg, now time.Time) (Editor, tea.Cmd) { //nolint:cyclop // キーバインド分岐
	if handled, cmd := e.handleCtrlKey(msg, now); handled {
		return *e, cmd
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

	isBackspace := msg.Code == tea.KeyBackspace || (msg.Code == 'h' && msg.Mod == tea.ModCtrl)

	if e.HasSelection() {
		switch {
		case isArrow:
			e.ClearSelection()
		case isBackspace || msg.Code == tea.KeyDelete:
			e.saveSnapshotBefore(prevText, prevLine, prevCol, true, now)
			e.DeleteSelection()

			return *e, nil
		case msg.Text != "":
			e.DeleteSelection()
		}
	}

	forceSnapshot := msg.Code == tea.KeyEnter || isBackspace || msg.Code == tea.KeyDelete

	cmd := e.textarea.Update(msg)

	newText := e.textarea.Value()
	if newText != prevText {
		e.saveSnapshotBefore(prevText, prevLine, prevCol, forceSnapshot, now)
	}

	return *e, cmd
}

func (e *Editor) handleCtrlKey(msg tea.KeyPressMsg, now time.Time) (bool, tea.Cmd) { //nolint:cyclop // キーバインド分岐
	switch {
	case msg.Code == tea.KeyTab:
		return true, editorCmd(EditorBlur)
	case msg.Code == tea.KeyEscape:
		if e.HasSelection() {
			e.ClearSelection()

			return true, nil
		}

		return true, editorCmd(EditorBlur)
	case msg.Code == 's' && msg.Mod == tea.ModCtrl:
		return true, editorCmd(EditorSave)
	case msg.Code == 'z' && msg.Mod == tea.ModCtrl:
		e.Undo()
	case msg.Code == 'z' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		e.Redo()
	case msg.Code == 'a' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		e.SelectAll()
	case msg.Code == 'c' && msg.Mod == tea.ModCtrl:
		_ = e.CopySelection()
	case msg.Code == 'x' && msg.Mod == tea.ModCtrl:
		if e.HasSelection() {
			prevText := e.textarea.Value()
			prevLine := e.textarea.Line()
			prevCol := e.textarea.Column()
			e.saveSnapshotBefore(prevText, prevLine, prevCol, true, now)
			_ = e.CutSelection()
		}
	case msg.Code == 'v' && msg.Mod == tea.ModCtrl:
		prevText := e.textarea.Value()
		prevLine := e.textarea.Line()
		prevCol := e.textarea.Column()
		e.saveSnapshotBefore(prevText, prevLine, prevCol, true, now)
		_ = e.PasteFromClipboard()
	default:
		return false, nil
	}

	return true, nil
}

func editorCmd(msg EditorMsg) tea.Cmd {
	return func() tea.Msg { return msg }
}

func (e *Editor) handleShiftArrow(msg tea.KeyPressMsg) tea.Cmd {
	if !e.HasSelection() {
		anchor := SelectionAnchor{Line: e.textarea.Line(), Column: e.textarea.Column()}
		e.selStart = &anchor
	}

	plainMsg := tea.KeyPressMsg{Code: msg.Code, Mod: 0}
	cmd := e.textarea.Update(plainMsg)

	newPos := SelectionAnchor{Line: e.textarea.Line(), Column: e.textarea.Column()}
	e.selEnd = &newPos

	return cmd
}

// --- 選択 ---

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

// PasteFromClipboard はクリップボードの内容をカーソル位置に挿入する。
// 選択範囲がある場合は選択テキストを置換する。
func (e *Editor) PasteFromClipboard() error {
	text, err := clipboard.ReadAll()
	if err != nil {
		return errors.WithStack(err)
	}

	if text == "" {
		return nil
	}

	if e.HasSelection() {
		e.DeleteSelection()
	}

	e.textarea.InsertText(text)

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
	cellCol := max(x-1, 0) // padding分を差し引き
	visualRow := y + e.textarea.ScrollYOffset()

	totalVisual := e.textarea.layout.totalVisualLines()
	if totalVisual == 0 {
		return SelectionAnchor{Line: 0, Column: 0}
	}

	if visualRow >= totalVisual {
		visualRow = totalVisual - 1
	}

	logLine, runeCol := e.textarea.layout.viewCellToLogical(visualRow, cellCol)

	return SelectionAnchor{Line: logLine, Column: runeCol}
}

func (e *Editor) moveCursorTo(pos SelectionAnchor) {
	e.textarea.MoveToBegin()

	for range pos.Line {
		e.textarea.CursorDown()
	}

	e.textarea.SetCursorColumn(pos.Column)
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

// ScrollUp は表示を n 行上にスクロールする。カーソルは動かさない。
func (e *Editor) ScrollUp(n int) {
	e.textarea.ScrollUp(n)
}

// ScrollDown は表示を n 行下にスクロールする。カーソルは動かさない。
func (e *Editor) ScrollDown(n int) {
	e.textarea.ScrollDown(n)
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
