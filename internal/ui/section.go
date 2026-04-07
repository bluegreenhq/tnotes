package ui

import (
	"time"

	"github.com/bluegreenhq/tnotes/internal/note"
)

const sectionHeaderHeight = 1 // セクションヘッダーは1行

// Section はセクションのラベルとノートリストを表す。
type Section struct {
	Label string
	Notes []note.Note
}

func newSection(label string) Section {
	return Section{Label: label, Notes: nil}
}

// GroupNotesBySection はノートを更新日時で日付セクションに分類する。
// 該当ノートがないセクションは含まない。
func GroupNotesBySection(notes []note.Note, now time.Time) []Section {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	sevenDaysAgo := today.AddDate(0, 0, -7)

	all := []Section{
		newSection("Today"),
		newSection("Yesterday"),
		newSection("Previous 7 Days"),
		newSection("Previous 30 Days"),
	}

	for _, n := range notes {
		t := n.UpdatedAt
		switch {
		case !t.Before(today):
			all[0].Notes = append(all[0].Notes, n)
		case !t.Before(yesterday):
			all[1].Notes = append(all[1].Notes, n)
		case !t.Before(sevenDaysAgo):
			all[2].Notes = append(all[2].Notes, n)
		default:
			all[3].Notes = append(all[3].Notes, n)
		}
	}

	var sections []Section

	for _, s := range all {
		if len(s.Notes) > 0 {
			sections = append(sections, s)
		}
	}

	return sections
}
