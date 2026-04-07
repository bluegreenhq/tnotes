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

// FormatDate は日付を今日なら時刻、それ以外なら日付で返す。
func FormatDate(t time.Time, now time.Time) string {
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04")
	}

	return t.Format("2006/01/02")
}
