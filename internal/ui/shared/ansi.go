package shared

import (
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// AnsiReset は lipgloss Render が出力する全リセットシーケンス。
const AnsiReset = "\x1b[m"

// VisibleRuneByteRange は ANSI シーケンスをスキップして n 番目の可視ルーンのバイト範囲を返す。
// 可視ルーン数が n 以下の場合は (-1, -1) を返す。
func VisibleRuneByteRange(s string, n int) (int, int) {
	visible := 0
	i := 0

	for i < len(s) {
		if skip := skipANSI(s, i); skip > 0 {
			i += skip

			continue
		}

		if visible == n {
			_, size := utf8.DecodeRuneInString(s[i:])

			return i, i + size
		}

		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		visible++
	}

	return -1, -1
}

// CountVisibleRunes は ANSI シーケンスを除いた可視ルーン数を返す。
func CountVisibleRunes(s string) int {
	count := 0
	i := 0

	for i < len(s) {
		if skip := skipANSI(s, i); skip > 0 {
			i += skip

			continue
		}

		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		count++
	}

	return count
}

// CollectANSIState は s の先頭から bytePos までに出現する ANSI シーケンスを収集し、
// その位置で有効な ANSI 状態を再現する文字列を返す。
// 全リセット (\x1b[m) が見つかった場合はそれ以前の状態を破棄する。
func CollectANSIState(s string, bytePos int) string {
	var b strings.Builder

	i := 0
	for i < len(s) && i < bytePos {
		skip := skipANSI(s, i)
		if skip > 0 {
			seq := s[i : i+skip]
			if seq == AnsiReset {
				b.Reset()
			} else {
				b.WriteString(seq)
			}

			i += skip

			continue
		}

		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
	}

	return b.String()
}

// ExtractANSIOn は lipgloss Style から開始 ANSI シーケンスを抽出する。
func ExtractANSIOn(s lipgloss.Style) string {
	rendered := s.Render("\x00")
	before, _, _ := strings.Cut(rendered, "\x00")

	return before
}

// ExtractANSIOff は lipgloss Style で設定された属性に対応する
// 個別リセットシーケンスを生成する。全属性リセット (\x1b[m) を避け、
// 周囲のスタイルを維持する。
func ExtractANSIOff(s lipgloss.Style) string {
	var b strings.Builder

	if s.GetBold() {
		b.WriteString("\x1b[22m")
	}

	if s.GetItalic() {
		b.WriteString("\x1b[23m")
	}

	if s.GetUnderline() {
		b.WriteString("\x1b[24m")
	}

	if s.GetReverse() {
		b.WriteString("\x1b[27m")
	}

	if s.GetStrikethrough() {
		b.WriteString("\x1b[29m")
	}

	if _, ok := s.GetForeground().(lipgloss.NoColor); !ok {
		b.WriteString("\x1b[39m")
	}

	if _, ok := s.GetBackground().(lipgloss.NoColor); !ok {
		b.WriteString("\x1b[49m")
	}

	return b.String()
}

// skipANSI は位置 i が ANSI エスケープシーケンスの開始であればそのバイト長を返す。
// ANSI シーケンスでなければ 0 を返す。
func skipANSI(s string, i int) int {
	if i >= len(s) || s[i] != '\x1b' || i+1 >= len(s) {
		return 0
	}

	if s[i+1] == '[' {
		return skipCSI(s, i)
	}

	if s[i+1] == ']' {
		return skipOSC(s, i)
	}

	return 0
}

// skipCSI は CSI シーケンス (ESC [ ... 終端) のバイト長を返す。
func skipCSI(s string, start int) int {
	j := start + len("\x1b[")
	for j < len(s) && (s[j] < 0x40 || s[j] > 0x7E) {
		j++
	}

	if j < len(s) {
		j++ // 終端文字を含める
	}

	return j - start
}

// skipOSC は OSC シーケンス (ESC ] ... ST) のバイト長を返す。
func skipOSC(s string, start int) int {
	j := start + len("\x1b]")
	for j < len(s) {
		if s[j] == '\x1b' && j+1 < len(s) && s[j+1] == '\\' {
			return j + len("\x1b\\") - start
		}

		if s[j] == '\x07' {
			return j + 1 - start
		}

		j++
	}

	return j - start
}
