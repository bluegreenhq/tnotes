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
		result, err := a.CreateNote(now, "")
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
	result, err := a.CreateNote(now, "")
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

func TestListFolders_Empty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	folders, err := a.ListFolders()
	require.NoError(t, err)
	assert.Empty(t, folders)
}

func TestCreateFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	err = a.CreateFolder("Work")
	require.NoError(t, err)

	folders, err := a.ListFolders()
	require.NoError(t, err)
	assert.Equal(t, []string{"Work"}, folders)
}

func TestDeleteFolder_Empty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	require.NoError(t, a.CreateFolder("Work"))

	count, err := a.DeleteFolder("Work")
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	folders, err := a.ListFolders()
	require.NoError(t, err)
	assert.Empty(t, folders)
}

func TestFolderNoteCount_Empty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	require.NoError(t, a.CreateFolder("Work"))

	count, err := a.FolderNoteCount("Work")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestFolderExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	// Notes（デフォルトフォルダ）は常に存在する
	exists, err := a.FolderExists("Notes")
	require.NoError(t, err)
	assert.True(t, exists)

	// 存在しないフォルダ
	exists, err = a.FolderExists("Work")
	require.NoError(t, err)
	assert.False(t, exists)

	// フォルダ作成後
	require.NoError(t, a.CreateFolder("Work"))

	exists, err = a.FolderExists("Work")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestListByFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	now := time.Now()

	// Notesにノート作成
	r1, err := a.CreateNote(now, "")
	require.NoError(t, err)
	_, err = a.SaveNote(r1.Note.ID, "notes note\nbody", now)
	require.NoError(t, err)

	// Workフォルダにノート作成
	require.NoError(t, a.CreateFolder("Work"))

	r2, err := a.CreateNote(now, "Work")
	require.NoError(t, err)
	_, err = a.SaveNote(r2.Note.ID, "work note\nbody", now)
	require.NoError(t, err)

	// Notesフォルダのノートのみ
	notesNotes := a.ListByFolder("Notes")
	assert.Len(t, notesNotes, 1)
	assert.Equal(t, r1.Note.ID, notesNotes[0].ID)

	// Workフォルダのノートのみ
	workNotes := a.ListByFolder("Work")
	assert.Len(t, workNotes, 1)
	assert.Equal(t, r2.Note.ID, workNotes[0].ID)

	// 空フォルダ
	require.NoError(t, a.CreateFolder("Empty"))
	emptyNotes := a.ListByFolder("Empty")
	assert.Empty(t, emptyNotes)
}

func TestCreateNote_WithFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	require.NoError(t, a.CreateFolder("Work"))

	now := time.Now()
	result, err := a.CreateNote(now, "Work")
	require.NoError(t, err)

	assert.Contains(t, result.Note.Path, "Work/")
	assert.NotContains(t, result.Note.Path, "Notes/")
}

func TestCreateNote_DefaultFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	now := time.Now()
	result, err := a.CreateNote(now, "")
	require.NoError(t, err)

	assert.Contains(t, result.Note.Path, "Notes/")
}

func TestDeleteFolder_WithNotes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	require.NoError(t, a.CreateFolder("Work"))

	now := time.Now()
	r1, err := a.CreateNote(now, "Work")
	require.NoError(t, err)
	_, err = a.SaveNote(r1.Note.ID, "work note\nbody", now)
	require.NoError(t, err)

	count, err := a.DeleteFolder("Work")
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// ノートはゴミ箱に移動
	trash, err := a.ListTrash()
	require.NoError(t, err)
	assert.Len(t, trash, 1)

	// フォルダ一覧から消えている
	folders, err := a.ListFolders()
	require.NoError(t, err)
	assert.Empty(t, folders)
}

func TestMoveNoteToFolder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	require.NoError(t, a.CreateFolder("Work"))

	now := time.Now()
	result, err := a.CreateNote(now, "")
	require.NoError(t, err)

	_, err = a.SaveNote(result.Note.ID, "test\nbody", now)
	require.NoError(t, err)

	err = a.MoveNoteToFolder(result.Note.ID, "Work")
	require.NoError(t, err)

	notesNotes := a.ListByFolder("Notes")
	assert.Empty(t, notesNotes)

	workNotes := a.ListByFolder("Work")
	assert.Len(t, workNotes, 1)
	assert.Equal(t, result.Note.ID, workNotes[0].ID)
}

func TestMoveNoteToFolder_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	err = a.MoveNoteToFolder("nonexistent", "Work")
	assert.Error(t, err)
}

func TestMoveNoteToFolder_BackToNotes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	require.NoError(t, a.CreateFolder("Work"))

	now := time.Now()
	result, err := a.CreateNote(now, "Work")
	require.NoError(t, err)

	_, err = a.SaveNote(result.Note.ID, "test\nbody", now)
	require.NoError(t, err)

	err = a.MoveNoteToFolder(result.Note.ID, "Notes")
	require.NoError(t, err)

	notesNotes := a.ListByFolder("Notes")
	assert.Len(t, notesNotes, 1)

	workNotes := a.ListByFolder("Work")
	assert.Empty(t, workNotes)
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

func TestDiscardIfEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	now := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)

	// 空ノートを作成
	result, err := a.CreateNote(now, "")
	require.NoError(t, err)

	id := result.Note.ID

	// 空なので破棄される
	discarded := a.DiscardIfEmpty(id)
	assert.True(t, discarded)
	assert.Empty(t, a.Notes)

	// ストアからも消えている
	list, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestDiscardIfEmpty_NonEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	now := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)

	result, err := a.CreateNote(now, "")
	require.NoError(t, err)

	id := result.Note.ID

	// 本文を書き込む
	_, err = a.SaveNote(id, "hello", now)
	require.NoError(t, err)

	// 空でないので破棄されない
	discarded := a.DiscardIfEmpty(id)
	assert.False(t, discarded)
	assert.Len(t, a.Notes, 1)
}

func TestDiscardIfEmpty_NotFound(t *testing.T) {
	t.Parallel()

	a, err := app.New(nil)
	require.NoError(t, err)

	// 存在しないIDは何もしない
	discarded := a.DiscardIfEmpty("nonexistent")
	assert.False(t, discarded)
}
