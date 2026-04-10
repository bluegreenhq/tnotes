package app_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/store"
)

func TestPurgeTrash(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	// ノートを作成してゴミ箱に移動
	now := time.Now()
	for range 3 {
		result, err := a.CreateNote(now)
		require.NoError(t, err)

		_, err = a.SaveNote(result.Note.ID, "test\nbody", now)
		require.NoError(t, err)
	}

	for range 3 {
		_, err := a.TrashNote(0)
		require.NoError(t, err)
	}

	count, err := a.PurgeTrash()
	require.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Empty(t, a.TrashNotes)
}

func TestPinNote(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	now := time.Now()
	result, err := a.CreateNote(now)
	require.NoError(t, err)

	err = a.PinNote(result.Note.ID)
	require.NoError(t, err)
	assert.True(t, a.Notes[0].Pinned)

	err = a.UnpinNote(result.Note.ID)
	require.NoError(t, err)
	assert.False(t, a.Notes[0].Pinned)
}

func TestPinNote_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	err = a.PinNote("nonexistent")
	assert.Error(t, err)
}

func TestPurgeTrash_Empty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	count, err := a.PurgeTrash()
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
