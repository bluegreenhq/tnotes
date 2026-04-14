package ui

// lineInputResult は lineInput のキー処理結果を表す。
type lineInputResult int

const (
	lineInputNone   lineInputResult = iota // 通常の編集操作
	lineInputSubmit                        // Enter が押された
	lineInputCancel                        // Escape が押された
)

// lineInput は単一行のインライン入力を管理する。
// カーソル移動・kill/yank など simpleTextArea 相当の編集操作をサポートする。
type lineInput struct {
	value   []rune
	cursor  int
	killBuf []rune
}

func newLineInput() lineInput {
	return lineInput{value: nil, cursor: 0, killBuf: nil}
}

// Value は入力中のテキストを返す。
func (li *lineInput) Value() string { return string(li.value) }

// SetValue はテキストを設定し、カーソルを末尾に移動する。
func (li *lineInput) SetValue(s string) {
	li.value = []rune(s)
	li.cursor = len(li.value)
}

// Reset は入力をクリアする。
func (li *lineInput) Reset() {
	li.value = nil
	li.cursor = 0
}

// View はカーソル付きの表示文字列を返す。先頭にスペースを付与する。
// カーソル位置の文字をブロックカーソル（█）で置換して表示する。
// cursorVisible が false の場合はカーソルを非表示にする。
func (li *lineInput) View(cursorVisible bool) string {
	if !cursorVisible {
		return " " + string(li.value)
	}

	before := string(li.value[:li.cursor])
	if li.cursor >= len(li.value) {
		return " " + before + "█"
	}

	after := string(li.value[li.cursor+1:])

	return " " + before + "█" + after
}

// ViewWithWidth は最大表示幅を考慮した表示文字列を返す。
// カーソル位置が常に見えるよう、左側を切り詰める。
func (li *lineInput) ViewWithWidth(maxWidth int, cursorVisible bool) string {
	runes := li.value
	cursor := li.cursor

	if maxWidth <= 0 || len(runes) <= maxWidth {
		return li.viewContent(runes, cursor, cursorVisible)
	}

	// カーソルが見えるようにウィンドウをスライド
	start := 0
	if cursor > maxWidth-1 {
		start = cursor - maxWidth + 1
	}

	end := min(start+maxWidth, len(runes))

	return li.viewContent(runes[start:end], cursor-start, cursorVisible)
}

func (li *lineInput) viewContent(runes []rune, cursorPos int, cursorVisible bool) string {
	if !cursorVisible {
		return string(runes)
	}

	before := string(runes[:cursorPos])
	if cursorPos >= len(runes) {
		return before + "█"
	}

	after := string(runes[cursorPos+1:])

	return before + "█" + after
}
