package store

import (
	"time"

	"github.com/bluegreenhq/tnotes/internal/note"
)

// Store はノートの永続化インターフェースを定義する。
type Store interface {
	// List はインデックスからノート一覧を返す（Bodyは空）。
	List() ([]note.Note, error)
	// Load はノートのBodyを含む完全なNoteを返す。
	Load(id note.NoteID) (note.Note, error)
	// Save はノートをファイルに書き出し、インデックスを更新する。
	Save(n note.Note) error
	// Trash はノートをゴミ箱に移動する。
	Trash(id note.NoteID) error
	// ListTrashed はゴミ箱内のノート一覧を返す（Bodyは空）。
	ListTrashed() ([]note.Note, error)
	// Restore はゴミ箱からノートを復元する。
	Restore(id note.NoteID) error
	// PurgeTrash はゴミ箱内の全ノートを完全削除する。削除した件数を返す。
	PurgeTrash() (int, error)
	// DataDir はデータディレクトリのパスを返す。
	DataDir() string
	// IndexModTime はindex.jsonの最終更新日時を返す。
	IndexModTime() (time.Time, error)
	// Reload はindex.jsonを再読み込みしてインメモリ状態を更新する。
	Reload() error
}
