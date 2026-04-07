package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluegreenhq/tnotes/internal/store"
)

func TestDefaultDataDir_WithXDGEnv(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/custom/data")

	got := store.DefaultDataDir()
	assert.Equal(t, "/custom/data/tnotes", got)
}

func TestDefaultDataDir_WithoutXDGEnv(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	got := store.DefaultDataDir()
	assert.Equal(t, filepath.Join(home, ".local", "share", "tnotes"), got)
}
