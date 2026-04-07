package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/store"
)

func TestFileStore_SaveAndLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	n := note.Note{
		Metadata: note.Metadata{
			ID:        "test1",
			CreatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
		},
		Body: "Hello World",
	}
	err = s.Save(n)
	require.NoError(t, err)

	noteDir := filepath.Join(dir, "20260404")
	_, err = os.Stat(filepath.Join(noteDir, "test1.md"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "index.json"))
	require.NoError(t, err)

	loaded, err := s.Load("test1")
	require.NoError(t, err)
	assert.Equal(t, note.NoteID("test1"), loaded.ID)
	assert.Equal(t, "Hello World", loaded.Body)
	assert.True(t, n.CreatedAt.Equal(loaded.CreatedAt))
}

func TestFileStore_List(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	n1 := note.Note{
		Metadata: note.Metadata{
			ID:        "note1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Body: "First Note\nsome content",
	}
	n2 := note.Note{
		Metadata: note.Metadata{
			ID:        "note2",
			CreatedAt: now.Add(time.Hour),
			UpdatedAt: now.Add(time.Hour),
		},
		Body: "Second Note",
	}

	require.NoError(t, s.Save(n1))
	require.NoError(t, s.Save(n2))

	metas, err := s.List()
	require.NoError(t, err)
	assert.Len(t, metas, 2)
}

func TestFileStore_SaveUpdate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	n := note.Note{
		Metadata: note.Metadata{
			ID:        "upd1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Body: "Original",
	}
	require.NoError(t, s.Save(n))

	n.Body = "Updated"
	n.UpdatedAt = now.Add(time.Hour)
	require.NoError(t, s.Save(n))

	loaded, err := s.Load("upd1")
	require.NoError(t, err)
	assert.Equal(t, "Updated", loaded.Body)

	metas, err := s.List()
	require.NoError(t, err)
	assert.Len(t, metas, 1)
	assert.Equal(t, "Updated", metas[0].Title())
}

func TestFileStore_LoadNotFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	_, err = s.Load("nonexistent")
	assert.Error(t, err)
}

func TestFileStore_ListEmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	metas, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, metas)
}

func TestFileStore_PersistAcrossInstances(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	s1, err := store.NewFileStore(dir)
	require.NoError(t, err)

	n := note.Note{
		Metadata: note.Metadata{
			ID:        "persist1",
			CreatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC),
		},
		Body: "Persistent",
	}
	require.NoError(t, s1.Save(n))

	s2, err := store.NewFileStore(dir)
	require.NoError(t, err)
	metas, err := s2.List()
	require.NoError(t, err)
	assert.Len(t, metas, 1)
	assert.Equal(t, note.NoteID("persist1"), metas[0].ID)
	assert.Equal(t, "Persistent", metas[0].Title())

	loaded, err := s2.Load("persist1")
	require.NoError(t, err)
	assert.Equal(t, "Persistent", loaded.Body)
}

func TestFileStore_SaveListLoadRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	for i := range 3 {
		n := note.Note{
			Metadata: note.Metadata{
				ID:        note.NoteID(fmt.Sprintf("note%d", i)),
				CreatedAt: now.Add(time.Duration(i) * time.Hour),
				UpdatedAt: now.Add(time.Duration(i) * time.Hour),
			},
			Body: fmt.Sprintf("Note %d content\nSecond line", i),
		}
		require.NoError(t, s.Save(n))
	}

	s2, err := store.NewFileStore(dir)
	require.NoError(t, err)

	metas, err := s2.List()
	require.NoError(t, err)
	assert.Len(t, metas, 3)

	for _, meta := range metas {
		loaded, err := s2.Load(meta.ID)
		require.NoError(t, err)
		assert.Contains(t, loaded.Body, "content")
		assert.Contains(t, loaded.Body, "Second line")
	}
}

func TestFileStore_Trash(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	n := note.Note{
		Metadata: note.Metadata{
			ID:        "trash1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Body: "To be trashed",
	}
	require.NoError(t, s.Save(n))

	err = s.Trash("trash1")
	require.NoError(t, err)

	metas, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, metas)

	trashed, err := s.ListTrashed()
	require.NoError(t, err)
	assert.Len(t, trashed, 1)
	assert.Equal(t, note.NoteID("trash1"), trashed[0].ID)

	_, err = os.Stat(filepath.Join(dir, ".trash", "20260404", "trash1.md"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "20260404", "trash1.md"))
	assert.True(t, os.IsNotExist(err))
}

func TestFileStore_Restore(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	n := note.Note{
		Metadata: note.Metadata{
			ID:        "restore1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Body: "To be restored",
	}
	require.NoError(t, s.Save(n))
	require.NoError(t, s.Trash("restore1"))

	err = s.Restore("restore1")
	require.NoError(t, err)

	metas, err := s.List()
	require.NoError(t, err)
	assert.Len(t, metas, 1)
	assert.Equal(t, note.NoteID("restore1"), metas[0].ID)

	trashed, err := s.ListTrashed()
	require.NoError(t, err)
	assert.Empty(t, trashed)

	_, err = os.Stat(filepath.Join(dir, "20260404", "restore1.md"))
	require.NoError(t, err)

	loaded, err := s.Load("restore1")
	require.NoError(t, err)
	assert.Equal(t, "To be restored", loaded.Body)
}

func TestFileStore_TrashNotFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	err = s.Trash("nonexistent")
	assert.Error(t, err)
}

func TestFileStore_RestoreNotFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	err = s.Restore("nonexistent")
	assert.Error(t, err)
}

func TestFileStore_TrashPersistAcrossInstances(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s1, err := store.NewFileStore(dir)
	require.NoError(t, err)

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	n := note.Note{
		Metadata: note.Metadata{
			ID:        "persist-trash",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Body: "Persistent trash",
	}
	require.NoError(t, s1.Save(n))
	require.NoError(t, s1.Trash("persist-trash"))

	s2, err := store.NewFileStore(dir)
	require.NoError(t, err)

	metas, err := s2.List()
	require.NoError(t, err)
	assert.Empty(t, metas)

	trashed, err := s2.ListTrashed()
	require.NoError(t, err)
	assert.Len(t, trashed, 1)
	assert.Equal(t, note.NoteID("persist-trash"), trashed[0].ID)
}
