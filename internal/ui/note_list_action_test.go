package ui //nolint:testpackage // 内部アクション型（unexported）を検査するためのホワイトボックステスト

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/bluegreenhq/tnotes/internal/note"
)

func TestNoteListUpdateEmitsCreateAction(t *testing.T) {
	t.Parallel()

	now := time.Now()
	nl := NewNoteList(nil, nil, 30, 20)
	_, cmd := nl.Update(tea.KeyPressMsg{Code: 'n'}, now, false)
	assert.NotNil(t, cmd)

	am, ok := cmd().(actionMsg)
	assert.True(t, ok, "Cmd の中身は actionMsg であるべき")

	_, isCreate := am.action.(noteListCreateAction)
	assert.True(t, isCreate, "action は noteListCreateAction であるべき")
}

func TestNoteListUpdateTrashModeEmitsMenuAction(t *testing.T) {
	t.Parallel()

	now := time.Now()
	notes := []note.Note{
		{Metadata: note.Metadata{ID: "1", CreatedAt: now, UpdatedAt: now}, Body: "A"},
	}
	nl := NewNoteList(nil, notes, 30, 20)
	_, cmd := nl.Update(tea.KeyPressMsg{Code: 'm'}, now, true)
	assert.NotNil(t, cmd)

	am, ok := cmd().(actionMsg)
	assert.True(t, ok, "Cmd の中身は actionMsg であるべき")

	_, isMenu := am.action.(noteListMenuAction)
	assert.True(t, isMenu, "action は noteListMenuAction であるべき")
}
