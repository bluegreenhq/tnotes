package ui_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

const (
	termW = 100
	termH = 30
)

func screen(bts []byte) string {
	return renderScreen(bts, termW, termH)
}

func TestE2E_InitialRender(t *testing.T) {
	t.Parallel()

	m := newTestModel()

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(termW, termH))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := screen(bts)

		return strings.Contains(s, "Notes (0)") &&
			strings.Contains(s, "[New]") &&
			strings.Contains(s, "[Quit]")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'q'})
	tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestE2E_CreateAndEditNote(t *testing.T) {
	t.Parallel()

	m := newTestModel()

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(termW, termH))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Notes (0)")
	}, teatest.WithDuration(3*time.Second))

	// 'n' でノート作成 → New Note がサイドバーに表示される
	tm.Send(tea.KeyPressMsg{Code: 'n'})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "New Note")
	}, teatest.WithDuration(3*time.Second))

	// テキスト入力
	tm.Send(tea.KeyPressMsg{Code: 'H', Text: "H"})
	tm.Send(tea.KeyPressMsg{Code: 'e', Text: "e"})
	tm.Send(tea.KeyPressMsg{Code: 'l', Text: "l"})
	tm.Send(tea.KeyPressMsg{Code: 'l', Text: "l"})
	tm.Send(tea.KeyPressMsg{Code: 'o', Text: "o"})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Hello")
	}, teatest.WithDuration(3*time.Second))

	// Esc でサイドバーに戻る → タイトルが "Hello" に更新される
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEscape})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := screen(bts)

		return strings.Contains(s, "Hello") && !strings.Contains(s, "New Note")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'q'})
	tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestE2E_MultipleNotesNavigation(t *testing.T) {
	t.Parallel()

	m := newTestModel()

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(termW, termH))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Notes (0)")
	}, teatest.WithDuration(3*time.Second))

	// 1つ目のノート作成
	tm.Send(tea.KeyPressMsg{Code: 'n'})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "New Note")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'A', Text: "A"})
	tm.Send(tea.KeyPressMsg{Code: 'A', Text: "A"})
	tm.Send(tea.KeyPressMsg{Code: 'A', Text: "A"})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEscape})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "AAA")
	}, teatest.WithDuration(3*time.Second))

	// 2つ目のノート作成
	tm.Send(tea.KeyPressMsg{Code: 'n'})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := screen(bts)

		return strings.Contains(s, "Notes (2)") || strings.Contains(s, "2)")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'B', Text: "B"})
	tm.Send(tea.KeyPressMsg{Code: 'B', Text: "B"})
	tm.Send(tea.KeyPressMsg{Code: 'B', Text: "B"})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEscape})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "BBB")
	}, teatest.WithDuration(3*time.Second))

	// j で下に移動（AAAのノートへ）→ Tab でエディタに切り替え
	tm.Send(tea.KeyPressMsg{Code: 'j'})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyTab})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "AAA")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEscape})
	tm.Send(tea.KeyPressMsg{Code: 'q'})
	tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestE2E_QuitWithCtrlC(t *testing.T) {
	t.Parallel()

	m := newTestModel()

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(termW, termH))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Notes (0)")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestE2E_TrashAndRestore(t *testing.T) {
	t.Parallel()

	m := newTestModel()

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(termW, termH))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Notes (0)")
	}, teatest.WithDuration(3*time.Second))

	// ノート作成
	tm.Send(tea.KeyPressMsg{Code: 'n'})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "New Note")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'T', Text: "T"})
	tm.Send(tea.KeyPressMsg{Code: 'e', Text: "e"})
	tm.Send(tea.KeyPressMsg{Code: 's', Text: "s"})
	tm.Send(tea.KeyPressMsg{Code: 't', Text: "t"})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Test")
	}, teatest.WithDuration(3*time.Second))

	// Esc でサイドバーに戻る → タイトルが "Test" に更新される
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEscape})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		s := screen(bts)

		return strings.Contains(s, "Test") && !strings.Contains(s, "New Note")
	}, teatest.WithDuration(3*time.Second))

	// d で削除 → エディタが空になり "Press 'n'" が表示される
	tm.Send(tea.KeyPressMsg{Code: 'd'})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Press 'n'")
	}, teatest.WithDuration(3*time.Second))

	// g でゴミ箱モードに入る → "Trash" ヘッダーが表示される
	tm.Send(tea.KeyPressMsg{Code: 'g'})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Trash")
	}, teatest.WithDuration(3*time.Second))

	// r で復元 → "Notes" ヘッダーに戻る
	tm.Send(tea.KeyPressMsg{Code: 'r'})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(screen(bts), "Notes")
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyPressMsg{Code: 'q'})
	tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second))
}
