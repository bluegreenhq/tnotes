package utils

import (
	"time"

	"github.com/mattn/go-runewidth"
)

// Truncate は文字列を表示幅 maxWidth に切り詰め、超過時は末尾を "…" にする。
// 全角文字の表示幅を正しく考慮する。
func Truncate(s string, maxWidth int) string {
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}

	ellipsis := "…"
	ellipsisW := runewidth.RuneWidth('…')
	w := 0

	runes := []rune(s)

	for i, r := range runes {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxWidth-ellipsisW {
			return string(runes[:i]) + ellipsis
		}

		w += rw
	}

	return s
}

// CellToRuneIndex はセル幅の位置をルーンインデックスに変換する。
// 全角文字（幅2セル）を考慮して正確なルーン位置を返す。
func CellToRuneIndex(runes []rune, cellCol int) int {
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

// ClampInt は値を [lo, hi] の範囲に制限する。
func ClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}

	if v > hi {
		return hi
	}

	return v
}

// FormatDate は日付を相対的な表現で返す。
// 今日→時刻、昨日→Yesterday、過去7日以内→曜日名、それ以前→YYYY/MM/DD。
func FormatDate(t time.Time, now time.Time) string {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	sevenDaysAgo := today.AddDate(0, 0, -7)

	switch {
	case !t.Before(today):
		return t.Format("15:04")
	case !t.Before(yesterday):
		return "Yesterday"
	case !t.Before(sevenDaysAgo):
		return t.Format("Monday")
	default:
		return t.Format("2006/01/02")
	}
}
