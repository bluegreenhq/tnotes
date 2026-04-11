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
