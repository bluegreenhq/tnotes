package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bluegreenhq/dogubako/tui"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/note"
)

// Model はUIの状態を表す。
type Model struct {
	App              *app.App
	NoteList         NoteList
	Editor           Editor
	Footer           Footer
	Focus            FocusArea
	FolderList       FolderList
	layout           Layout
	dragTarget       dragTarget
	hoverFolderSep   bool
	hoverSeparator   bool
	lastHovered      hoverClearer
	errMsg           string
	infoMsg          string
	infoMsgID        int
	indexModTime     time.Time
	overlay          overlayComponent // オーバーレイ（ヘルプ / ポップアップメニュー等、nil = 非表示）
	lastPopupAnchor  *menuAnchor      // 直前のアンカー付きポップアップの位置（サブメニュー復元用）
	searchDebounceID int              // デバウンスタイマーの世代ID
}

var _ tea.Model = (*Model)(nil)

// InitialModel は初期状態の Model を生成する。
func InitialModel(a *app.App, noWrap bool) *Model {
	m := &Model{
		App:              a,
		NoteList:         NewNoteList(a, a.ListByFolder(app.DefaultFolder), defaultNoteListW, defaultHeight),
		Editor:           NewEditor(minWidth-defaultNoteListW, defaultHeight, noWrap),
		Footer:           NewFooter(),
		Focus:            FocusNoteList,
		FolderList:       NewFolderList(a, defaultFolderListW, defaultHeight),
		layout:           NewLayout(defaultFolderListW, defaultNoteListW),
		dragTarget:       dragNone,
		hoverFolderSep:   false,
		hoverSeparator:   false,
		lastHovered:      nil,
		errMsg:           "",
		infoMsg:          "",
		infoMsgID:        0,
		indexModTime:     time.Time{},
		overlay:          nil,
		lastPopupAnchor:  nil,
		searchDebounceID: 0,
	}

	m.Editor.SetApp(a)
	m.Editor.layout = &m.layout
	m.NoteList.layout = &m.layout
	m.FolderList.layout = &m.layout

	return m
}

// Init は初回のコマンドを返す。
func (m *Model) Init() tea.Cmd {
	mt, err := m.App.IndexModTime()
	if err != nil {
		m.errMsg = err.Error()
	} else {
		m.indexModTime = mt
	}

	if len(m.App.ListByFolder(app.DefaultFolder)) > 0 {
		m.loadSelectedNote()
	}

	return nil
}

// NoteListWidth は現在のノート一覧幅を返す。
func (m *Model) NoteListWidth() int { return m.layout.noteListWidth }

// HelpVisible はヘルプオーバーレイが表示中かを返す。
func (m *Model) HelpVisible() bool {
	_, ok := m.overlay.(*HelpOverlay)

	return ok
}

// OverlayPopupKind は表示中の AnchoredPopupOverlay / FixedPopupOverlay の Kind を返す。
// オーバーレイがポップアップでなければ PopupKindNone を返す。テスト用ヘルパー。
func (m *Model) OverlayPopupKind() PopupKind {
	switch ov := m.overlay.(type) {
	case *AnchoredPopupOverlay:
		return ov.Kind()
	case *FixedPopupOverlay:
		return ov.Kind()
	}

	return PopupKindNone
}

// OverlayMenu は表示中の AnchoredPopupOverlay / FixedPopupOverlay の PopupMenu を返す。
// オーバーレイがポップアップでなければ nil を返す。テスト用ヘルパー。
func (m *Model) OverlayMenu() *tui.PopupMenu {
	switch ov := m.overlay.(type) {
	case *AnchoredPopupOverlay:
		return ov.Menu()
	case *FixedPopupOverlay:
		return ov.Menu()
	}

	return nil
}

// Update はメッセージに応じて状態を更新する。
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:cyclop // type switch dispatch
	now := time.Now()

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.handleResize(msg, now)
	case tea.KeyPressMsg:
		cmd = m.handleKey(msg, now)
	case tea.MouseClickMsg:
		cmd = m.handleClick(msg, now)
	case tea.MouseMotionMsg:
		cmd = m.handleDrag(msg, now)
	case tea.MouseReleaseMsg:
		cmd = m.handleRelease(msg, now)
	case tea.MouseWheelMsg:
		cmd = m.routeMouse(msg, now)
	case tea.MouseMsg:
		cmd = m.handleHover(msg, now)
	case tea.FocusMsg:
		m.handleFocusRestore()
	case tea.BlurMsg: // 他アプリへ切り替え時に編集中の内容を保存
		m.syncEditorToNote(now)
	case tui.CursorBlinkMsg:
		cmd = m.dispatchBlink(msg)
	case searchDebounceMsg:
		if msg.id != m.searchDebounceID {
			return m, nil // 古いタイマーは無視
		}

		m.applySearchFilter(msg.query, now)
	case searchClearedMsg:
		m.applySearchFilter("", now)
	case clearInfoMsg:
		m.handleClearInfo(msg)
	default:
		cmd = m.routeFocus(msg, now)
	}

	m.syncViewState()

	return m, cmd
}

// View はターミナルに描画する内容を返す。
func (m *Model) View() tea.View {
	now := time.Now()
	v := tea.NewView(m.renderView(now))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	v.ReportFocus = true

	return v
}

// dispatchBlink は CursorBlinkMsg を全 BlinkHandler に配信する。
func (m *Model) dispatchBlink(msg tui.CursorBlinkMsg) tea.Cmd {
	targets := m.blinkTargets()
	cmds := make([]tea.Cmd, 0, len(targets))

	for _, h := range targets {
		cmds = append(cmds, h.HandleBlinkMsg(msg))
	}

	return tea.Batch(cmds...)
}

// updateIndexModTime は indexModTime を現在の index.json の modtime で更新する。
// 自身の保存操作による modtime 変更を外部変更と誤検知しないようにする。
func (m *Model) updateIndexModTime() {
	mt, err := m.App.IndexModTime()
	if err == nil {
		m.indexModTime = mt
	}
}

func (m *Model) rebuildFooterButtons() {
	m.Footer.RebuildButtons()
}

func (m *Model) openHelp() {
	h := NewHelpOverlay(m.Focus)
	h.SetScreenSize(m.layout.width, m.layout.BodyHeight())
	m.overlay = h
}

// openAnchoredPopup はアンカー付きポップアップメニューをオーバーレイとして開く。
func (m *Model) openAnchoredPopup(menu *tui.PopupMenu, anchorX, anchorY int, kind PopupKind) {
	overlay := NewAnchoredPopupOverlay(menu, anchorX, anchorY, kind)
	overlay.SetScreenSize(m.layout.width, m.layout.BodyHeight())
	m.overlay = overlay
}

// openConfirmDeleteFolder はフォルダ削除確認ダイアログをオーバーレイとして開く。
func (m *Model) openConfirmDeleteFolder(name string, noteCount int) {
	overlay := NewConfirmDeleteFolderDialog(name, noteCount)
	overlay.SetScreenSize(m.layout.width, m.layout.BodyHeight())
	m.overlay = overlay
}

// openFixedPopup は固定位置ポップアップメニューをオーバーレイとして開く。
func (m *Model) openFixedPopup(menu *tui.PopupMenu, origin func() (int, int), kind PopupKind) {
	overlay := NewFixedPopupOverlay(menu, origin, kind)
	overlay.SetScreenSize(m.layout.width, m.layout.BodyHeight())
	m.overlay = overlay
}

// dismissOverlay は現在表示中のオーバーレイを閉じる（コンポーネント側のフラグも整える）。
func (m *Model) dismissOverlay() {
	switch ov := m.overlay.(type) {
	case *AnchoredPopupOverlay:
		m.closePopupOverlay(ov.Kind())
	case *FixedPopupOverlay:
		m.closePopupOverlay(ov.Kind())
	default:
		m.overlay = nil
	}

	m.lastPopupAnchor = nil
}

// editorHeaderMenuOrigin は EditorHeader の「…」メニュー左上座標を返す（FixedPopupOverlay 用）。
func (m *Model) editorHeaderMenuOrigin() (int, int) {
	return m.layout.EditorStartX() + m.Editor.Header.MenuLeftX(), editorHeaderMenuTopY
}

// editorHeaderMoveMenuOrigin は移動先メニュー左上座標を返す（FixedPopupOverlay 用）。
func (m *Model) editorHeaderMoveMenuOrigin() (int, int) {
	return m.layout.EditorStartX() + m.Editor.Header.MoveMenuLeftX(), editorHeaderMenuTopY
}

// folderListMenuOrigin は FolderList の moreメニュー左上座標を返す（FixedPopupOverlay 用）。
func (m *Model) folderListMenuOrigin() (int, int) {
	return m.FolderList.MenuLeftX(), folderListHeaderLines
}

// footerMenuOrigin は Footer メニュー左上座標を返す（FixedPopupOverlay 用）。
func (m *Model) footerMenuOrigin() (int, int) {
	return 1, m.layout.BodyHeight() - m.Footer.MenuHeight()
}

// toggleFooterMenu はフッターメニュー overlay の開閉をトグルする。
func (m *Model) toggleFooterMenu() {
	if _, ok := m.overlay.(*FixedPopupOverlay); ok {
		if fp, _ := m.overlay.(*FixedPopupOverlay); fp.Kind() == PopupKindFooter {
			m.closePopupOverlay(PopupKindFooter)

			return
		}
	}

	m.Footer.OpenMenu()
	m.openFixedPopup(m.Footer.PopupMenu, m.footerMenuOrigin, PopupKindFooter)
}

// syncViewState は View() に必要な派生状態を同期する。
// Update の最後に呼ばれ、View() の純粋性を保証する。
func (m *Model) syncViewState() {
	m.FolderList.UpdateCounts()

	if m.Editor.Dirty() {
		m.NoteList.SetDirtyNoteID(m.Editor.NoteID())
	} else {
		m.NoteList.SetDirtyNoteID("")
	}

	m.rebuildFooterButtons()
}

func (m *Model) handleKey(msg tea.KeyPressMsg, now time.Time) tea.Cmd {
	// モーダル状態の処理（優先度順）
	if cmd, handled := m.handleModalKey(msg, now); handled {
		return cmd
	}

	m.errMsg = ""
	m.infoMsg = ""
	m.Footer.CloseMenu()
	m.Editor.Header.CloseMenu()
	m.Editor.Header.CloseMoveMenu()

	// グローバルキー
	if cmd, handled := m.handleGlobalKey(msg, now); handled {
		return cmd
	}

	// フォーカス先に委譲
	cmd := m.routeFocus(msg, now)

	return tea.Batch(cmd, m.resetFocusBlink())
}

func (m *Model) handleModalKey(msg tea.KeyPressMsg, now time.Time) (tea.Cmd, bool) {
	// オーバーレイ表示中（ヘルプ / ポップアップメニュー / 確認ダイアログ等）
	if m.overlay != nil {
		action, cmd := m.overlay.UpdateOverlay(msg)

		return m.applyPaneResult(action, cmd, now), true
	}

	return nil, false
}

// closePopupOverlay は overlay をクリアし、対応するコンポーネント側の状態フラグも整える。
// AnchoredPopupOverlay の場合はサブメニュー復元用にアンカーを保存する。
func (m *Model) closePopupOverlay(kind PopupKind) {
	if anchored, ok := m.overlay.(*AnchoredPopupOverlay); ok {
		a := menuAnchor{x: anchored.AnchorX(), y: anchored.AnchorY()}
		m.lastPopupAnchor = &a
	}

	m.overlay = nil

	switch kind {
	case PopupKindEditorHeader:
		m.Editor.Header.CloseMenu()
	case PopupKindMoveMenu:
		m.Editor.Header.CloseMoveMenu()
	case PopupKindFolderList:
		m.FolderList.CloseMenu()
	case PopupKindFooter:
		m.Footer.CloseMenu()
	case PopupKindEditorContext, PopupKindNone:
		// noop（EditorContext はメニューが overlay 内に閉じているため後始末不要）
	}
}

// executePopupAction はポップアップメニュー選択結果を対応するハンドラに振り分ける。
func (m *Model) executePopupAction(kind PopupKind, idx int, now time.Time) tea.Cmd {
	switch kind {
	case PopupKindEditorContext:
		m.Editor.ExecuteContextMenuAction(idx)

		return nil
	case PopupKindEditorHeader:
		return m.applyAction(m.Editor.Header.ExecuteMenuAction(idx), now)
	case PopupKindMoveMenu:
		return m.applyAction(m.Editor.Header.ExecuteMoveMenuAction(idx), now)
	case PopupKindFolderList:
		return m.handleFolderMenuAction(idx, now)
	case PopupKindFooter:
		return m.applyAction(m.Footer.ExecuteMenuAction(idx), now)
	case PopupKindNone:
	}

	return nil
}

func (m *Model) handleGlobalKey(msg tea.KeyPressMsg, now time.Time) (tea.Cmd, bool) {
	switch {
	case msg.Code == 'q' && msg.Mod&tea.ModCtrl != 0:
		m.syncEditorToNote(now)

		return tea.Quit, true
	case msg.Code == 'f' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		m.Editor.Header.SetSearchFocused(true)
		m.Focus = FocusEditor

		return m.Editor.Header.searchBlink.Reset(), true
	case msg.Code == '/' && msg.Mod == (tea.ModCtrl|tea.ModShift):
		m.openHelp()

		return nil, true
	}

	return nil, false
}

// routeFocus はフォーカス先のコンポーネントに msg を委譲する。
func (m *Model) routeFocus(msg tea.Msg, now time.Time) tea.Cmd {
	action, cmd := m.activePane().UpdatePane(msg, m.paneContext(now))

	return m.applyPaneResult(action, cmd, now)
}

// activePane は現在フォーカスのある PaneComponent を返す。
//
//nolint:ireturn // フォーカス遷移のためインターフェースを返す
func (m *Model) activePane() PaneComponent {
	switch m.Focus {
	case FocusFolderList:
		return &m.FolderList
	case FocusNoteList:
		return &m.NoteList
	case FocusEditor:
		return &m.Editor
	}

	return &m.NoteList
}

// paneAt は X 座標から対応する PaneComponent を返す。
//
//nolint:ireturn // 座標分岐のためインターフェースを返す
func (m *Model) paneAt(x int) PaneComponent {
	noteListStart := m.layout.NoteListOffset()
	noteListEnd := m.layout.EditorStartX()

	switch {
	case m.layout.folderVisible && x < noteListStart:
		return &m.FolderList
	case x < noteListEnd:
		return &m.NoteList
	default:
		return &m.Editor
	}
}

// paneContext は PaneComponent に渡す共通コンテキストを生成する。
func (m *Model) paneContext(now time.Time) PaneContext {
	return PaneContext{Now: now, TrashMode: m.Editor.Header.TrashMode()}
}

// blinkTargets は CursorBlinkMsg を配信する BlinkHandler のリストを返す。
func (m *Model) blinkTargets() []BlinkHandler {
	return []BlinkHandler{&m.Editor, &m.FolderList, m.Editor.Header}
}

// resetFocusBlink はキー入力後にフォーカス先の blink をリセットする。
func (m *Model) resetFocusBlink() tea.Cmd {
	switch m.Focus { //nolint:exhaustive // NoteList にはカーソル blink がない
	case FocusFolderList:
		if m.FolderList.InputMode() || m.FolderList.RenameMode() {
			return m.FolderList.blink.Reset()
		}
	case FocusEditor:
		if m.Editor.Header.SearchFocused() {
			return m.Editor.Header.ResetSearchBlink()
		}

		return m.Editor.resetBlink()
	}

	return nil
}

func (m *Model) handleResize(msg tea.WindowSizeMsg, now time.Time) {
	m.layout.width = msg.Width
	m.layout.height = msg.Height
	m.layout.folderVisible = m.FolderList.Visible()
	m.layout.noteListWidth = max(m.layout.noteListWidth, minNoteListWidth)
	m.layout.noteListWidth = min(m.layout.noteListWidth, m.layout.MaxNoteListWidth())

	if m.layout.folderVisible {
		maxNoteWidth := m.layout.width - m.layout.folderListWidth - minEditorWidth
		m.layout.noteListWidth = min(m.layout.noteListWidth, maxNoteWidth)
		m.layout.noteListWidth = max(m.layout.noteListWidth, minNoteListWidth)

		maxFolderWidth := m.layout.width - m.layout.noteListWidth - minEditorWidth
		m.layout.folderListWidth = max(m.layout.folderListWidth, minFolderListWidth)
		m.layout.folderListWidth = min(m.layout.folderListWidth, maxFolderWidth)
	}

	m.recalcLayout(now)
	m.Footer.CloseMenu()

	if m.overlay != nil {
		m.overlay.SetScreenSize(m.layout.width, m.layout.BodyHeight())
	}
}

func (m *Model) handleFocusRestore() {
	changed, mt, err := m.App.RefreshNotes(m.indexModTime)
	if err != nil {
		m.errMsg = err.Error()

		return
	}

	if !changed {
		return
	}

	m.indexModTime = mt

	kind, name := m.FolderList.SelectedKind(), m.FolderList.SelectedName()
	m.NoteList.RefreshKeepSelection(kind, name, m.Editor.NoteID(), time.Now())
	m.loadSelectedNote()
}

func (m *Model) handleClearInfo(msg clearInfoMsg) {
	if msg.id == m.infoMsgID {
		m.infoMsg = ""
	}
}

func (m *Model) handleClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	m.errMsg = ""

	if msg.Button == tea.MouseRight {
		// 既存のオーバーレイを閉じてから右クリック処理
		m.dismissOverlay()

		return m.routeMouse(msg, now)
	}

	if msg.Button != tea.MouseLeft {
		return nil
	}

	// インライン入力中はクリックで確定
	commitCmd := m.commitFolderLineInput(now)
	clickCmd := m.handleClickInner(msg, now)

	return tea.Batch(commitCmd, clickCmd)
}

func (m *Model) commitFolderLineInput(now time.Time) tea.Cmd {
	var action ModelAction

	switch {
	case m.FolderList.InputMode():
		action = m.FolderList.CommitInput()
	case m.FolderList.RenameMode():
		action = m.FolderList.CommitRename()
	}

	return m.applyAction(action, now)
}

func (m *Model) handleClickInner(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	// モーダル状態の処理（優先度順）
	if cmd, handled := m.handleModalClick(msg, now); handled {
		return cmd
	}

	// 検索フォーカス中にエディタヘッダー以外をクリックしたらフォーカス解除
	if m.Editor.Header.SearchFocused() {
		isEditorHeader := msg.Y == 0 && msg.X >= m.layout.EditorStartX()

		if !isEditorHeader {
			m.Editor.Header.SetSearchFocused(false)
		}
	}

	return m.handleZoneClick(msg, now)
}

func (m *Model) handleModalClick(msg tea.MouseClickMsg, now time.Time) (tea.Cmd, bool) {
	// オーバーレイ表示中（ヘルプ / ポップアップメニュー / 確認ダイアログ全般）
	if m.overlay != nil {
		action, cmd := m.overlay.UpdateOverlay(msg)

		return m.applyPaneResult(action, cmd, now), true
	}

	return nil, false
}

func (m *Model) handleZoneClick(msg tea.MouseClickMsg, now time.Time) tea.Cmd {
	zone := m.layout.HitTest(msg.X, msg.Y)
	m.dragTarget = m.zoneToDragTarget(zone)

	switch zone { //nolint:exhaustive // コンポーネントゾーンは routeMouse で処理
	case ZoneFooterLabel:
		return m.handleFooterClick(msg.X, now)
	case ZoneFooterBorder:
		return nil // フッターの罫線行
	case ZoneFolderSeparator:
		return nil
	case ZoneNoteSeparator:
		return nil
	case ZoneFolderList:
		m.Focus = FocusFolderList

		return m.routeMouse(msg, now)
	default:
		return m.routeMouse(msg, now)
	}
}

// zoneToDragTarget は HitZone から dragTarget を算出する。
func (m *Model) zoneToDragTarget(zone HitZone) dragTarget {
	switch zone { //nolint:exhaustive // フッター等は dragTarget に対応しない
	case ZoneFolderSeparator:
		return dragFolderSeparator
	case ZoneNoteSeparator:
		return dragNoteSeparator
	case ZoneFolderList:
		return dragFolderList
	case ZoneNoteList:
		return dragNoteList
	case ZoneEditorHeader, ZoneEditorBody:
		return dragEditor
	default:
		return dragNone
	}
}

func (m *Model) handleFooterClick(x int, now time.Time) tea.Cmd {
	m.rebuildFooterButtons()

	return m.applyAction(m.Footer.HandleClick(x), now)
}

// routeDragTarget はキャプチャ中の dragTarget に応じたコンポーネントに msg を委譲する。
func (m *Model) routeDragTarget(msg mouseMsg, now time.Time) tea.Cmd {
	pane := m.dragTargetPane()
	if pane == nil {
		return nil
	}

	action, cmd := pane.UpdatePane(msg, m.paneContext(now))

	return m.applyPaneResult(action, cmd, now)
}

// dragTargetPane は現在のドラッグ対象 pane を返す（セパレーター・None の場合は nil）。
//
//nolint:ireturn // dragTarget 種別ごとにインターフェースを返す
func (m *Model) dragTargetPane() PaneComponent {
	switch m.dragTarget { //nolint:exhaustive // セパレーター・None は呼び出し元で処理済み
	case dragFolderList:
		return &m.FolderList
	case dragNoteList:
		return &m.NoteList
	case dragEditor:
		return &m.Editor
	}

	return nil
}

// calcHoverTarget はマウス座標から hover 対象コンポーネントを算出する。
func (m *Model) calcHoverTarget(x, y int) hoverClearer { //nolint:ireturn // 複数の具象型を返すため
	if y >= m.layout.FooterLabelY() {
		return &m.Footer
	}

	return m.paneAt(x)
}

// updateHoverTarget は hover 対象が変わった場合に旧対象の hover をクリアする。
func (m *Model) updateHoverTarget(target hoverClearer) {
	if m.lastHovered == target {
		return
	}

	if m.lastHovered != nil {
		m.lastHovered.ClearHover()
	}

	m.lastHovered = target
}

// routeMouse はマウス位置に応じたコンポーネントに msg を委譲する。
func (m *Model) routeMouse(msg mouseMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()

	// フッター領域
	if mouse.Y >= m.layout.FooterLabelY() {
		m.rebuildFooterButtons()
		m.Footer.SetHover(m.Footer.HitTest(mouse.X))

		return nil
	}

	action, cmd := m.paneAt(mouse.X).UpdatePane(msg, m.paneContext(now))

	return m.applyPaneResult(action, cmd, now)
}

func (m *Model) handleDrag(msg tea.MouseMotionMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()

	if m.overlay != nil {
		action, cmd := m.overlay.UpdateOverlay(msg)

		return m.applyPaneResult(action, cmd, now)
	}

	m.hoverSeparator = m.dragTarget == dragNoteSeparator || m.layout.IsOnSeparator(mouse.X)
	m.hoverFolderSep = m.dragTarget == dragFolderSeparator ||
		(m.layout.folderVisible && m.layout.IsOnFolderSeparator(mouse.X))

	switch m.dragTarget {
	case dragFolderSeparator:
		maxFolderWidth := m.layout.width - m.layout.noteListWidth - minEditorWidth
		newWidth := max(mouse.X, minFolderListWidth)
		newWidth = min(newWidth, maxFolderWidth)
		m.layout.folderListWidth = newWidth
		m.recalcLayout(now)

		return nil
	case dragNoteSeparator:
		newWidth := max(mouse.X-m.layout.NoteListOffset(), minNoteListWidth)
		newWidth = min(newWidth, m.layout.MaxNoteListWidth())
		m.layout.noteListWidth = newWidth
		m.recalcLayout(now)

		return nil
	case dragFolderList, dragNoteList, dragEditor:
		return m.routeDragTarget(msg, now)
	case dragNone:
		return m.handleIdleHover(msg, mouse, now)
	}

	return nil
}

func (m *Model) handleRelease(msg tea.MouseReleaseMsg, now time.Time) tea.Cmd {
	defer func() { m.dragTarget = dragNone }()

	switch m.dragTarget {
	case dragFolderSeparator, dragNoteSeparator:
		return nil
	case dragFolderList, dragNoteList, dragEditor:
		return m.routeDragTarget(msg, now)
	case dragNone:
		return nil
	}

	return nil
}

func (m *Model) handleHover(msg tea.MouseMsg, now time.Time) tea.Cmd {
	mouse := msg.Mouse()
	m.hoverSeparator = m.layout.IsOnSeparator(mouse.X)
	m.hoverFolderSep = m.layout.folderVisible && m.layout.IsOnFolderSeparator(mouse.X)

	if m.overlay != nil {
		action, cmd := m.overlay.UpdateOverlay(msg)

		return m.applyPaneResult(action, cmd, now)
	}

	return m.handleIdleHover(msg, mouse, now)
}

// handleIdleHover はドラッグしていない状態のホバー処理。
// handleDrag(dragNone) と handleHover の共通ロジック。
func (m *Model) handleIdleHover(msg mouseMsg, mouse tea.Mouse, now time.Time) tea.Cmd {
	newTarget := m.calcHoverTarget(mouse.X, mouse.Y)
	m.updateHoverTarget(newTarget)

	return m.routeMouse(msg, now)
}

// applyAction は ModelAction を即時適用し、結果の tea.Cmd を返す。nil 安全。
func (m *Model) applyAction(action ModelAction, now time.Time) tea.Cmd {
	if action == nil {
		return nil
	}

	return action(m, ActionContext{Now: now})
}

// applyPaneResult は pane.UpdatePane / 各種 helper の戻り値 (action, cmd) を適用する。
// action があれば即時実行し、追加の cmd があれば一緒に batch する。
func (m *Model) applyPaneResult(action ModelAction, cmd tea.Cmd, now time.Time) tea.Cmd {
	actCmd := m.applyAction(action, now)

	switch {
	case actCmd == nil:
		return cmd
	case cmd == nil:
		return actCmd
	default:
		return tea.Batch(actCmd, cmd)
	}
}

func (m *Model) recalcLayout(now time.Time) {
	m.layout.folderVisible = m.FolderList.Visible()
	bodyHeight := m.layout.BodyHeight()

	if m.layout.folderVisible {
		folderW := max(m.layout.folderListWidth, 0)
		m.FolderList.SetSize(folderW, bodyHeight)
	}

	noteW := max(m.layout.noteListWidth, 0)
	editorW := m.layout.EditorWidth()

	m.NoteList.SetSize(noteW, bodyHeight, now)
	m.Editor.SetSize(editorW, bodyHeight)
}

func (m *Model) scheduleSearchDebounce() tea.Cmd {
	m.searchDebounceID++
	id := m.searchDebounceID
	query := m.Editor.Header.SearchQuery()

	return tea.Tick(searchDebounceDuration, func(_ time.Time) tea.Msg {
		return searchDebounceMsg{id: id, query: query}
	})
}

func (m *Model) applySearchFilter(query string, now time.Time) {
	m.NoteList.ApplySearchFilter(m.FolderList.CurrentFolderName(), query, now)
	m.Editor.SetSearchQuery(query)

	// 最初のノートを選択してエディタに読み込む
	if _, ok := m.NoteList.SelectedNote(); ok {
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}
}

func (m *Model) renderView(now time.Time) string {
	if m.layout.width < minWidth {
		return "Terminal too small — please resize to at least 80 columns"
	}

	noteSepActive := m.hoverSeparator || m.dragTarget == dragNoteSeparator
	noteListView := m.NoteList.View(m.Focus == FocusNoteList, noteSepActive, now, m.FolderList.Visible())

	var body string

	if m.FolderList.Visible() {
		folderSepActive := m.hoverFolderSep || m.dragTarget == dragFolderSeparator
		folderView := m.FolderList.View(m.Focus == FocusFolderList, folderSepActive)
		body = lipgloss.JoinHorizontal(lipgloss.Top, folderView, noteListView, m.Editor.View())
	} else {
		body = lipgloss.JoinHorizontal(lipgloss.Top, noteListView, m.Editor.View())
	}

	footer, footerLines := m.Footer.View(m.errMsg, m.infoMsg, m.layout.width)

	// bodyを正確に height-footerLines 行に切り詰め/パディング
	bodyLines := strings.Split(body, "\n")
	targetBodyLines := m.layout.height - footerLines

	targetBodyLines = max(targetBodyLines, 1)
	if len(bodyLines) > targetBodyLines {
		bodyLines = bodyLines[:targetBodyLines]
	}

	for len(bodyLines) < targetBodyLines {
		bodyLines = append(bodyLines, "")
	}

	body = strings.Join(bodyLines, "\n")
	if m.overlay != nil {
		body = m.overlay.RenderOn(body, m.layout.width, m.layout.BodyHeight())
	}

	return body + "\n" + footer
}

func (m *Model) createNote(now time.Time) tea.Cmd {
	m.syncEditorToNote(now)

	folder := ""
	if m.FolderList.Visible() && m.FolderList.SelectedKind() == FolderUser {
		folder = m.FolderList.SelectedName()
	}

	result, err := m.App.CreateNote(now, folder)

	return m.applyNoteAction(result, err, now)
}

// folderView はフォルダ切替時のUI状態をまとめた構造体。
type folderView struct {
	name      string      // 表示名
	notes     []note.Note // 表示するノート一覧
	sectioned bool        // Today/Yesterday 等のセクション分けを行うか
	readOnly  bool        // エディタを読み取り専用にするか
	trash     bool        // ゴミ箱モードか
	selectID  note.NoteID // 指定IDのノートを選択（空なら先頭）
}

// switchFolder は表示フォルダを切り替え、Editor・NoteList の状態を同期する。
func (m *Model) switchFolder(fv folderView, now time.Time) {
	m.Editor.SetReadOnly(fv.readOnly)
	m.Editor.Header.SetTrashMode(fv.trash)

	if fv.trash {
		m.Editor.Blur()

		if m.Focus != FocusFolderList {
			m.Focus = FocusNoteList
		}
	}

	m.NoteList.Reset(fv.name, fv.sectioned, fv.notes, now)

	if fv.selectID != "" {
		for i, n := range fv.notes {
			if n.ID == fv.selectID {
				m.NoteList.SelectIndex(i, now)

				break
			}
		}
	}

	if len(fv.notes) > 0 {
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}
}

func (m *Model) enterTrashMode(now time.Time) tea.Cmd {
	m.syncEditorToNote(now)

	err := m.App.RefreshTrashNotes()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderTrash))
	m.switchFolder(folderView{
		name:      "Trash",
		notes:     m.App.ListTrashNotes(),
		sectioned: false,
		readOnly:  true,
		trash:     true,
		selectID:  "",
	}, now)

	return nil
}

func (m *Model) exitTrashMode(now time.Time) tea.Cmd { //nolint:unparam // 他アクションメソッドとシグネチャを統一
	m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderNotes))
	m.switchFolder(folderView{
		name:      app.DefaultFolder,
		notes:     m.NoteList.CurrentFolderNotes(m.FolderList.SelectedKind(), m.FolderList.SelectedName()),
		sectioned: true,
		readOnly:  false,
		trash:     false,
		selectID:  "",
	}, now)

	return nil
}

func (m *Model) undoRedoNote(now time.Time, undo bool) tea.Cmd {
	if m.Editor.Header.TrashMode() {
		m.exitTrashMode(now)
	}

	var (
		result app.NoteResult
		err    error
	)

	if undo {
		result, err = m.App.UndoNote()
	} else {
		result, err = m.App.RedoNote()
	}

	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	if result.SelectIdx < 0 && result.InfoHint == "" {
		return nil
	}

	return m.applyNoteResult(result, now)
}

func (m *Model) setNotePin(pin bool) tea.Cmd {
	id := m.Editor.NoteID()
	if id == "" {
		return nil
	}

	var err error
	if pin {
		err = m.App.PinNote(id)
	} else {
		err = m.App.UnpinNote(id)
	}

	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	m.Editor.Header.SetPinned(pin)

	kind, name := m.FolderList.SelectedKind(), m.FolderList.SelectedName()
	m.NoteList.RefreshKeepSelection(kind, name, m.Editor.NoteID(), time.Now())

	msg := "Pinned"
	if !pin {
		msg = "Unpinned"
	}

	return m.setInfoMsg(msg)
}

func (m *Model) openMoveMenu() tea.Cmd {
	opened, err := m.Editor.BuildMoveMenu()
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	if !opened {
		return m.setInfoMsg("No folders to move to")
	}

	// 直前が右クリックメニューならアンカーを引き継ぎ、それ以外は移動ボタン直下の固定位置に表示
	if anchor := m.lastPopupAnchor; anchor != nil {
		m.lastPopupAnchor = nil
		m.openAnchoredPopup(m.Editor.Header.MoveMenu, anchor.x, anchor.y, PopupKindMoveMenu)
	} else {
		m.openFixedPopup(m.Editor.Header.MoveMenu, m.editorHeaderMoveMenuOrigin, PopupKindMoveMenu)
	}

	return nil
}

func (m *Model) handleNoteMove(destFolder string, now time.Time) tea.Cmd {
	id := m.Editor.NoteID()
	if id == "" {
		return nil
	}

	wasTrash := m.Editor.Header.TrashMode()

	if !wasTrash {
		m.syncEditorToNote(now)
	}

	err := m.App.MoveNoteToFolder(id, destFolder)
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	// 移動先フォルダに切り替え
	if m.FolderList.Visible() {
		_ = m.FolderList.RefreshFromApp()
		m.FolderList.SelectIndex(m.FolderList.IndexByName(destFolder))
		m.switchFolder(folderView{
			name:      destFolder,
			notes:     m.App.ListByFolder(destFolder),
			sectioned: destFolder == app.DefaultFolder,
			readOnly:  false,
			trash:     false,
			selectID:  id,
		}, now)
	} else {
		m.NoteList.RefreshKeepSelection(m.FolderList.SelectedKind(), m.FolderList.SelectedName(), m.Editor.NoteID(), now)
	}

	return m.setInfoMsg("Moved to " + destFolder)
}

func (m *Model) toggleFolderList(now time.Time) tea.Cmd {
	m.FolderList.ToggleVisible()

	if m.FolderList.Visible() {
		m.Focus = FocusFolderList
		_ = m.FolderList.RefreshFromApp()

		// 現在のTrash表示状態をフォルダ選択に反映
		if m.Editor.Header.TrashMode() {
			m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderTrash))
		} else {
			m.FolderList.SelectIndex(m.FolderList.IndexByKind(FolderNotes))
		}
	} else if m.Focus == FocusFolderList {
		m.Focus = FocusNoteList
	}

	m.recalcLayout(now)

	return nil
}

func (m *Model) handleFolderSelect(now time.Time) tea.Cmd {
	switch m.FolderList.SelectedKind() {
	case FolderNotes:
		m.switchFolder(folderView{
			name:      app.DefaultFolder,
			notes:     m.App.ListByFolder(app.DefaultFolder),
			sectioned: true,
			readOnly:  false,
			trash:     false,
			selectID:  "",
		}, now)
	case FolderTrash:
		return m.enterTrashMode(now)
	case FolderUser:
		name := m.FolderList.SelectedName()
		m.switchFolder(folderView{
			name:      name,
			notes:     m.App.ListByFolder(name),
			sectioned: false,
			readOnly:  false,
			trash:     false,
			selectID:  "",
		}, now)
	}

	return nil
}

func (m *Model) handleFolderMenuAction(idx int, now time.Time) tea.Cmd {
	const (
		menuRename = 0
		menuDelete = 1
	)

	switch idx {
	case menuRename:
		return m.FolderList.StartRename()
	case menuDelete:
		name := m.FolderList.SelectedName()

		return m.applyAction(m.FolderList.TryDeleteFolder(name), now)
	}

	return nil
}

func (m *Model) loadSelectedNote() {
	err := m.Editor.LoadSelected(&m.NoteList)
	if err != nil {
		m.errMsg = err.Error()
	}
}

func (m *Model) syncEditorToNote(now time.Time) {
	saved, err := m.Editor.Save(now)
	if err != nil {
		m.errMsg = err.Error()
	}

	if m.App.DiscardIfEmpty(m.Editor.NoteID()) || saved {
		m.NoteList.RefreshKeepSelection(m.FolderList.SelectedKind(), m.FolderList.SelectedName(), m.Editor.NoteID(), now)
		m.updateIndexModTime()
	}
}

// applyNoteAction は NoteResult とエラーを処理し、UI状態に反映する。
func (m *Model) applyNoteAction(result app.NoteResult, err error, now time.Time) tea.Cmd {
	if err != nil {
		m.errMsg = err.Error()

		return nil
	}

	return m.applyNoteResult(result, now)
}

// applyNoteResult は NoteResult をUI状態に反映する。
func (m *Model) applyNoteResult(r app.NoteResult, now time.Time) tea.Cmd {
	notes := m.NoteList.CurrentFolderNotes(m.FolderList.SelectedKind(), m.FolderList.SelectedName())
	selectIdx := m.NoteList.ResolveSelectIdx(r, notes)

	m.NoteList.SetNotes(notes, now)

	if selectIdx >= 0 {
		m.NoteList.SelectIndex(selectIdx, now)
		m.loadSelectedNote()
	} else {
		m.Editor.Clear()
	}

	if r.LoadNote {
		m.Editor.LoadNote(r.Note)
	}

	var cmds []tea.Cmd

	if r.FocusEditor {
		m.Focus = FocusEditor
		cmds = append(cmds, m.Editor.Focus())
	}

	if r.InfoHint != "" {
		cmds = append(cmds, m.setInfoMsg(r.InfoHint))
	}

	return tea.Batch(cmds...)
}

func (m *Model) setInfoMsg(msg string) tea.Cmd {
	m.infoMsgID++
	m.infoMsg = msg

	id := m.infoMsgID

	return tea.Tick(infoMsgDuration, func(_ time.Time) tea.Msg {
		return clearInfoMsg{id: id}
	})
}

func (m *Model) openNoteListMenu(now time.Time) tea.Cmd {
	m.Editor.Header.OpenMenu()
	menuW := m.Editor.Header.PopupMenu.Width()
	x := m.layout.EditorStartX() - menuW
	y := m.NoteList.SelectedY(now)
	m.openAnchoredPopup(m.Editor.Header.PopupMenu, x, y, PopupKindEditorHeader)

	return nil
}

func (m *Model) focusEditor() tea.Cmd {
	if m.FolderList.IsTrash() {
		return nil
	}

	if m.Editor.NoteID() == "" {
		return nil
	}

	m.Focus = FocusEditor

	return m.Editor.Focus()
}

func (m *Model) blurEditor(now time.Time) tea.Cmd {
	m.syncEditorToNote(now)
	m.Focus = FocusNoteList
	m.Editor.Blur()

	return nil
}
