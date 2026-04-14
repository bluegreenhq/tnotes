package ui

// HitZone はクリック座標がどの領域に属するかを表す。
type HitZone int

const (
	// ZoneFolderList はフォルダ一覧領域。
	ZoneFolderList HitZone = iota
	// ZoneFolderSeparator はフォルダ一覧とノート一覧の間のセパレーター。
	ZoneFolderSeparator
	// ZoneNoteList はノート一覧領域。
	ZoneNoteList
	// ZoneNoteSeparator はノート一覧とエディタの間のセパレーター。
	ZoneNoteSeparator
	// ZoneEditorHeader はエディタヘッダー領域。
	ZoneEditorHeader
	// ZoneEditorBody はエディタ本文領域。
	ZoneEditorBody
	// ZoneFooterLabel はフッターラベル行。
	ZoneFooterLabel
	// ZoneFooterBorder はフッター罫線行。
	ZoneFooterBorder
)

// Layout は画面レイアウトの座標計算を担う。
type Layout struct {
	folderListWidth int
	noteListWidth   int
	width           int
	height          int
	folderVisible   bool
}

// NewLayout は新しい Layout を生成する。
func NewLayout(folderListWidth, noteListWidth int) Layout {
	return Layout{
		folderListWidth: folderListWidth,
		noteListWidth:   noteListWidth,
		width:           0,
		height:          0,
		folderVisible:   false,
	}
}

// NoteListOffset はノート一覧の開始X座標を返す。
func (l *Layout) NoteListOffset() int {
	if l.folderVisible {
		return l.folderListWidth
	}

	return 0
}

// BodyHeight はフッターを除いた本文領域の高さを返す。
func (l *Layout) BodyHeight() int {
	return l.height - footerLineCount
}

// EditorWidth はエディタ領域の幅を返す。
func (l *Layout) EditorWidth() int {
	noteW := max(l.noteListWidth, 0)

	if l.folderVisible {
		folderW := max(l.folderListWidth, 0)

		return max(l.width-noteW-folderW, minEditorWidth)
	}

	return max(l.width-noteW, minEditorWidth)
}

// EditorStartX はエディタ領域の開始X座標を返す。
func (l *Layout) EditorStartX() int {
	return l.NoteListOffset() + l.noteListWidth
}

// EditorLocalX は画面X座標をエディタローカルX座標に変換する。
func (l *Layout) EditorLocalX(screenX int) int {
	return screenX - l.EditorStartX()
}

// IsOnSeparator はX座標がノートリスト・エディタ間のセパレーター上かを返す。
func (l *Layout) IsOnSeparator(x int) bool {
	sepX := l.NoteListOffset() + l.noteListWidth

	return x >= sepX-1 && x <= sepX
}

// IsOnFolderSeparator はX座標がフォルダリスト・ノートリスト間のセパレーター上かを返す。
func (l *Layout) IsOnFolderSeparator(x int) bool {
	return x >= l.folderListWidth-1 && x <= l.folderListWidth
}

// MaxNoteListWidth はノートリストの最大幅を返す。
func (l *Layout) MaxNoteListWidth() int {
	pctLimit := l.width * maxNoteListPct / percentDivisor

	editorLimit := l.width - minEditorWidth
	if l.folderVisible {
		editorLimit -= l.folderListWidth
	}

	return min(pctLimit, editorLimit)
}

// FooterLabelY はフッターラベル行のY座標を返す。
func (l *Layout) FooterLabelY() int {
	return l.height - footerLineCount + 1
}

// HitTest はクリック座標がどの領域に属するかを返す。
func (l *Layout) HitTest(x, y int) HitZone { //nolint:cyclop // zone dispatch
	switch {
	case y == l.FooterLabelY():
		return ZoneFooterLabel
	case y >= l.height-footerLineCount:
		return ZoneFooterBorder
	case l.folderVisible && l.IsOnFolderSeparator(x):
		return ZoneFolderSeparator
	case l.IsOnSeparator(x):
		return ZoneNoteSeparator
	case l.folderVisible && x < l.folderListWidth:
		return ZoneFolderList
	case x < l.NoteListOffset()+l.noteListWidth:
		return ZoneNoteList
	default:
		if y < editorHeaderHeight {
			return ZoneEditorHeader
		}

		return ZoneEditorBody
	}
}

// NoteListLocalX は画面X座標をノートリストローカルX座標に変換する。
func (l *Layout) NoteListLocalX(screenX int) int {
	return screenX - l.NoteListOffset()
}
