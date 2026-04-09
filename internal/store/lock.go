package store

import (
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/gofrs/flock"
)

// lockFile はデータディレクトリの排他ロックを取得する。
// 戻り値の関数を呼ぶとロックを解放しファイルを閉じる。
func lockFile(dir string) (func(), error) {
	lockPath := filepath.Join(dir, ".lock")
	fl := flock.New(lockPath)

	err := fl.Lock()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return func() {
		_ = fl.Unlock()
	}, nil
}
