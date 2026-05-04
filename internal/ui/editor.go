package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/bluegreenhq/dogubako/tui"
	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/utils"
)

// Editor はテキスト編集ペインの状態を表す。
type Editor struct {
	Header        *EditorHeader
	app           *app.App
	textarea      simpleTextArea
	noteID        note.NoteID
	original      string
	width         int
	height        int
	layout        *Layout
	readOnly      bool
	selecting     bool // ドラッグ中か
	selStart      *SelectionAnchor
	selEnd        *SelectionAnchor
	UndoMgr       *EditorUndoManager
	blink         tui.CursorBlink
	searchQuery   string // 検索クエリ
	lastClickTime time.Time
	lastClickPos  SelectionAnchor
	clickCount    int
}

// NewEditor は新しい Editor を生成する。
func NewEditor(width, height int, noWrap bool) Editor {
	ta := newSimpleTextArea(noWrap)
	ta.SetWidth(width - editorPadding)
	ta.SetHeight(height - editorHeaderHeight)

	return Editor{
		Header:        NewEditorHeader(width),
		app:           nil,
		textarea:      ta,
		noteID:        "",
		original:      "",
		width:         width,
		height:        height,
		layout:        nil,
		readOnly:      false,
		selecting:     false,
		selStart:      nil,
		selEnd:        nil,
		UndoMgr:       NewEditorUndoManager(),
		blink:         tui.NewCursorBlink(blinkOwnerEditor),
		searchQuery:   "",
		lastClickTime: time.Time{},
		lastClickPos:  NewSelectionAnchor(0, 0),
		clickCount:    0,
	}
}

// NoteID は現在編集中のノートIDを返す。
func (e *Editor) NoteID() note.NoteID { return e.noteID }

// Value はテキストエリアの現在の値を返す。
func (e *Editor) Value() string { return e.textarea.Value() }

// Dirty は未保存の変更があるかを返す。
func (e *Editor) Dirty() bool { return e.textarea.Value() != e.original }

// Focused はフォーカス状態を返す。
func (e *Editor) Focused() bool { return e.textarea.Focused() }

// ReadOnly は読み取り専用モードかを返す。
func (e *Editor) ReadOnly() bool { return e.readOnly }

// HasSelection は選択範囲があるかを返す。
func (e *Editor) HasSelection() bool {
	return e.selStart != nil && e.selEnd != nil && *e.selStart != *e.selEnd
}

// Selecting はドラッグ中かを返す。
func (e *Editor) Selecting() bool { return e.selecting }

// BlinkVisible はカーソルの表示状態を返す。
func (e *Editor) BlinkVisible() bool { return e.blink.Visible() }

// SetSearchQuery は検索クエリを設定する。
func (e *Editor) SetSearchQuery(q string) { e.searchQuery = q }

// HandleBlinkMsg は CursorBlinkMsg を処理して blink 状態を切り替える。
func (e *Editor) HandleBlinkMsg(msg tui.CursorBlinkMsg) tea.Cmd {
	return e.blink.HandleMsg(msg)
}

// LoadNote はノートをエディタに読み込む。
func (e *Editor) LoadNote(n note.Note) {
	e.noteID = n.ID
	e.original = n.Body
	e.textarea.SetValue(n.Body)
	e.textarea.MoveToBegin()
	e.ClearSelection()
	e.UndoMgr.Clear()
	e.Header.SetHasNote(true)
	e.Header.SetPinned(n.Pinned)
	e.Header.CloseMenu()
}

// LoadSelected は NoteList の選択中ノートを App から読み込んでエディタに表示する。
func (e *Editor) LoadSelected(nl *NoteList) error {
	n, ok := nl.SelectedNote()
	if !ok {
		e.Clear()

		return nil
	}

	if e.app != nil {
		loaded, err := e.app.LoadNote(n)
		if err != nil {
			return err
		}

		n = loaded
	}

	e.LoadNote(n)

	return nil
}

// BuildMoveMenu は移動先フォルダの候補を構築してメニューを開く。
// 候補がない場合は false を返す。
func (e *Editor) BuildMoveMenu() (bool, error) {
	if e.noteID == "" || e.app == nil {
		return false, nil
	}

	currentFolder := e.app.FindNoteFolder(e.noteID)

	folders, err := e.app.ListFolders()
	if err != nil {
		return false, err
	}

	candidates := make([]string, 0, len(folders)+1)
	if currentFolder != app.DefaultFolder {
		candidates = append(candidates, app.DefaultFolder)
	}

	for _, f := range folders {
		if f != currentFolder {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return false, nil
	}

	e.Header.OpenMoveMenu(candidates)

	return true, nil
}

// CopyToClipboard はエディタの内容をクリップボードにコピーする。
// 内容が空の場合は何もしない。
func (e *Editor) CopyToClipboard() tea.Cmd {
	content := e.Value()
	if content == "" {
		return nil
	}

	err := clipboard.WriteAll(content)
	if err != nil {
		return actionResultMsg{Err: errors.WithStack(err), Info: ""}.Cmd()
	}

	return actionResultMsg{Err: nil, Info: "Copied"}.Cmd()
}

// SetValue はテキストエリアの値を設定する。
func (e *Editor) SetValue(s string) { e.textarea.SetValue(s) }

// MarkClean は現在の値を基準値として記録する。
func (e *Editor) MarkClean() { e.original = e.textarea.Value() }

// SetApp は App 参照を設定する。
func (e *Editor) SetApp(a *app.App) { e.app = a }

// Save は未保存の変更を永続化する。保存した場合は true を返す。
func (e *Editor) Save(now time.Time) (bool, error) {
	if !e.Dirty() || e.app == nil {
		return false, nil
	}

	_, err := e.app.SaveNote(e.noteID, e.Value(), now)
	if err != nil {
		return false, err
	}

	e.MarkClean()

	return true, nil
}

// Focus はエディタにフォーカスを当て、blink タイマーを開始する。
func (e *Editor) Focus() tea.Cmd {
	focusCmd := e.textarea.Focus()
	blinkCmd := e.blink.Reset()

	return tea.Batch(focusCmd, blinkCmd)
}

// Blur はエディタのフォーカスを外し、blink を停止する。
func (e *Editor) Blur() {
	e.textarea.Blur()
	e.blink.Stop()
}

// SetSize はサイズを更新する。
func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height
	e.Header.SetWidth(width)
	e.textarea.SetWidth(width - editorPadding)
	e.textarea.SetHeight(height - editorHeaderHeight)
}

// SetReadOnly は読み取り専用モードを設定する。
func (e *Editor) SetReadOnly(v bool) { e.readOnly = v }

// Clear はエディタをクリアする。
func (e *Editor) Clear() {
	e.noteID = ""
	e.original = ""
	e.textarea.SetValue("")
	e.ClearSelection()
	e.Header.SetHasNote(false)
	e.Header.CloseMenu()
}

// UpdatePane は PaneComponent インターフェース実装。
func (e *Editor) UpdatePane(msg tea.Msg, ctx PaneContext) tea.Cmd {
	var cmd tea.Cmd

	*e, cmd = e.Update(msg, ctx.Now)

	return cmd
}

// Update はメッセージに応じて状態を更新する。
func (e *Editor) Update(msg tea.Msg, now time.Time) (Editor, tea.Cmd) { //nolint:cyclop,funlen // type switch dispatch
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseRight {
			return e.handleRightClickMsg(msg)
		}

		if e.readOnly {
			return *e, nil
		}

		return e.handleClickMsg(msg, now)
	case tea.MouseMotionMsg:
		if e.selecting {
			localX := e.layout.EditorLocalX(msg.Mouse().X)
			e.UpdateDragSelection(localX, msg.Mouse().Y-editorHeaderHeight)
		} else {
			localX := e.layout.EditorLocalX(msg.Mouse().X)
			e.HandleHover(localX, msg.Mouse().Y)
		}

		return *e, nil
	case tea.MouseReleaseMsg:
		if e.selecting {
			e.StopDragSelection()
		}

		return *e, nil
	case tea.MouseWheelMsg:
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			e.scrollUp(1)
		case tea.MouseWheelDown:
			e.scrollDown(1)
		}

		return *e, nil
	case tea.KeyPressMsg:
		if e.readOnly {
			return *e, nil
		}

		// 検索フィールドにフォーカスがある場合
		if e.Header.SearchFocused() {
			return e.handleSearchKey(msg)
		}

		return e.handleKey(msg, now)
	case tea.MouseMsg:
		localX := e.layout.EditorLocalX(msg.Mouse().X)
		e.HandleHover(localX, msg.Mouse().Y)

		return *e, nil
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

// SelectWord はワード選択を行う。
// line, col は論理行・列（rune 単位）。
func (e *Editor) SelectWord(line, col int) {
	lines := e.textarea.lines
	if line < 0 || line >= len(lines) {
		return
	}

	runes := lines[line]
	if len(runes) == 0 {
		return
	}

	// col が行末の場合は1つ前の文字を基準にする
	idx := col
	if idx >= len(runes) {
		idx = len(runes) - 1
	}

	cls := classifyRune(runes[idx])

	// 左に探索
	left := idx
	for left > 0 && classifyRune(runes[left-1]) == cls {
		left--
	}

	// 右に探索
	right := idx + 1
	for right < len(runes) && classifyRune(runes[right]) == cls {
		right++
	}

	start := NewSelectionAnchor(line, left)
	end := NewSelectionAnchor(line, right)
	e.SetSelection(start, end)
	e.moveCursorTo(end)
}

// SelectLine は論理行全体を選択する。
func (e *Editor) SelectLine(line int) {
	lines := e.textarea.lines
	if line < 0 || line >= len(lines) {
		return
	}

	runes := lines[line]
	if len(runes) == 0 {
		return
	}

	start := NewSelectionAnchor(line, 0)
	end := NewSelectionAnchor(line, len(runes))
	e.SetSelection(start, end)
	e.moveCursorTo(end)
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

	start := NewSelectionAnchor(0, 0)
	end := NewSelectionAnchor(lastLine, len([]rune(lines[lastLine])))
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
	e.textarea.MoveTo(start.Line, start.Column)

	e.ClearSelection()
}

// HandleTextAreaClick はエディタ textarea 領域のクリックを処理する。
// x, y はエディタ左上（ヘッダー除く）を原点とする相対座標。
// クリック回数に応じてシングル→ドラッグ開始、ダブル→ワード選択、トリプル→行選択を行う。
func (e *Editor) HandleTextAreaClick(x, y int, now time.Time) {
	pos := e.positionFromMouse(x, y)

	if now.Sub(e.lastClickTime) < multiClickTimeout && pos == e.lastClickPos {
		e.clickCount++
	} else {
		e.clickCount = 1
	}

	if e.clickCount >= clickResetOver {
		e.clickCount = 1
	}

	e.lastClickTime = now
	e.lastClickPos = pos

	switch e.clickCount {
	case clickDouble:
		e.ClearSelection()
		e.SelectWord(pos.Line, pos.Column)
	case clickTriple:
		e.ClearSelection()
		e.SelectLine(pos.Line)
	default:
		e.ClearSelection()
		e.moveCursorTo(pos)
		e.selStart = &pos
		e.selEnd = &pos
		e.selecting = true
	}
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

// ClearHover はエディタの全 hover 状態をクリアする。
func (e *Editor) ClearHover() {
	e.Header.ClearHover()
}

// HandleHover はエディタ領域のホバーを処理する。
// x, y はエディタ左上を原点とする相対座標。
func (e *Editor) HandleHover(x, y int) {
	// ヘッダーボタンのホバー
	if y == 0 {
		e.Header.SetHover(x)
	} else {
		e.Header.ClearHover()
	}

	// メニューのホバー
	if e.Header.MenuOpen() {
		menuTopY := editorHeaderMenuTopY
		menuHeight := e.Header.MenuHeight()

		if y >= menuTopY && y < menuTopY+menuHeight {
			menuRelX := x - e.Header.MenuLeftX()
			e.Header.SetMenuHover(menuRelX, y-menuTopY)
		}
	}
}

// SaveSnapshot はエディタのスナップショットをデバウンス付きで保存する。
func (e *Editor) SaveSnapshot(now time.Time) {
	e.UndoMgr.MaybeSave(NewEditorSnapshot(e.textarea.Value(), e.textarea.Line(), e.textarea.Column()), now)
}

// Undo はエディタの状態を1つ前に戻す。
func (e *Editor) Undo() {
	snap := e.UndoMgr.PopUndo()
	if snap == nil {
		return
	}

	e.UndoMgr.PushRedo(NewEditorSnapshot(e.textarea.Value(), e.textarea.Line(), e.textarea.Column()))
	e.restoreSnapshot(snap)
}

// Redo はエディタの状態を1つ先に進める。
func (e *Editor) Redo() {
	snap := e.UndoMgr.PopRedo()
	if snap == nil {
		return
	}

	e.UndoMgr.PushUndo(NewEditorSnapshot(e.textarea.Value(), e.textarea.Line(), e.textarea.Column()))
	e.restoreSnapshot(snap)
}

// ExecuteContextMenuAction はインデックスに対応するコンテキストメニューのアクションを実行する。
func (e *Editor) ExecuteContextMenuAction(idx int) {
	switch editorContextMsg(idx) {
	case editorContextCopy:
		_ = e.CopySelection()
	case editorContextCut:
		_ = e.CutSelection()
	case editorContextPaste:
		_ = e.PasteFromClipboard()
	}
}

// View はエディタの描画内容を返す。
func (e *Editor) View() string {
	headerLine := e.Header.View()

	if e.noteID == "" {
		placeholder := "Press 'n' to create a note"
		if e.readOnly {
			placeholder = ""
		}

		body := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Width(e.width).
			Height(e.height-editorHeaderHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render(placeholder)

		return headerLine + "\n" + body
	}

	raw := e.textarea.View()

	raw = e.applyTitleBold(raw)
	raw = e.applySearchHighlight(raw)
	raw = e.applyURLStyle(raw)

	if e.HasSelection() {
		raw = e.applySelectionHighlight(raw)
	} else if e.textarea.Focused() && e.blink.Visible() && !e.Header.SearchFocused() {
		raw = e.applyCursor(raw)
	}

	textBody := editorStyle.Width(e.width).Height(e.height - editorHeaderHeight).Render(raw)

	return headerLine + "\n" + textBody
}

// resetBlink はカーソルを表示状態にリセットし、新しい blink タイマーを開始する。
func (e *Editor) resetBlink() tea.Cmd {
	return e.blink.Reset()
}

func (e *Editor) handleSearchKey(msg tea.KeyPressMsg) (Editor, tea.Cmd) {
	handled, _ := e.Header.HandleSearchKey(msg)
	if !handled {
		return *e, nil
	}

	if !e.Header.SearchFocused() {
		// Esc/Enter で検索フォーカスを外した
		e.Header.searchBlink.Stop()

		return *e, EditorSearchBlur.Cmd()
	}

	return *e, EditorSearchChanged.Cmd()
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
		return true, EditorBlur.Cmd()
	case msg.Code == tea.KeyEscape:
		if e.HasSelection() {
			e.ClearSelection()

			return true, nil
		}

		return true, EditorBlur.Cmd()
	case msg.Code == 's' && msg.Mod == tea.ModCtrl:
		return true, EditorSave.Cmd()
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
	case msg.Code == 'o' && msg.Mod == tea.ModCtrl:
		if u := e.urlAtCursor(); u != "" {
			return true, editorOpenURLMsg{URL: u}.Cmd()
		}

		return true, nil
	default:
		return false, nil
	}

	return true, nil
}

func (e *Editor) handleShiftArrow(msg tea.KeyPressMsg) tea.Cmd {
	if !e.HasSelection() {
		anchor := NewSelectionAnchor(e.textarea.Line(), e.textarea.Column())
		e.selStart = &anchor
	}

	plainMsg := tea.KeyPressMsg{Code: msg.Code, Mod: 0}
	cmd := e.textarea.Update(plainMsg)

	newPos := NewSelectionAnchor(e.textarea.Line(), e.textarea.Column())
	e.selEnd = &newPos

	return cmd
}

func (e *Editor) positionFromMouse(x, y int) SelectionAnchor {
	cellCol := max(x-1, 0) // padding分を差し引き
	visualRow := y + e.textarea.ScrollYOffset()

	logLine, runeCol := e.textarea.positionFromCell(visualRow, cellCol)

	return NewSelectionAnchor(logLine, runeCol)
}

func (e *Editor) moveCursorTo(pos SelectionAnchor) {
	e.textarea.MoveTo(pos.Line, pos.Column)
}

// handleClickMsg は tea.MouseClickMsg を処理する。
// layout を参照して絶対座標をローカル座標に変換し、ヘッダー/本文に振り分ける。
func (e *Editor) handleClickMsg(msg tea.MouseClickMsg, now time.Time) (Editor, tea.Cmd) {
	localX := e.layout.EditorLocalX(msg.X)

	// ヘッダー行のクリック
	if msg.Y == 0 {
		cmd := e.handleClick(localX, 0)

		// 検索フォーカス中はエディタへのフォーカス取得も要求
		if e.Header.SearchFocused() {
			return *e, tea.Batch(cmd, EditorClickBody.Cmd())
		}

		return *e, cmd
	}

	// 本文クリック: readOnly またはノート未選択なら無視
	if e.readOnly || e.noteID == "" {
		return *e, nil
	}

	e.HandleTextAreaClick(localX, msg.Y-editorHeaderHeight, now)

	return *e, EditorClickBody.Cmd()
}

func (e *Editor) handleClick(x, y int) tea.Cmd {
	// メニューが開いている場合
	if e.Header.MenuOpen() {
		menuTopY := editorHeaderMenuTopY
		menuHeight := e.Header.MenuHeight()

		if y >= menuTopY && y < menuTopY+menuHeight {
			menuRelX := x - e.Header.MenuLeftX()

			return e.Header.HandleMenuClick(menuRelX, y-menuTopY)
		}

		e.Header.CloseMenu()

		return nil
	}

	// ヘッダー行
	if y == 0 {
		e.Header.SetHasContent(e.textarea.Value() != "")

		return e.Header.HandleClick(x)
	}

	return nil
}

func (e *Editor) scrollUp(n int) {
	e.textarea.ScrollUp(n)
}

func (e *Editor) scrollDown(n int) {
	e.textarea.ScrollDown(n)
}

func (e *Editor) saveSnapshotBefore(prevText string, prevLine, prevCol int, force bool, now time.Time) {
	snap := NewEditorSnapshot(prevText, prevLine, prevCol)

	if force {
		e.UndoMgr.ForceSave(snap, now)
	} else {
		e.UndoMgr.MaybeSave(snap, now)
	}
}

// urlAtCursor はカーソル位置にある URL を返す。URL 上にない場合は空文字列を返す。
func (e *Editor) urlAtCursor() string {
	line := e.textarea.Line()
	if line >= len(e.textarea.lines) {
		return ""
	}

	logicalText := string(e.textarea.lines[line])
	cursorByte := len(string([]rune(logicalText)[:e.textarea.Column()]))

	for _, loc := range urlPattern.FindAllStringIndex(logicalText, -1) {
		if loc[0] <= cursorByte && cursorByte <= loc[1] {
			return logicalText[loc[0]:loc[1]]
		}
	}

	return ""
}

func (e *Editor) handleRightClickMsg(msg tea.MouseClickMsg) (Editor, tea.Cmd) {
	if e.readOnly {
		return *e, nil
	}

	return *e, EditorRightClickMsg{
		Menu:    e.buildContextMenu(),
		AnchorX: msg.X,
		AnchorY: msg.Y,
	}.Cmd()
}

// buildContextMenu は現在の選択・読み取り専用状態に応じたコンテキストメニューを構築する。
func (e *Editor) buildContextMenu() *tui.PopupMenu {
	hasSel := e.HasSelection()

	newItem := func(label string, disabled bool) tui.MenuItem {
		if disabled {
			return tui.NewDisabledMenuItem(label)
		}

		return tui.NewMenuItem(label)
	}

	return tui.NewPopupMenu([]tui.MenuItem{
		newItem("Copy", !hasSel),
		newItem("Cut", !hasSel || e.readOnly),
		newItem("Paste", e.readOnly),
	})
}

func (e *Editor) restoreSnapshot(snap *EditorSnapshot) {
	e.textarea.SetValue(snap.Text)
	e.textarea.MoveTo(snap.CursorLine, snap.CursorCol)
	e.ClearSelection()
}

// applyTitleBold は先頭のタイトル行を太字にする。
// カーソルや選択のANSIエスケープが含まれていても正しく動作する。
func (e *Editor) applyTitleBold(raw string) string {
	scrollOffset := e.textarea.ScrollYOffset()

	// タイトル行（論理行0）が表示する視覚行数を取得
	titleVisualLines := len(e.textarea.layout.visualLinesFor(0))
	if titleVisualLines == 0 {
		titleVisualLines = 1
	}

	// スクロールでタイトル行が完全に画面外なら何もしない
	if scrollOffset >= titleVisualLines {
		return raw
	}

	viewLines := strings.Split(raw, "\n")
	if len(viewLines) == 0 {
		return raw
	}

	// 画面に表示されているタイトルの視覚行数
	boldCount := min(titleVisualLines-scrollOffset, len(viewLines))

	for i := range boldCount {
		line := viewLines[i]
		// 内部のリセットシーケンス後に太字を再適用する
		line = strings.ReplaceAll(line, ansiReset, ansiReset+editorBoldOn)
		viewLines[i] = editorBoldOn + line + editorBoldOff
	}

	return strings.Join(viewLines, "\n")
}

// applyURLStyle はテキスト中の URL をグレー文字色にし、OSC 8 ハイパーリンクを付与する。
// 論理行の元テキストから URL 位置を検出し、視覚行上の対応範囲にスタイルを適用する。
func (e *Editor) applyURLStyle(raw string) string {
	scrollOffset := e.textarea.ScrollYOffset()
	viewLines := strings.Split(raw, "\n")

	for i, line := range viewLines {
		visualRow := i + scrollOffset
		logLine, startRuneOff := e.textarea.layout.viewLineStartRune(visualRow, e.textarea.scrollX)

		if logLine >= len(e.textarea.lines) {
			continue
		}

		logicalText := string(e.textarea.lines[logLine])
		locs := urlPattern.FindAllStringIndex(logicalText, -1)

		if len(locs) == 0 {
			continue
		}

		visLen := e.textarea.visualLineLength(visualRow)
		styled := styleURLsInLine([]rune(line), logicalText, locs, startRuneOff, visLen)

		if styled != "" {
			viewLines[i] = styled
		}
	}

	return strings.Join(viewLines, "\n")
}

// applyCursor はカーソル位置の文字を反転表示する。
// ANSI エスケープシーケンスをスキップして可視ルーン位置を計算する。
func (e *Editor) applyCursor(raw string) string {
	visualRow := e.textarea.layout.logicalToVisual(e.textarea.Line(), e.textarea.Column())
	cursorViewRow := visualRow - e.textarea.ScrollYOffset()

	_, startRuneOff := e.textarea.layout.viewLineStartRune(visualRow, e.textarea.scrollX)
	cursorCol := e.textarea.Column() - startRuneOff

	viewLines := strings.Split(raw, "\n")
	if cursorViewRow < 0 || cursorViewRow >= len(viewLines) {
		return raw
	}

	line := viewLines[cursorViewRow]

	byteStart, byteEnd := visibleRuneByteRange(line, cursorCol)
	if byteStart < 0 {
		// カーソルが行末の場合
		viewLines[cursorViewRow] = line + editorCursorOn + " " + editorCursorOff
	} else {
		before := line[:byteStart]
		cursor := editorCursorOn + line[byteStart:byteEnd] + editorCursorOff
		after := line[byteEnd:]
		viewLines[cursorViewRow] = before + cursor + after
	}

	return strings.Join(viewLines, "\n")
}

// applySearchHighlight は検索クエリにマッチする箇所をハイライトする。
func (e *Editor) applySearchHighlight(raw string) string {
	if e.searchQuery == "" {
		return raw
	}

	lowerQuery := strings.ToLower(e.searchQuery)
	scrollOffset := e.textarea.ScrollYOffset()
	viewLines := strings.Split(raw, "\n")

	for i, line := range viewLines {
		visualRow := i + scrollOffset
		logLine, startRuneOff := e.textarea.layout.viewLineStartRune(visualRow, e.textarea.scrollX)

		if logLine >= len(e.textarea.lines) {
			continue
		}

		logicalText := string(e.textarea.lines[logLine])
		visLen := e.textarea.visualLineLength(visualRow)

		styled := highlightSearchInLine([]rune(line), logicalText, lowerQuery, startRuneOff, visLen)
		if styled != "" {
			viewLines[i] = styled
		}
	}

	return strings.Join(viewLines, "\n")
}

func (e *Editor) applySelectionHighlight(raw string) string {
	if !e.HasSelection() {
		return raw
	}

	start, end := e.NormalizedSelection()
	scrollOffset := e.textarea.ScrollYOffset()

	viewLines := strings.Split(raw, "\n")

	for i, line := range viewLines {
		visualRow := i + scrollOffset
		logLine, startRuneOff := e.textarea.layout.viewLineStartRune(visualRow, e.textarea.scrollX)

		if logLine < start.Line || logLine > end.Line {
			continue
		}

		visibleCount := countVisibleRunes(line)

		var colStart, colEnd int
		if logLine == start.Line {
			colStart = start.Column - startRuneOff
		}

		if logLine == end.Line {
			colEnd = end.Column - startRuneOff
		} else {
			colEnd = visibleCount
		}

		colStart = utils.ClampInt(colStart, 0, visibleCount)
		colEnd = utils.ClampInt(colEnd, 0, visibleCount)

		if colStart >= colEnd {
			continue
		}

		byteStart, _ := visibleRuneByteRange(line, colStart)
		_, byteEnd := visibleRuneByteRange(line, colEnd-1)

		if byteStart < 0 || byteEnd < 0 {
			continue
		}

		before := line[:byteStart]
		middle := line[byteStart:byteEnd]
		after := line[byteEnd:]

		// 選択範囲内の検索ハイライトを除去して選択スタイルを優先する
		middle = strings.ReplaceAll(middle, editorSearchHighlightOn, "")
		middle = strings.ReplaceAll(middle, editorSearchHighlightOff, "")

		restore := collectANSIState(line, byteEnd)
		viewLines[i] = before + editorSelectionOn + middle + editorSelectionOff + restore + after
	}

	return strings.Join(viewLines, "\n")
}
