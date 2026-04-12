package store

import (
	"io"
	"time"

	"github.com/bluegreenhq/tnotes/internal/note"
)

// Store はノートの永続化インターフェースを定義する。
type Store interface { //nolint:interfacebloat // import/exportの責務もStoreに含める
	// List はインデックスからノート一覧を返す（Bodyは空）。
	List() ([]note.Note, error)
	// Load はノートのBodyを含む完全なNoteを返す。
	Load(id note.NoteID) (note.Note, error)
	// Save はノートをファイルに書き出し、インデックスを更新する。
	Save(n note.Note) error
	// Delete はノートを完全に削除する（ゴミ箱には移動しない）。
	Delete(id note.NoteID) error
	// Trash はノートをゴミ箱に移動する。
	Trash(id note.NoteID) error
	// ListTrashed はゴミ箱内のノート一覧を返す（Bodyは空）。
	ListTrashed() ([]note.Note, error)
	// PurgeTrash はゴミ箱内の全ノートを完全削除する。削除した件数を返す。
	PurgeTrash() (int, error)
	// DataDir はデータディレクトリのパスを返す。
	DataDir() string
	// IndexModTime はindex.jsonの最終更新日時を返す。
	IndexModTime() (time.Time, error)
	// Reload はindex.jsonを再読み込みしてインメモリ状態を更新する。
	Reload() error
	// Export はデータディレクトリの内容を指定された io.Writer に zip 形式で書き出す。
	Export(w io.Writer) error
	// Import は zip 形式のデータを io.Reader から読み込み、データディレクトリに展開する。
	// データディレクトリが空でない場合はエラーを返す。
	Import(r io.Reader) error
	// HasData はデータが存在するかを返す（index.json の存在チェック）。
	HasData() bool
	// ListFolders はユーザー定義フォルダ名一覧をアルファベット順で返す。
	ListFolders() ([]string, error)
	// CreateFolder はユーザー定義フォルダを作成する。
	CreateFolder(name string) error
	// DeleteFolder はユーザー定義フォルダを削除する（空ディレクトリのみ）。
	DeleteFolder(name string) error
	// RenameFolder はユーザー定義フォルダをリネームする。
	RenameFolder(oldName, newName string) error
	// MoveNote はノートを別のフォルダに移動する。
	MoveNote(id note.NoteID, destFolder string) error
}
