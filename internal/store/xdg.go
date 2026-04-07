package store

import (
	"os"
	"path/filepath"
)

const appName = "tnotes"

// DefaultDataDir はXDG Base Directory仕様に基づくデータディレクトリを返す。
func DefaultDataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, appName)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic("failed to get home directory: " + err.Error())
	}

	return filepath.Join(home, ".local", "share", appName)
}
