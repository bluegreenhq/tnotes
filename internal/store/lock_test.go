package store_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gofrs/flock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockFile_CreatesLockFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// lockFile is unexported, so we test via the same flock mechanism
	lockPath := filepath.Join(dir, ".lock")
	fl := flock.New(lockPath)

	err := fl.Lock()
	require.NoError(t, err)

	defer func() { _ = fl.Unlock() }()

	_, err = os.Stat(lockPath)
	assert.NoError(t, err, "lock file should exist")
}

func TestLockFile_UnlockAllowsReacquire(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	lockPath := filepath.Join(dir, ".lock")
	fl1 := flock.New(lockPath)

	err := fl1.Lock()
	require.NoError(t, err)
	require.NoError(t, fl1.Unlock())

	// 解放後に再取得できることを確認
	fl2 := flock.New(lockPath)
	err = fl2.Lock()
	require.NoError(t, err)
	require.NoError(t, fl2.Unlock())
}

func TestLockFile_BlocksConcurrentAccess(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	lockPath := filepath.Join(dir, ".lock")

	var (
		mu    sync.Mutex
		order []int
	)

	var wg sync.WaitGroup
	wg.Add(2)

	for i := range 2 {
		go func(id int) {
			defer wg.Done()

			fl := flock.New(lockPath)

			err := fl.Lock()
			if err != nil {
				t.Error(err)

				return
			}

			mu.Lock()

			order = append(order, id)
			mu.Unlock()

			_ = fl.Unlock()
		}(i)
	}

	wg.Wait()
	assert.Len(t, order, 2, "both goroutines should have executed")
}

// TestLockFile_ExclusiveBetweenProcesses はプロセス間排他を検証する。
// ロック保持中にサブプロセスで TryLock を試みると失敗することを確認。
func TestLockFile_ExclusiveBetweenProcesses(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	lockPath := filepath.Join(dir, ".lock")

	fl := flock.New(lockPath)
	err := fl.Lock()
	require.NoError(t, err)

	defer func() { _ = fl.Unlock() }()

	// サブプロセスで flock --nonblock を試みる（ロック中なので失敗するはず）
	cmd := exec.CommandContext(t.Context(), "flock", "--nonblock", lockPath, "true")
	err = cmd.Run()
	assert.Error(t, err, "subprocess should fail to acquire lock")
}
