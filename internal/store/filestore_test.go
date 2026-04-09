package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

	noteDir := filepath.Join(dir, "Notes", "20260404")
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

	_, err = os.Stat(filepath.Join(dir, "Notes", "20260404", "trash1.md"))
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

	_, err = os.Stat(filepath.Join(dir, "Notes", "20260404", "restore1.md"))
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

func TestFileStore_ConcurrentSave(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)

	const numNotes = 10

	var wg sync.WaitGroup
	wg.Add(numNotes)

	errs := make([]error, numNotes)

	for i := range numNotes {
		go func(idx int) {
			defer wg.Done()
			// 各ゴルーチンが独立したFileStoreインスタンスを使用（別プロセスを模倣）
			s, err := store.NewFileStore(dir)
			if err != nil {
				errs[idx] = err

				return
			}

			n := note.Note{
				Metadata: note.Metadata{
					ID:        note.NoteID(fmt.Sprintf("concurrent-%d", idx)),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Body: fmt.Sprintf("Note %d", idx),
			}
			errs[idx] = s.Save(n)
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "goroutine %d failed", i)
	}

	// 新しいFileStoreでindex.jsonを読み直し、全ノートが存在することを確認
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	notes, err := s.List()
	require.NoError(t, err)
	assert.Len(t, notes, numNotes)
}

func TestFileStore_ConcurrentTrash(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	// ノートを作成
	for i := range 5 {
		n := note.Note{
			Metadata: note.Metadata{
				ID:        note.NoteID(fmt.Sprintf("trash-%d", i)),
				CreatedAt: now,
				UpdatedAt: now,
			},
			Body: fmt.Sprintf("Note %d", i),
		}
		require.NoError(t, s.Save(n))
	}

	// 別々のFileStoreインスタンスから同時にTrash
	var wg sync.WaitGroup
	wg.Add(5)

	trashErrs := make([]error, 5)

	for i := range 5 {
		go func(idx int) {
			defer wg.Done()

			si, err := store.NewFileStore(dir)
			if err != nil {
				trashErrs[idx] = err

				return
			}

			trashErrs[idx] = si.Trash(note.NoteID(fmt.Sprintf("trash-%d", idx)))
		}(i)
	}

	wg.Wait()

	for i, err := range trashErrs {
		require.NoError(t, err, "trash goroutine %d failed", i)
	}

	// 検証
	s2, err := store.NewFileStore(dir)
	require.NoError(t, err)

	notes, err := s2.List()
	require.NoError(t, err)
	assert.Empty(t, notes)

	trashed, err := s2.ListTrashed()
	require.NoError(t, err)
	assert.Len(t, trashed, 5)
}

func TestFileStore_PurgeTrash(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	for i := range 3 {
		n := note.Note{
			Metadata: note.Metadata{
				ID:        note.NoteID(fmt.Sprintf("purge-%d", i)),
				CreatedAt: now,
				UpdatedAt: now,
			},
			Body: fmt.Sprintf("Note %d", i),
		}
		require.NoError(t, s.Save(n))
		require.NoError(t, s.Trash(note.NoteID(fmt.Sprintf("purge-%d", i))))
	}

	count, err := s.PurgeTrash()
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	trashed, err := s.ListTrashed()
	require.NoError(t, err)
	assert.Empty(t, trashed)

	// .trash ディレクトリ内のノートファイルが削除されていること
	for i := range 3 {
		_, err := os.Stat(filepath.Join(dir, ".trash", "20260404", fmt.Sprintf("purge-%d.md", i)))
		assert.True(t, os.IsNotExist(err))
	}

	// 通常ノートに影響がないこと
	metas, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, metas)
}

func TestFileStore_PurgeTrash_Empty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	count, err := s.PurgeTrash()
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestFileStore_PurgeTrash_PreservesNormalNotes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	now := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)

	// 通常ノートを作成
	normal := note.Note{
		Metadata: note.Metadata{
			ID:        "normal1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Body: "Keep this",
	}
	require.NoError(t, s.Save(normal))

	// ゴミ箱ノートを作成
	trash := note.Note{
		Metadata: note.Metadata{
			ID:        "trash1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Body: "Delete this",
	}
	require.NoError(t, s.Save(trash))
	require.NoError(t, s.Trash("trash1"))

	count, err := s.PurgeTrash()
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// 通常ノートは残っている
	metas, err := s.List()
	require.NoError(t, err)
	assert.Len(t, metas, 1)
	assert.Equal(t, note.NoteID("normal1"), metas[0].ID)

	// ゴミ箱は空
	trashed, err := s.ListTrashed()
	require.NoError(t, err)
	assert.Empty(t, trashed)
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
