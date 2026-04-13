package app_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/app"
	"github.com/bluegreenhq/tnotes/internal/store"
)

func TestSearchByFolder_MatchesTitle(t *testing.T) {
	t.Parallel()

	a := newSearchTestApp(t, []testNote{
		{body: "Meeting Notes\nDiscuss Q2 plan", folder: "Notes"},
		{body: "Shopping List\nMilk, Eggs", folder: "Notes"},
	})

	results := a.SearchByFolder("Notes", "meeting")
	assert.Len(t, results, 1)
	assert.Equal(t, "Meeting Notes", results[0].Title())
}

func TestSearchByFolder_MatchesBody(t *testing.T) {
	t.Parallel()

	a := newSearchTestApp(t, []testNote{
		{body: "Meeting Notes\nDiscuss Q2 plan", folder: "Notes"},
		{body: "Shopping List\nMilk, Eggs", folder: "Notes"},
	})

	results := a.SearchByFolder("Notes", "eggs")
	assert.Len(t, results, 1)
	assert.Equal(t, "Shopping List", results[0].Title())
}

func TestSearchByFolder_CaseInsensitive(t *testing.T) {
	t.Parallel()

	a := newSearchTestApp(t, []testNote{
		{body: "Hello World\nSome content", folder: "Notes"},
	})

	results := a.SearchByFolder("Notes", "HELLO")
	assert.Len(t, results, 1)
}

func TestSearchByFolder_EmptyQuery(t *testing.T) {
	t.Parallel()

	a := newSearchTestApp(t, []testNote{
		{body: "Note A\nBody A", folder: "Notes"},
		{body: "Note B\nBody B", folder: "Notes"},
	})

	results := a.SearchByFolder("Notes", "")
	assert.Len(t, results, 2)
}

func TestSearchByFolder_FolderScope(t *testing.T) {
	t.Parallel()

	a := newSearchTestApp(t, []testNote{
		{body: "Work Note\nWork body", folder: "Work"},
		{body: "Personal Note\nPersonal body", folder: "Notes"},
	})

	results := a.SearchByFolder("Work", "note")
	assert.Len(t, results, 1)
	assert.Equal(t, "Work Note", results[0].Title())
}

func TestExtractSnippets(t *testing.T) {
	t.Parallel()

	body := "Hello World meeting notes for project alpha"
	snippets := app.ExtractSnippets(body, "meeting", 10)
	assert.Len(t, snippets, 1)
	assert.Contains(t, snippets[0], "meeting")
}

type testNote struct {
	body   string
	folder string
}

func newSearchTestApp(t *testing.T, notes []testNote) *app.App {
	t.Helper()

	dir := t.TempDir()
	s, err := store.NewFileStore(dir)
	require.NoError(t, err)

	a, err := app.New(s)
	require.NoError(t, err)

	now := time.Now()

	for _, tn := range notes {
		result, err := a.CreateNote(now, tn.folder)
		require.NoError(t, err)

		_, err = a.SaveNote(result.Note.ID, tn.body, now)
		require.NoError(t, err)
	}

	return a
}
