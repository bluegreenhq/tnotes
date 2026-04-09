package ui

// lineLayout は論理行と視覚行のマッピングを提供する。
type lineLayout interface { //nolint:interfacebloat // マッピング5+操作6、分割すると実装側が複雑化する
	// rebuild はテキストや幅の変更後にマッピングを再構築する。
	rebuild(lines [][]rune, width int)
	// totalVisualLines は全視覚行数を返す。
	totalVisualLines() int
	// logicalToVisual は論理行・列から視覚行インデックスを返す。
	logicalToVisual(line, col int) int
	// visualToLogical は視覚行から論理行・開始ルーンインデックスを返す。
	visualToLogical(visualRow int) (line, startRune int)
	// visualLinesFor は指定論理行の視覚行スライスを返す。
	visualLinesFor(line int) []visualLine

	// adjustScroll はカーソル位置に基づいてスクロールオフセットを調整する。
	adjustScroll(row, col, scrollY, scrollX, width, height int) (newScrollY, newScrollX int)
	// moveCursorUp はカーソルを1つ上の視覚行に移動した結果を返す。
	moveCursorUp(row, col int) (newRow, newCol int, moved bool)
	// moveCursorDown はカーソルを1つ下の視覚行に移動した結果を返す。
	moveCursorDown(row, col int) (newRow, newCol int, moved bool)
	// renderViewLine は指定視覚行のテキストを返す。
	renderViewLine(visualRow, scrollX, width int) string
	// viewLineStartRune は視覚行の論理行と表示開始ルーンオフセットを返す。
	viewLineStartRune(visualRow, scrollX int) (logicalLine, startRuneOff int)
	// viewCellToLogical はビュー上のセル位置から論理行・ルーン列を返す。
	viewCellToLogical(visualRow, cellCol int) (logicalLine, runeCol int)
}

// visualLine は1つの視覚行を表す。
type visualLine struct {
	logicalLine int // 元の論理行インデックス
	startRune   int // この視覚行が論理行の何文字目から始まるか
	length      int // この視覚行のルーン数
}
