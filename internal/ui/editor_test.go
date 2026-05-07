package ui_test

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/ui"
)

const (
	testBodyHello       = "Hello"
	testBodyTwoLines    = "Hello\nWorld"
	testBodyThreeLines  = "Hello\nWorld\nFoo"
	testBodyHelloSpaced = "Hello World"
)

func TestEditorLoadNote(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyTwoLines}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	assert.Equal(t, note.NoteID("1"), ed.NoteID())
	assert.Equal(t, testBodyTwoLines, ed.Value())
}

func TestEditorEmpty(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	assert.Empty(t, ed.NoteID())
}

func TestEditorDirty(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "original"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	assert.False(t, ed.Dirty())

	ed.SetValue("modified")
	assert.True(t, ed.Dirty())

	ed.MarkClean()
	assert.False(t, ed.Dirty())
}

func TestEditorReadOnly(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "read only content"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SetReadOnly(true)
	assert.True(t, ed.ReadOnly())

	ed2, _ := ed.Update(tea.KeyPressMsg{Code: 'x', Text: "x"}, now)
	assert.Equal(t, "read only content", ed2.Value())

	ed2.SetReadOnly(false)
	assert.False(t, ed2.ReadOnly())
}

func TestEditorSelectionBasic(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyThreeLines}
	ed.LoadNote(n)

	assert.False(t, ed.HasSelection())

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 1}, ui.SelectionAnchor{Line: 0, Column: 4})
	assert.True(t, ed.HasSelection())

	start, end := ed.NormalizedSelection()
	assert.Equal(t, ui.SelectionAnchor{Line: 0, Column: 1}, start)
	assert.Equal(t, ui.SelectionAnchor{Line: 0, Column: 4}, end)
}

func TestEditorSelectionNormalize(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyTwoLines}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 1, Column: 3}, ui.SelectionAnchor{Line: 0, Column: 1})
	start, end := ed.NormalizedSelection()
	assert.Equal(t, ui.SelectionAnchor{Line: 0, Column: 1}, start)
	assert.Equal(t, ui.SelectionAnchor{Line: 1, Column: 3}, end)
}

func TestEditorClearSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHello}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 0}, ui.SelectionAnchor{Line: 0, Column: 3})
	assert.True(t, ed.HasSelection())

	ed.ClearSelection()
	assert.False(t, ed.HasSelection())
}

func TestEditorSelectedText(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyThreeLines}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 1}, ui.SelectionAnchor{Line: 0, Column: 4})
	assert.Equal(t, "ell", ed.SelectedText())

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 3}, ui.SelectionAnchor{Line: 1, Column: 2})
	assert.Equal(t, "lo\nWo", ed.SelectedText())

	ed.SetSelection(ui.SelectionAnchor{Line: 1, Column: 2}, ui.SelectionAnchor{Line: 0, Column: 3})
	assert.Equal(t, "lo\nWo", ed.SelectedText())
}

func TestEditorSelectedTextNoSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHello}
	ed.LoadNote(n)

	assert.Empty(t, ed.SelectedText())
}

func TestEditorDeleteSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyThreeLines}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 1}, ui.SelectionAnchor{Line: 0, Column: 4})
	ed.DeleteSelection()
	assert.Equal(t, "Ho\nWorld\nFoo", ed.Value())
	assert.False(t, ed.HasSelection())
}

func TestEditorDeleteSelectionMultiLine(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyThreeLines}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 3}, ui.SelectionAnchor{Line: 2, Column: 1})
	ed.DeleteSelection()
	assert.Equal(t, "Heloo", ed.Value())
	assert.False(t, ed.HasSelection())
}

func TestEditorCopySelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyTwoLines}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 0}, ui.SelectionAnchor{Line: 0, Column: 5})
	err := ed.CopySelection()
	require.NoError(t, err)
	assert.Equal(t, testBodyTwoLines, ed.Value())
	assert.False(t, ed.HasSelection())
}

func TestEditorCutSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyTwoLines}
	ed.LoadNote(n)

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 0}, ui.SelectionAnchor{Line: 0, Column: 5})
	err := ed.CutSelection()
	require.NoError(t, err)
	assert.Equal(t, "\nWorld", ed.Value())
	assert.False(t, ed.HasSelection())
}

func TestEditorShiftArrowSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyTwoLines}
	ed.LoadNote(n)
	ed.Focus()

	ed2, _ := ed.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}, now)
	assert.True(t, ed2.HasSelection())
	assert.Equal(t, "H", ed2.SelectedText())

	ed3, _ := ed2.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}, now)
	assert.Equal(t, "He", ed3.SelectedText())
}

func TestEditorShiftArrowThenPlainArrowClearsSelection(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHello}
	ed.LoadNote(n)
	ed.Focus()

	ed2, _ := ed.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: tea.ModShift}, now)
	assert.True(t, ed2.HasSelection())

	ed3, _ := ed2.Update(tea.KeyPressMsg{Code: tea.KeyRight, Mod: 0}, now)
	assert.False(t, ed3.HasSelection())
}

func TestEditorViewHasHighlight(t *testing.T) {
	t.Parallel()

	ed := ui.NewEditor(60, 20, false)
	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyTwoLines}
	ed.LoadNote(n)
	ed.Focus()

	viewNoSel := ed.View()

	ed.SetSelection(ui.SelectionAnchor{Line: 0, Column: 1}, ui.SelectionAnchor{Line: 0, Column: 4})
	viewWithSel := ed.View()

	assert.NotEqual(t, viewNoSel, viewWithSel)
}

func TestEditorBlinkReset(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHello}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	ed.Focus()

	assert.True(t, ed.BlinkVisible())
}

func TestEditorBlinkStopsOnBlur(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHello}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	ed.Focus()

	ed.Blur()
	assert.True(t, ed.BlinkVisible())
}

func TestEditorBlinkResetsOnKeyPress(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHello}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)
	ed.Focus()

	ed, _ = ed.Update(tea.KeyPressMsg{Code: 'a', Text: "a"}, now)
	assert.True(t, ed.BlinkVisible())
}

func TestEditorSelectWord(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHelloSpaced}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 1) // testBodyHello の中
	assert.True(t, ed.HasSelection())
	assert.Equal(t, testBodyHello, ed.SelectedText())
}

func TestEditorSelectWordPunct(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "foo==bar"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 3) // "==" の中
	assert.True(t, ed.HasSelection())
	assert.Equal(t, "==", ed.SelectedText())
}

func TestEditorSelectWordSpace(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "foo   bar"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 4) // 空白の中
	assert.True(t, ed.HasSelection())
	assert.Equal(t, "   ", ed.SelectedText())
}

func TestEditorSelectWordEmptyLine(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\n\nWorld"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(1, 0) // 空行
	assert.False(t, ed.HasSelection())
}

func TestEditorSelectWordAtEnd(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHello}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 5) // 行末（len == 5）
	assert.True(t, ed.HasSelection())
	assert.Equal(t, testBodyHello, ed.SelectedText())
}

func TestEditorSelectWordUnderscore(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "foo_bar baz"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 4) // "foo_bar" の "_" 上
	assert.True(t, ed.HasSelection())
	assert.Equal(t, "foo_bar", ed.SelectedText())
}

func TestEditorSelectLine(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyThreeLines}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectLine(1)
	assert.True(t, ed.HasSelection())
	assert.Equal(t, "World", ed.SelectedText())
}

func TestEditorSelectLineEmpty(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello\n\nWorld"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectLine(1) // 空行
	assert.False(t, ed.HasSelection())
}

func TestEditorSelectLineFirst(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyTwoLines}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectLine(0)
	assert.True(t, ed.HasSelection())
	assert.Equal(t, testBodyHello, ed.SelectedText())
}

func TestEditorHandleTextAreaClickSingle(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHelloSpaced}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.HandleTextAreaClick(1, 0, now) // シングルクリック
	assert.True(t, ed.Selecting(), "should start drag selection")
}

func TestEditorHandleTextAreaClickDouble(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHelloSpaced}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.HandleTextAreaClick(1, 0, now) // 1回目
	ed.StopDragSelection()
	ed.HandleTextAreaClick(1, 0, now.Add(100*time.Millisecond)) // 2回目（100ms後）
	assert.True(t, ed.HasSelection())
	assert.Equal(t, testBodyHello, ed.SelectedText())
	assert.False(t, ed.Selecting(), "should not be in drag mode after word select")
}

func TestEditorHandleTextAreaClickTriple(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello World\nSecond line"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.HandleTextAreaClick(1, 0, now)
	ed.StopDragSelection()
	ed.HandleTextAreaClick(1, 0, now.Add(100*time.Millisecond))
	ed.HandleTextAreaClick(1, 0, now.Add(200*time.Millisecond))
	assert.True(t, ed.HasSelection())
	assert.Equal(t, testBodyHelloSpaced, ed.SelectedText())
}

func TestEditorHandleTextAreaClickQuadResets(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHelloSpaced}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.HandleTextAreaClick(1, 0, now)
	ed.StopDragSelection()
	ed.HandleTextAreaClick(1, 0, now.Add(100*time.Millisecond))
	ed.HandleTextAreaClick(1, 0, now.Add(200*time.Millisecond))
	ed.HandleTextAreaClick(1, 0, now.Add(300*time.Millisecond)) // 4回目 → リセット
	assert.True(t, ed.Selecting(), "should restart drag selection on 4th click")
}

func TestEditorHandleTextAreaClickTimeout(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHelloSpaced}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.HandleTextAreaClick(1, 0, now)
	ed.StopDragSelection()
	ed.HandleTextAreaClick(1, 0, now.Add(600*time.Millisecond)) // 600ms後 → タイムアウト
	assert.True(t, ed.Selecting(), "should be single click after timeout")
	assert.False(t, ed.HasSelection())
}

func TestEditorHandleTextAreaClickDiffPos(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: testBodyHelloSpaced}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.HandleTextAreaClick(1, 0, now) // 位置 (1,0)
	ed.StopDragSelection()
	ed.HandleTextAreaClick(8, 0, now.Add(100*time.Millisecond)) // 位置 (8,0) → 別の位置
	assert.True(t, ed.Selecting(), "should be single click at different position")
}

func TestEditorSelectWordHiragana(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "これはテスト"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 1) // "は" → ひらがな "これは"
	assert.True(t, ed.HasSelection())
	assert.Equal(t, "これは", ed.SelectedText())
}

func TestEditorSelectWordKatakana(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "これはテスト"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 4) // "ス" → カタカナ "テスト"
	assert.True(t, ed.HasSelection())
	assert.Equal(t, "テスト", ed.SelectedText())
}

func TestEditorSelectWordKatakanaProlonged(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "コントロール"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 3) // "ト" → "コントロール" 全体
	assert.True(t, ed.HasSelection())
	assert.Equal(t, "コントロール", ed.SelectedText())
}

func TestEditorSelectWordKanji(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "今日は良い天気です"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 0) // "今" → 漢字 "今日"
	assert.True(t, ed.HasSelection())
	assert.Equal(t, "今日", ed.SelectedText())
}

func TestEditorSelectWordMixed(t *testing.T) {
	t.Parallel()

	now := time.Now()
	n := note.Note{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "Hello世界テスト"}
	ed := ui.NewEditor(60, 20, false)
	ed.LoadNote(n)

	ed.SelectWord(0, 0) // "H" → ASCII testBodyHello
	assert.Equal(t, testBodyHello, ed.SelectedText())

	ed.SelectWord(0, 5) // "世" → 漢字 "世界"
	assert.Equal(t, "世界", ed.SelectedText())

	ed.SelectWord(0, 7) // "テ" → カタカナ "テスト"
	assert.Equal(t, "テスト", ed.SelectedText())
}
