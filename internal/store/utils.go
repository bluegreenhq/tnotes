package store

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
)

// atomicWrite は一時ファイル経由でアトミックにファイルを書き込む。
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return errors.WithStack(err)
	}

	tmpName := tmp.Name()

	_, err = tmp.Write(data)
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)

		return errors.WithStack(err)
	}

	err = tmp.Close()
	if err != nil {
		_ = os.Remove(tmpName)

		return errors.WithStack(err)
	}

	err = os.Rename(tmpName, path)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
