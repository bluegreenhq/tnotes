package cli_test

import (
	"archive/zip"
	"bytes"
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

func TestRun_NoArgs_ReturnsFalse(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes"}, nil, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.False(t, got)
}

func TestRun_UnknownCommand_PrintsError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "unknown"}, nil, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, buf.String(), "unknown command")
}

func TestRun_Help_PrintsUsage(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "help"}, nil, strings.NewReader(""), &buf)
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

	got, err := cli.Run([]string{"tnotes", "list"}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), "No notes")
}

func TestRun_List_WithNotes(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	result, _ := a.CreateNote(time.Now())
	_, _ = a.SaveNote(result.Note.ID, "Hello World\nThis is body", time.Now())

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "list"}, a, strings.NewReader(""), &buf)
	require.NoError(t, err)
	assert.True(t, got)
	assert.Contains(t, buf.String(), string(result.Note.ID))
	assert.Contains(t, buf.String(), "Hello World")
}

func TestRun_Get_MissingID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "get"}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, buf.String(), "Usage: tnotes get <id>")
}

func TestRun_Get_NotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "get", "nonexistent"}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	assert.Error(t, err)
}

func TestRun_Get_Found(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	result, _ := a.CreateNote(time.Now())
	_, _ = a.SaveNote(result.Note.ID, "My Title\nMy body content", time.Now())

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "get", string(result.Note.ID)}, a, strings.NewReader(""), &buf)
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

	got, err := cli.Run([]string{"tnotes", "create"}, a, stdin, &buf)
	require.NoError(t, err)
	assert.True(t, got)

	// 出力にノートIDが含まれる
	id := strings.TrimSpace(buf.String())
	assert.Len(t, id, 16, "ノートIDは16文字のhex")

	// ノートが保存されていることを確認
	var getBuf bytes.Buffer

	_, _ = cli.Run([]string{"tnotes", "get", id}, a, strings.NewReader(""), &getBuf)
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

	got, createErr := cli.Run([]string{"tnotes", "create", tmpFile}, a, strings.NewReader(""), &buf)
	require.NoError(t, createErr)
	assert.True(t, got)

	id := strings.TrimSpace(buf.String())
	assert.Len(t, id, 16)

	var getBuf bytes.Buffer

	_, _ = cli.Run([]string{"tnotes", "get", id}, a, strings.NewReader(""), &getBuf)
	assert.Contains(t, getBuf.String(), "File note title")
	assert.Contains(t, getBuf.String(), "File body")
}

func TestRun_Create_EmptyInput(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "create"}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty input")
}

func TestRun_Create_FileNotFound(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "create", "/nonexistent/file.md"}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	assert.Error(t, err)
}

func TestRun_Export_MissingArg(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "export"}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, buf.String(), "Usage: tnotes export")
}

func TestRun_Export_Success(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now)
	_, _ = a.SaveNote(result.Note.ID, "Export test\nBody", now)

	outPath := filepath.Join(t.TempDir(), "backup.zip")

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "export", outPath}, a, strings.NewReader(""), &buf)
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
	result, _ := a.CreateNote(now)
	_, _ = a.SaveNote(result.Note.ID, "Trash test\nBody", now)
	_, _ = a.TrashNote(0)

	outPath := filepath.Join(t.TempDir(), "backup.zip")

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "export", outPath}, a, strings.NewReader(""), &buf)
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

	got, runErr := cli.Run([]string{"tnotes", "export", outPath}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, runErr)
}

func TestRun_Import_MissingArg(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "import"}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, buf.String(), "Usage: tnotes import")
}

func TestRun_Import_DataExists(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	now := time.Now()
	result, _ := a.CreateNote(now)
	_, _ = a.SaveNote(result.Note.ID, "Existing note\nBody", now)

	zipPath := filepath.Join(t.TempDir(), "import.zip")
	createEmptyZip(t, zipPath)

	var buf bytes.Buffer

	got, err := cli.Run([]string{"tnotes", "import", zipPath}, a, strings.NewReader(""), &buf)
	assert.True(t, got)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "data directory is not empty")
}

func TestRun_Import_Success(t *testing.T) {
	t.Parallel()

	// まずexportでzipを作成
	srcApp := newTestApp(t)
	now := time.Now()
	result, _ := srcApp.CreateNote(now)
	_, _ = srcApp.SaveNote(result.Note.ID, "Import test\nBody", now)

	zipPath := filepath.Join(t.TempDir(), "export.zip")

	var exportBuf bytes.Buffer

	_, exportErr := cli.Run([]string{"tnotes", "export", zipPath}, srcApp, strings.NewReader(""), &exportBuf)
	require.NoError(t, exportErr)

	// 空のAppにimport
	dstDir := t.TempDir()
	dstStore, _ := store.NewFileStore(dstDir)
	dstApp, _ := app.New(dstStore)

	var importBuf bytes.Buffer

	got, importErr := cli.Run([]string{"tnotes", "import", zipPath}, dstApp, strings.NewReader(""), &importBuf)
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
