package app

import (
	"io"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/store"
)

var (
	// ErrNoteNotFound はノートが見つからない場合のエラー。
	ErrNoteNotFound = errors.New("note not found")
	// ErrFolderNotFound はフォルダが見つからない場合のエラー。
	ErrFolderNotFound = errors.New("folder not found")
)

// App はノート管理のアプリケーションロジックを提供する。
// Notes は全ノート（通常 + ゴミ箱）を保持する。
type App struct {
	Notes    []note.Note
	store    store.Store
	NoteUndo NoteUndoManager
}

// New は App を生成し、ストアからノート一覧を読み込む。
func New(s store.Store) (*App, error) {
	a := &App{
		Notes: nil, store: s,
		NoteUndo: NoteUndoManager{undoStack: nil, redoStack: nil},
	}
	if s == nil {
		return a, nil
	}

	err := a.loadAllNotes()
	if err != nil {
		return a, err
	}

	return a, nil
}

// ListNotes は通常ノート一覧を返す（Trash を除外）。
func (a *App) ListNotes() []note.Note {
	result := make([]note.Note, 0, len(a.Notes))

	for _, n := range a.Notes {
		if !n.IsTrash() {
			result = append(result, n)
		}
	}

	return result
}

// ListTrashNotes はゴミ箱ノート一覧を返す。
func (a *App) ListTrashNotes() []note.Note {
	result := make([]note.Note, 0)

	for _, n := range a.Notes {
		if n.IsTrash() {
			result = append(result, n)
		}
	}

	return result
}

// ListTrash はゴミ箱ノート一覧をストアから読み込んで返す。
func (a *App) ListTrash() ([]note.Note, error) {
	return a.ListTrashNotes(), nil
}

// CreateNote は新しいノートを作成し、リストの先頭に追加する。
// folder が空の場合はデフォルトフォルダ（Notes）に保存する。
// undo記録も内部で行う。
func (a *App) CreateNote(now time.Time, folder string) (NoteResult, error) {
	n, err := note.New(now)
	if err != nil {
		return NoteResult{}, err
	}

	if folder == "" {
		folder = DefaultFolder
	}

	n.Path = store.NotePath(folder, now, n.ID)

	a.Notes = append([]note.Note{n}, a.Notes...)

	if a.store != nil {
		err := a.store.Save(n)
		if err != nil {
			return NoteResult{}, err
		}
	}

	a.NoteUndo.Push(&CreateAction{NoteID: n.ID})

	return NoteResult{
		Note:      n,
		Notes:     a.Notes,
		SelectIdx: 0,
		InfoHint:  "Undo: Ctrl+Z",
	}, nil
}

// DuplicateNote は指定IDのノートを複製する。
// 複製されたノートはコピー元の直下に挿入される。
// undo記録も内部で行う。
func (a *App) DuplicateNote(id note.NoteID) (NoteResult, error) {
	srcIdx := a.findNoteIndex(id)
	if srcIdx < 0 {
		return NoteResult{}, ErrNoteNotFound
	}

	dup, err := a.store.Duplicate(id)
	if err != nil {
		return NoteResult{}, err
	}

	insertIdx := srcIdx + 1
	a.Notes = slices.Insert(a.Notes, insertIdx, dup)

	a.NoteUndo.Push(&DuplicateAction{NoteID: dup.ID, note: dup})

	return NoteResult{
		Note:      dup,
		Notes:     a.Notes,
		SelectIdx: insertIdx,
		InfoHint:  "Undo: Ctrl+Z",
	}, nil
}

// TrashNote は指定IDのノートをゴミ箱に移動する。
// undo記録も内部で行う。
func (a *App) TrashNote(id note.NoteID) (NoteResult, error) {
	idx := a.findNoteIndex(id)
	if idx < 0 {
		return NoteResult{}, ErrNoteNotFound
	}

	originalFolder := a.Notes[idx].Folder()

	err := a.trashNoteInternal(id)
	if err != nil {
		return NoteResult{}, err
	}

	a.NoteUndo.Push(&TrashAction{NoteID: id, OriginalFolder: originalFolder})

	return NoteResult{Notes: a.Notes, SelectIdx: -1, InfoHint: "Undo: Ctrl+Z"}, nil //nolint:exhaustruct // Noteはゴミ箱移動で不要
}

// RefreshTrashNotes はストアからノートを再読み込みする（ゴミ箱表示切替時）。
func (a *App) RefreshTrashNotes() error {
	if a.store == nil {
		return nil
	}

	return a.loadAllNotes()
}

// PurgeTrash はゴミ箱内の全ノートを完全削除する。削除した件数を返す。
func (a *App) PurgeTrash() (int, error) {
	if a.store == nil {
		trashNotes := a.ListTrashNotes()
		count := len(trashNotes)

		a.Notes = a.ListNotes()

		return count, nil
	}

	count, err := a.store.PurgeTrash()
	if err != nil {
		return 0, err
	}

	// インメモリからも除去
	a.Notes = a.ListNotes()

	return count, nil
}

// Export はデータディレクトリの内容を zip 形式で書き出す。
func (a *App) Export(w io.Writer) error { return a.store.Export(w) }

// Import は zip 形式のデータを読み込み、データディレクトリに展開する。
func (a *App) Import(r io.Reader) error { return a.store.Import(r) }

// HasData はデータが存在するかを返す。
func (a *App) HasData() bool { return a.store.HasData() }

// DataDir はデータディレクトリのパスを返す。
func (a *App) DataDir() string {
	return a.store.DataDir()
}

// RefreshNotes はindex.jsonが更新されていればノート一覧を再読み込みする。
// 更新があった場合は true を返す。
func (a *App) RefreshNotes(lastModTime time.Time) (bool, time.Time, error) {
	if a.store == nil {
		return false, lastModTime, nil
	}

	mt, err := a.store.IndexModTime()
	if err != nil {
		return false, lastModTime, err
	}

	if !mt.After(lastModTime) {
		return false, lastModTime, nil
	}

	err = a.store.Reload()
	if err != nil {
		return false, lastModTime, err
	}

	err = a.loadAllNotes()
	if err != nil {
		return false, mt, err
	}

	return true, mt, nil
}

// IndexModTime はindex.jsonの最終更新日時を返す。
func (a *App) IndexModTime() (time.Time, error) {
	if a.store == nil {
		return time.Time{}, nil
	}

	return a.store.IndexModTime()
}

// GetNote は指定IDのノートを検索し、本文を含むNoteを返す。
func (a *App) GetNote(id note.NoteID) (note.Note, error) {
	for _, n := range a.Notes {
		if n.ID == id {
			return a.LoadNote(n)
		}
	}

	return note.Note{}, ErrNoteNotFound
}

// LoadNote はノートの本文をストアから読み込む。
// 既にBodyがある場合やストアがnilの場合はそのまま返す。
func (a *App) LoadNote(n note.Note) (note.Note, error) {
	if n.Body != "" || a.store == nil {
		return n, nil
	}

	loaded, err := a.store.Load(n.ID)
	if err != nil {
		return n, err
	}

	a.updateNoteBody(n.ID, loaded.Body)

	return loaded, nil
}

// SaveNote はノートの本文を更新し、ストアに保存する。
// 戻り値はソート後のノートのインデックス。
func (a *App) SaveNote(id note.NoteID, body string, now time.Time) (int, error) {
	for i := range a.Notes {
		if a.Notes[i].ID == id {
			a.Notes[i].Body = body
			a.Notes[i].UpdatedAt = now

			if a.store != nil {
				err := a.store.Save(a.Notes[i])
				if err != nil {
					return i, err
				}
			}

			note.SortByUpdatedDesc(a.Notes)

			return 0, nil
		}
	}

	return 0, nil
}

// DiscardIfEmpty は Body が空のノートを完全削除する。
// 削除した場合は true を返す。undo には積まない。
func (a *App) DiscardIfEmpty(id note.NoteID) bool {
	idx := a.findNoteIndex(id)
	if idx < 0 {
		return false
	}

	if a.Notes[idx].Body != "" {
		return false
	}

	if a.store != nil {
		_ = a.store.Delete(id)
	}

	a.Notes = slices.Delete(a.Notes, idx, idx+1)

	return true
}

// PinNote は指定IDのノートをピン留めする。
func (a *App) PinNote(id note.NoteID) error {
	return a.setPinned(id, true)
}

// UnpinNote は指定IDのノートのピン留めを解除する。
func (a *App) UnpinNote(id note.NoteID) error {
	return a.setPinned(id, false)
}

// ListFolders はユーザー定義フォルダ名一覧を返す。
func (a *App) ListFolders() ([]string, error) {
	if a.store == nil {
		return nil, nil
	}

	return a.store.ListFolders()
}

// CreateFolder はユーザー定義フォルダを作成する。
func (a *App) CreateFolder(name string) error {
	if a.store == nil {
		return nil
	}

	return a.store.CreateFolder(name)
}

// DeleteFolder はユーザー定義フォルダを削除する。
// フォルダ内のノートをゴミ箱に移動してからディレクトリを削除する。
// 戻り値はゴミ箱に移動したノート件数。
func (a *App) DeleteFolder(name string) (int, error) {
	if a.store == nil {
		return 0, nil
	}

	count := 0

	for _, n := range a.ListByFolder(name) {
		err := a.trashNoteInternal(n.ID)
		if err != nil {
			return count, err
		}

		count++
	}

	err := a.store.DeleteFolder(name)
	if err != nil {
		return count, err
	}

	return count, nil
}

// RenameFolder はユーザー定義フォルダをリネームする。
// フォルダ内のノートのパスも更新する。
func (a *App) RenameFolder(oldName, newName string) error {
	if a.store == nil {
		return nil
	}

	err := a.store.RenameFolder(oldName, newName)
	if err != nil {
		return err
	}

	// インメモリのノートパスを更新
	oldPrefix := oldName + string(filepath.Separator)
	newPrefix := newName + string(filepath.Separator)

	for i := range a.Notes {
		if after, ok := strings.CutPrefix(a.Notes[i].Path, oldPrefix); ok {
			a.Notes[i].Path = newPrefix + after
		}
	}

	return nil
}

// MoveNoteToFolder は指定IDのノートを別のフォルダに移動する。
func (a *App) MoveNoteToFolder(id note.NoteID, destFolder string) error {
	idx := a.findNoteIndex(id)
	if idx < 0 {
		return ErrNoteNotFound
	}

	if a.store != nil {
		err := a.store.MoveNote(id, destFolder)
		if err != nil {
			return err
		}
	}

	oldPath := a.Notes[idx].Path

	parts := strings.SplitN(oldPath, string(filepath.Separator), pathSplitParts)
	if len(parts) >= pathSplitParts {
		a.Notes[idx].Path = filepath.Join(destFolder, parts[1])
	}

	return nil
}

// FolderNoteCount は指定フォルダのノート件数を返す。
func (a *App) FolderNoteCount(name string) (int, error) {
	count := 0

	for _, n := range a.Notes {
		if n.Folder() == name {
			count++
		}
	}

	return count, nil
}

// DefaultFolder はデフォルトのノートフォルダ名。
const DefaultFolder = "Notes"

// ListByFolder は指定フォルダに属するノート一覧を返す。
func (a *App) ListByFolder(folderName string) []note.Note {
	filtered := make([]note.Note, 0, len(a.Notes))

	for _, n := range a.Notes {
		if n.Folder() == folderName {
			filtered = append(filtered, n)
		}
	}

	return filtered
}

// FolderExists は指定名のフォルダが存在するかを返す。
func (a *App) FolderExists(name string) (bool, error) {
	if name == DefaultFolder {
		return true, nil
	}

	folders, err := a.ListFolders()
	if err != nil {
		return false, err
	}

	return slices.Contains(folders, name), nil
}

func (a *App) setPinned(id note.NoteID, pinned bool) error {
	idx := a.findNoteIndex(id)
	if idx < 0 {
		return ErrNoteNotFound
	}

	a.Notes[idx].Pinned = pinned

	if a.store != nil {
		n, err := a.store.Load(id)
		if err != nil {
			return err
		}

		n.Pinned = pinned

		return a.store.Save(n)
	}

	return nil
}

const pathSplitParts = 2

// trashNoteInternal はundo記録なしでノートをゴミ箱に移動する。
func (a *App) trashNoteInternal(id note.NoteID) error {
	if a.store != nil {
		err := a.store.Trash(id)
		if err != nil {
			return err
		}
	}

	// インメモリの Path を .trash/ プレフィックスに更新
	for i := range a.Notes {
		if a.Notes[i].ID == id {
			oldPath := a.Notes[i].Path
			pathWithoutFolder := strings.TrimPrefix(oldPath, a.Notes[i].Folder()+string(filepath.Separator))
			a.Notes[i].Path = filepath.Join(note.TrashDir, pathWithoutFolder)

			break
		}
	}

	return nil
}

func (a *App) updateNoteBody(id note.NoteID, body string) {
	for i := range a.Notes {
		if a.Notes[i].ID == id {
			a.Notes[i].Body = body

			return
		}
	}
}

func (a *App) findNoteIndex(id note.NoteID) int {
	for i, n := range a.Notes {
		if n.ID == id {
			return i
		}
	}

	return -1
}

func (a *App) loadAllNotes() error {
	list, err := a.store.List()
	if err != nil {
		return err
	}

	trashed, err := a.store.ListTrashed()
	if err != nil {
		return err
	}

	all := make([]note.Note, 0, len(list)+len(trashed))
	all = append(all, list...)
	all = append(all, trashed...)
	a.Notes = all
	note.SortByUpdatedDesc(a.Notes)

	return nil
}
