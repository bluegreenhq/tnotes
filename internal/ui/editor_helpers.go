package ui

import (
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"github.com/bluegreenhq/tnotes/internal/ui/shared"
)

const editorPadding = 2 // lipgloss Padding(0,1) の左右合計

const multiClickTimeout = 500 * time.Millisecond

const (
	clickDouble    = 2 // ダブルクリック
	clickTriple    = 3 // トリプルクリック
	clickResetOver = 4 // これ以上はリセット
)

// SelectionAnchor はテキスト内の位置を表す。
type SelectionAnchor struct {
	Line   int
	Column int
}

// NewSelectionAnchor は新しい SelectionAnchor を生成する。
func NewSelectionAnchor(line, col int) SelectionAnchor {
	return SelectionAnchor{Line: line, Column: col}
}

// selBefore は a が b より前にあるかを返す。
func selBefore(a, b SelectionAnchor) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}

	return a.Column < b.Column
}

// runeClass はワード選択のための文字種分類を表す。
type runeClass int

const (
	classWord     runeClass = iota // ASCII英数字 + '_'
	classHiragana                  // ひらがな
	classKatakana                  // カタカナ
	classKanji                     // 漢字
	classSpace                     // 空白
	classPunct                     // その他
)

// classifyRune はルーンの文字種を返す。
func classifyRune(r rune) runeClass {
	if unicode.Is(unicode.Hiragana, r) {
		return classHiragana
	}

	if unicode.Is(unicode.Katakana, r) || r == 'ー' {
		return classKatakana
	}

	if unicode.Is(unicode.Han, r) {
		return classKanji
	}

	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
		return classWord
	}

	if unicode.IsSpace(r) {
		return classSpace
	}

	return classPunct
}

var (
	urlPattern   = regexp.MustCompile(`https?://[^\s]+`)
	urlGrayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func styleURLsInLine(runes []rune, logicalText string, locs [][]int, startRuneOff, visLen int) string {
	lineLen := len(runes)

	var result strings.Builder

	pos := 0

	for _, loc := range locs {
		urlStartRune := utf8.RuneCountInString(logicalText[:loc[0]])
		urlEndRune := utf8.RuneCountInString(logicalText[:loc[1]])
		urlText := logicalText[loc[0]:loc[1]]

		colStart := max(urlStartRune-startRuneOff, 0)
		colEnd := min(urlEndRune-startRuneOff, visLen)
		colEnd = min(colEnd, lineLen)

		if colStart >= colEnd {
			continue
		}

		styled := urlGrayStyle.Hyperlink(urlText).Render(string(runes[colStart:colEnd]))
		result.WriteString(string(runes[pos:colStart]))
		result.WriteString(styled)

		pos = colEnd
	}

	if pos > 0 {
		result.WriteString(string(runes[pos:]))

		return result.String()
	}

	return ""
}

type searchMatch struct{ start, end int }

func findSearchMatches(logicalText, lowerQuery string) []searchMatch {
	lowerLogical := strings.ToLower(logicalText)
	logicalRunes := []rune(lowerLogical)
	queryRunes := []rune(lowerQuery)
	queryLen := len(queryRunes)

	if queryLen == 0 || len(logicalRunes) < queryLen {
		return nil
	}

	var matches []searchMatch

	for i := 0; i <= len(logicalRunes)-queryLen; i++ {
		if string(logicalRunes[i:i+queryLen]) == lowerQuery {
			matches = append(matches, searchMatch{start: i, end: i + queryLen})
			i += queryLen - 1
		}
	}

	return matches
}

func highlightSearchInLine(runes []rune, logicalText, lowerQuery string, startRuneOff, visLen int) string {
	matches := findSearchMatches(logicalText, lowerQuery)
	if len(matches) == 0 {
		return ""
	}

	line := string(runes)
	lineLen := len(runes)

	var result strings.Builder

	pos := 0

	for _, m := range matches {
		colStart := max(m.start-startRuneOff, 0)
		colEnd := min(m.end-startRuneOff, visLen)
		colEnd = min(colEnd, lineLen)

		if colStart >= colEnd {
			continue
		}

		byteStart, _ := shared.VisibleRuneByteRange(line, colStart)
		_, byteEnd := shared.VisibleRuneByteRange(line, colEnd-1)

		if byteStart < 0 || byteEnd < 0 || byteStart < pos {
			continue
		}

		result.WriteString(line[pos:byteStart])
		result.WriteString(editorSearchHighlightOn)
		result.WriteString(line[byteStart:byteEnd])
		result.WriteString(editorSearchHighlightOff)

		pos = byteEnd
	}

	if pos > 0 {
		result.WriteString(line[pos:])

		return result.String()
	}

	return ""
}
