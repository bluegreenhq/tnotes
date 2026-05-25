package ui //nolint:testpackage // 内部アクション関数（unexported）を検査するためのホワイトボックステスト

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/note"
)

func TestNoteListUpdateEmitsCreateAction(t *testing.T) {
	t.Parallel()

	now := time.Now()
	nl := NewNoteList(nil, nil, 30, 20)
	_, action := nl.Update(tea.KeyPressMsg{Code: 'n'}, now, false)
	require.NotNil(t, action, "Update は ModelAction を返すべき")

	assert.True(t, sameAction(action, createNote), "action は createNote であるべき")
}

func TestNoteListUpdateTrashModeEmitsMenuAction(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
	}
	nl := NewNoteList(nil, notes, 30, 20)
	_, action := nl.Update(tea.KeyPressMsg{Code: 'm'}, now, true)
	require.NotNil(t, action, "Update は ModelAction を返すべき")

	assert.True(t, sameAction(action, openNoteListMenu), "action は openNoteListMenu であるべき")
}
