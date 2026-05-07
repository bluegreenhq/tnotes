package cli_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/cli"
	"github.com/bluegreenhq/tnotes/internal/store"
)

const (
	cmdName        = "tnotes"
	cmdList        = "list"
	cmdGet         = "get"
	cmdCreate      = "create"
	cmdExport      = "export"
	cmdImport      = "import"
	cmdFolder      = "folder"
	cmdMove        = "move"
	cmdSearch      = "search"
	cmdUpdate      = "update"
	cmdDelete      = "delete"
	cmdPurge       = "purge"
	cmdUnknown     = "unknown"
	cmdNonexistent = "nonexistent"
	flagFolder     = "--folder"
	flagJSON       = "--json"
	flagForce      = "--force"
	folderWork     = "Work"
)

func TestRun_NoArgs_ReturnsFalse(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName}, nil, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.False(t, got)
}

func TestRun_UnknownCommand_PrintsError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdUnknown}, nil, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, buf.String(), "unknown command")
}

func TestRun_Help_PrintsUsage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, "help"}, nil, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Usage:")
}

func newTestApp(t *testing.T) *app.App {
	t.Helper()

	dir := t.TempDir()

	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	return a
}

func TestRun_List_Empty(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdList}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "No notes")
}

func TestRun_List_WithNotes(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	result, _ := a.CreateNote(time.Now(), "")
	_, _ = a.SaveNote(result.Note.ID, "Hello World\nThis is body", time.Now())

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdList}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), string(result.Note.ID))
	assert.Contains(t, buf.String(), "Hello World")
}

func TestRun_Get_MissingID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdGet}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, buf.String(), "Usage: tnotes get <id>")
}

func TestRun_Get_NotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdGet, cmdNonexistent}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	assert.Error(t, err)
}

func TestRun_Get_Found(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	result, _ := a.CreateNote(time.Now(), "")
	_, _ = a.SaveNote(result.Note.ID, "My Title\nMy body content", time.Now())

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdGet, string(result.Note.ID)}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "My Title")
	assert.Contains(t, buf.String(), "My body content")
}

func TestRun_Create_FromStdin(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	stdin := strings.NewReader("Hello from stdin\nBody line")

	got, err := cli.Run([]string{cmdName, cmdCreate}, a, stdin, &buf)
	require.NoError(t, err)
	assert.True(t, got)

	// 出力にノートIDが含まれる
	id := strings.TrimSpace(buf.String())
	assert.Len(t, id, 16, "ノートIDは16文字のhex")

	// ノートが保存されていることを確認
	var getBuf bytes.Buffer

	_, _ = cli.Run([]string{cmdName, cmdGet, id}, a, strings.NewReader(""), &getBuf)
	assert.Contains(t, getBuf.String(), "Hello from stdin")
	assert.Contains(t, getBuf.String(), "Body line")
}

func TestRun_Create_FromFile(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	// テスト用ファイルを作成
	tmpFile := filepath.Join(t.TempDir(), "note.md")
	err := os.WriteFile(tmpFile, []byte("File note title\nFile body"), 0o600)
	require.NoError(t, err)

	var buf bytes.Buffer

	got, createErr := cli.Run([]string{cmdName, cmdCreate, tmpFile}, a, strings.NewReader(""), &buf)
	require.NoError(t, createErr)
	assert.True(t, got)

	id := strings.TrimSpace(buf.String())
	assert.Len(t, id, 16)

	var getBuf bytes.Buffer

	_, _ = cli.Run([]string{cmdName, cmdGet, id}, a, strings.NewReader(""), &getBuf)
	assert.Contains(t, getBuf.String(), "File note title")
	assert.Contains(t, getBuf.String(), "File body")
}

func TestRun_Create_EmptyInput(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdCreate}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty input")
}

func TestRun_Create_FileNotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdCreate, "/nonexistent/file.md"}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	assert.Error(t, err)
}

func TestRun_Export_MissingArg(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdExport}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, buf.String(), "Usage: tnotes export")
}

func TestRun_Export_Success(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Export test\nBody", now)

	outPath := filepath.Join(t.TempDir(), "backup.zip")

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdExport, outPath}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)

	_, statErr := os.Stat(outPath)
	require.NoError(t, statErr)

	r, zipErr := zip.OpenReader(outPath)
	require.NoError(t, zipErr)

	defer r.Close()

	fileNames := make([]string, 0, len(r.File))

	for _, f := range r.File {
		fileNames = append(fileNames, f.Name)
	}

	assert.Contains(t, fileNames, "index.json")

	hasMD := false

	for _, name := range fileNames {
		if strings.HasSuffix(name, ".md") {
			hasMD = true

			break
		}
	}

	assert.True(t, hasMD, "zipに.mdファイルが含まれるべき")
}

func TestRun_Export_WithTrash(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Trash test\nBody", now)
	_, _ = a.TrashNote(result.Note.ID)

	outPath := filepath.Join(t.TempDir(), "backup.zip")

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdExport, outPath}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)

	r, zipErr := zip.OpenReader(outPath)
	require.NoError(t, zipErr)

	defer r.Close()

	hasTrash := false

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, ".trash/") {
			hasTrash = true

			break
		}
	}

	assert.True(t, hasTrash, "zipにtrashファイルが含まれるべき")
}

func TestRun_Export_FileExists(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	outPath := filepath.Join(t.TempDir(), "backup.zip")
	err := os.WriteFile(outPath, []byte("existing"), 0o600)
	require.NoError(t, err)

	var buf bytes.Buffer

	got, runErr := cli.Run([]string{cmdName, cmdExport, outPath}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, runErr)
}

func TestRun_Import_MissingArg(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdImport}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, buf.String(), "Usage: tnotes import")
}

func TestRun_Import_DataExists(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Existing note\nBody", now)

	zipPath := filepath.Join(t.TempDir(), "import.zip")
	createEmptyZip(t, zipPath)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdImport, zipPath}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data directory is not empty")
}

func TestRun_Import_Success(t *testing.T) {
	t.Parallel()

	// まずexportでzipを作成
	srcApp := newTestApp(t)
	now := time.Now()
	result, _ := srcApp.CreateNote(now, "")
	_, _ = srcApp.SaveNote(result.Note.ID, "Import test\nBody", now)

	zipPath := filepath.Join(t.TempDir(), "export.zip")

	var exportBuf bytes.Buffer

	_, exportErr := cli.Run([]string{cmdName, cmdExport, zipPath}, srcApp, strings.NewReader(""), &exportBuf)
	require.NoError(t, exportErr)

	// 空のAppにimport
	dstDir := t.TempDir()
	dstStore, _ := store.NewFileStore(dstDir)
	dstApp, _ := app.New(dstStore)

	var importBuf bytes.Buffer

	got, importErr := cli.Run([]string{cmdName, cmdImport, zipPath}, dstApp, strings.NewReader(""), &importBuf)
	require.NoError(t, importErr)
	assert.True(t, got)

	// import先でノートが読めることを確認
	reloadStore, _ := store.NewFileStore(dstDir)
	reloadApp, _ := app.New(reloadStore)
	assert.Len(t, reloadApp.Notes, 1)
	assert.Equal(t, result.Note.ID, reloadApp.Notes[0].ID)
}

func createEmptyZip(t *testing.T, path string) {
	t.Helper()

	f, err := os.Create(path)
	require.NoError(t, err)

	defer f.Close()

	zw := zip.NewWriter(f)
	require.NoError(t, zw.Close())
}

func TestRun_Purge_Force(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Purge test\nBody", now)
	_, _ = a.TrashNote(result.Note.ID)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdPurge, flagForce}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "1")
}

func TestRun_Purge_Empty(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdPurge, flagForce}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Trash is empty")
}

func TestRun_Purge_ConfirmYes(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Confirm test\nBody", now)
	_, _ = a.TrashNote(result.Note.ID)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdPurge}, a, strings.NewReader("y\n"), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "1")
}

func TestRun_Purge_ConfirmNo(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Cancel test\nBody", now)
	_, _ = a.TrashNote(result.Note.ID)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdPurge}, a, strings.NewReader("n\n"), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Cancelled")

	// ゴミ箱にはまだノートが残っている
	require.NoError(t, a.RefreshTrashNotes())
	assert.Len(t, a.ListTrashNotes(), 1)
}

func TestRun_Purge_ConfirmEmpty(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Default no\nBody", now)
	_, _ = a.TrashNote(result.Note.ID)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdPurge}, a, strings.NewReader("\n"), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Cancelled")
}

func TestRun_List_WithFolderFlag(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Default note\nBody", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdList, flagFolder, "Notes"}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Default note")
}

func TestRun_List_WithFolderFlag_Empty(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	require.NoError(t, a.CreateFolder(folderWork))

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdList, flagFolder, folderWork}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "No notes")
}

func TestRun_Folder_List_Empty(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdFolder, cmdList}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Notes")
	assert.Contains(t, buf.String(), "Trash")
}

func TestRun_Folder_Create(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdFolder, cmdCreate, folderWork}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Created folder: Work")

	var listBuf bytes.Buffer

	_, _ = cli.Run([]string{cmdName, cmdFolder, cmdList}, a, strings.NewReader(""), &listBuf)
	assert.Contains(t, listBuf.String(), folderWork)
}

func TestRun_Folder_Create_MissingName(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdFolder, cmdCreate}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Folder_Delete_Empty(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	require.NoError(t, a.CreateFolder(folderWork))

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdFolder, cmdDelete, folderWork}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Deleted folder: Work")
}

func TestRun_Folder_Delete_MissingName(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdFolder, cmdDelete}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Folder_Delete_Force(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	require.NoError(t, a.CreateFolder(folderWork))

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdFolder, cmdDelete, folderWork, flagForce}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Deleted folder: Work")
}

func TestRun_List_WithFolderFlag_NotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdList, flagFolder, cmdUnknown}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "folder not found")
}

func TestRun_Create_WithFolder(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	require.NoError(t, a.CreateFolder(folderWork))

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdCreate, flagFolder, folderWork}, a, strings.NewReader("work note\nbody"), &buf)
	require.NoError(t, err)
	assert.True(t, got)

	workNotes := a.ListByFolder(folderWork)
	assert.Len(t, workNotes, 1)
}

func TestRun_Create_WithFolder_NotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdCreate, flagFolder, cmdUnknown}, a, strings.NewReader("test\nbody"), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "folder not found")
}

func TestRun_Move_Success(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	require.NoError(t, a.CreateFolder(folderWork))

	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "move test\nbody", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdMove, string(result.Note.ID), folderWork}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Moved")

	workNotes := a.ListByFolder(folderWork)
	assert.Len(t, workNotes, 1)
}

func TestRun_Move_MissingArgs(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdMove}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Move_NoteNotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	require.NoError(t, a.CreateFolder(folderWork))

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdMove, cmdNonexistent, folderWork}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Move_FolderNotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "test\nbody", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdMove, string(result.Note.ID), cmdUnknown}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Folder_NoSubcommand(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdFolder}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Version_PrintsVersion(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, "version"}, nil, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "tnotes version ")
}

func TestRun_List_JSON(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "JSON test\nBody", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdList, flagJSON}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)

	var notes []map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &notes))
	assert.Len(t, notes, 1)
	assert.Equal(t, string(result.Note.ID), notes[0]["id"])
	assert.Equal(t, "JSON test", notes[0]["title"])
}

func TestRun_List_JSON_Empty(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdList, flagJSON}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)

	var notes []map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &notes))
	assert.Empty(t, notes)
}

func TestRun_Get_JSON(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Get JSON\nBody content", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdGet, string(result.Note.ID), flagJSON}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)

	var note map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &note))
	assert.Equal(t, string(result.Note.ID), note["id"])
	assert.Equal(t, "Get JSON", note["title"])
	assert.Contains(t, note["body"], "Body content")
}

func TestRun_Folder_List_JSON(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	require.NoError(t, a.CreateFolder(folderWork))

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdFolder, cmdList, flagJSON}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)

	var folders []string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &folders))
	assert.Contains(t, folders, "Notes")
	assert.Contains(t, folders, folderWork)
}

func TestRun_Search_Found(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Meeting notes\nDiscuss project plan", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdSearch, "meeting"}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), string(result.Note.ID))
}

func TestRun_Search_NotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Meeting notes\nBody", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdSearch, cmdNonexistent}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "No matches")
}

func TestRun_Search_JSON(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Search JSON test\nBody content", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdSearch, "json", flagJSON}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)

	var notes []map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &notes))
	assert.Len(t, notes, 1)
	assert.Equal(t, string(result.Note.ID), notes[0]["id"])

	snippets, ok := notes[0]["snippets"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, snippets)
	assert.Contains(t, snippets[0], "JSON")
}

func TestRun_Search_BodyMatch(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Title only\nSecret body keyword here", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdSearch, "keyword"}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), string(result.Note.ID))
	// スニペットにマッチ箇所の前後が含まれる
	assert.Contains(t, buf.String(), "keyword")
	assert.Contains(t, buf.String(), "Secret body")
}

func TestRun_Search_CaseInsensitive(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "UPPERCASE Title\nBody", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdSearch, "uppercase"}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), string(result.Note.ID))
}

func TestRun_Search_WithFolder(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	require.NoError(t, a.CreateFolder(folderWork))

	now := time.Now()
	r1, _ := a.CreateNote(now, folderWork)
	_, _ = a.SaveNote(r1.Note.ID, "Work note\nBody", now)
	r2, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(r2.Note.ID, "Default note\nBody", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdSearch, "note", flagFolder, folderWork}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), string(r1.Note.ID))
	assert.NotContains(t, buf.String(), string(r2.Note.ID))
}

func TestRun_Search_Context(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "AAAA target BBBB", now)

	var buf bytes.Buffer

	// context=2 で前後2文字のみ
	got, err := cli.Run([]string{cmdName, cmdSearch, "target", "--context", "2"}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	// "...A target BB..." のようなスニ���ット
	assert.Contains(t, buf.String(), "target")
	assert.Contains(t, buf.String(), "...")
}

func TestRun_Search_Snippet_NoEllipsis(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "short match", now)

	var buf bytes.Buffer

	// context が十分大きければ ... がつかない
	got, err := cli.Run([]string{cmdName, cmdSearch, "short", "--context", "100"}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "short match")
	assert.NotContains(t, buf.String(), "...")
}

func TestRun_Search_MissingQuery(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdSearch}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Update_Success(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Old title\nOld body", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdUpdate, string(result.Note.ID)}, a, strings.NewReader("New title\nNew body"), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), string(result.Note.ID))

	// 更新された内容を確認
	var getBuf bytes.Buffer

	_, _ = cli.Run([]string{cmdName, cmdGet, string(result.Note.ID)}, a, strings.NewReader(""), &getBuf)
	assert.Contains(t, getBuf.String(), "New title")
	assert.Contains(t, getBuf.String(), "New body")
}

func TestRun_Update_FromFile(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Old content", now)

	tmpFile := filepath.Join(t.TempDir(), "update.md")
	err := os.WriteFile(tmpFile, []byte("Updated from file\nNew body"), 0o600)
	require.NoError(t, err)

	var buf bytes.Buffer

	got, runErr := cli.Run([]string{cmdName, cmdUpdate, string(result.Note.ID), tmpFile}, a, strings.NewReader(""), &buf)
	require.NoError(t, runErr)
	assert.True(t, got)

	var getBuf bytes.Buffer

	_, _ = cli.Run([]string{cmdName, cmdGet, string(result.Note.ID)}, a, strings.NewReader(""), &getBuf)
	assert.Contains(t, getBuf.String(), "Updated from file")
}

func TestRun_Update_MissingID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdUpdate}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Update_NotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdUpdate, cmdNonexistent}, a, strings.NewReader("new body"), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Update_EmptyInput(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Existing", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdUpdate, string(result.Note.ID)}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Delete_Success(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now, "")
	_, _ = a.SaveNote(result.Note.ID, "Delete me\nBody", now)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdDelete, string(result.Note.ID)}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "Deleted")

	// ゴミ箱に移動されたことを確認
	require.NoError(t, a.RefreshTrashNotes())
	assert.Len(t, a.ListTrashNotes(), 1)
	assert.Empty(t, a.ListNotes())
}

func TestRun_Delete_MissingID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdDelete}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}

func TestRun_Delete_NotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{cmdName, cmdDelete, cmdNonexistent}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
}
