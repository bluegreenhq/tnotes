package app

import (
	"time"

	"github.com/bluegreenhq/tnotes/internal/note"
	"github.com/bluegreenhq/tnotes/internal/store"
)

// App はノート管理のアプリケーションロジックを提供する。
type App struct {
	Notes      []note.Note
	TrashNotes []note.Note
	store      store.Store
	TrashMode  bool
	NoteUndo   NoteUndoManager
}

// New は App を生成し、ストアからノート一覧を読み込む。
func New(s store.Store) (*App, error) {
	a := &App{
		Notes: nil, TrashNotes: nil, store: s, TrashMode: false,
		NoteUndo: NoteUndoManager{undoStack: nil, redoStack: nil},
	}
	if s == nil {
		return a, nil
	}

	var err error

	a.Notes, err = s.List()
	if err != nil {
		return a, err
	}

	note.SortByUpdatedDesc(a.Notes)

	return a, nil
}

// CreateNote は新しいノートを作成し、リストの先頭に追加する。
// undo記録も内部で行う。
func (a *App) CreateNote(now time.Time) (NoteResult, error) {
	n, err := note.New(now)
	if err != nil {
		return NoteResult{}, err
	}

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

// TrashNote は指定インデックスのノートをゴミ箱に移動する。
// undo記録も内部で行う。
func (a *App) TrashNote(idx int) (NoteResult, error) {
	if idx < 0 || idx >= len(a.Notes) {
		return NoteResult{Notes: a.Notes, SelectIdx: -1}, nil //nolint:exhaustruct // 範囲外は操作なし
	}

	noteID := a.Notes[idx].ID

	selectIdx, err := a.trashNoteInternal(idx)
	if err != nil {
		return NoteResult{}, err
	}

	a.NoteUndo.Push(&TrashAction{NoteID: noteID, OriginalIndex: idx})

	return NoteResult{Notes: a.Notes, SelectIdx: selectIdx, InfoHint: "Undo: Ctrl+Z"}, nil //nolint:exhaustruct // Noteはゴミ箱移動で不要
}

// RestoreNote は指定インデックスのゴミ箱ノートを復元する。
// undo記録も内部で行う。
func (a *App) RestoreNote(idx int) (NoteResult, error) {
	n, restoredIdx, err := a.restoreNoteInternal(idx)
	if err != nil {
		return NoteResult{}, err
	}

	a.NoteUndo.Push(&RestoreAction{NoteID: n.ID})

	return NoteResult{
		Note:      n,
		Notes:     a.Notes,
		SelectIdx: restoredIdx,
		InfoHint:  "Undo: Ctrl+Z",
	}, nil
}

// EnterTrashMode はゴミ箱モードに切り替える。
func (a *App) EnterTrashMode() error {
	a.TrashMode = true

	if a.store != nil {
		var err error

		a.TrashNotes, err = a.store.ListTrashed()
		if err != nil {
			return err
		}

		note.SortByUpdatedDesc(a.TrashNotes)
	}

	return nil
}

// ExitTrashMode は通常モードに戻る。
func (a *App) ExitTrashMode() {
	a.TrashMode = false
}

// PurgeTrash はゴミ箱内の全ノートを完全削除する。削除した件数を返す。
func (a *App) PurgeTrash() (int, error) {
	if a.store == nil {
		count := len(a.TrashNotes)
		a.TrashNotes = nil

		return count, nil
	}

	count, err := a.store.PurgeTrash()
	if err != nil {
		return 0, err
	}

	a.TrashNotes = nil

	return count, nil
}

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

	a.Notes, err = a.store.List()
	if err != nil {
		return false, mt, err
	}

	note.SortByUpdatedDesc(a.Notes)

	return true, mt, nil
}

// IndexModTime はindex.jsonの最終更新日時を返す。
func (a *App) IndexModTime() (time.Time, error) {
	if a.store == nil {
		return time.Time{}, nil
	}

	return a.store.IndexModTime()
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

// trashNoteInternal はundo記録なしでノートをゴミ箱に移動する。
// undo/redoアクション実行用。
// 戻り値は次に選択すべきインデックス（ノートが空なら -1）。
func (a *App) trashNoteInternal(idx int) (int, error) {
	n := a.Notes[idx]
	if a.store != nil {
		err := a.store.Trash(n.ID)
		if err != nil {
			return -1, err
		}
	}

	a.TrashNotes = append([]note.Note{n}, a.TrashNotes...)
	a.Notes = append(a.Notes[:idx], a.Notes[idx+1:]...)

	if len(a.Notes) == 0 {
		return -1, nil
	}

	if idx >= len(a.Notes) {
		return len(a.Notes) - 1, nil
	}

	return idx, nil
}

// restoreNoteInternal はundo記録なしでノートを復元する。
// undo/redoアクション実行用。
// 戻り値は復元したノートとメインリストでのインデックス。
func (a *App) restoreNoteInternal(idx int) (note.Note, int, error) {
	n := a.TrashNotes[idx]

	if a.store != nil {
		err := a.store.Restore(n.ID)
		if err != nil {
			return note.Note{}, 0, err
		}
	}

	a.TrashNotes = append(a.TrashNotes[:idx], a.TrashNotes[idx+1:]...)
	a.Notes = append([]note.Note{n}, a.Notes...)
	note.SortByUpdatedDesc(a.Notes)
	a.TrashMode = false

	for i, nn := range a.Notes {
		if nn.ID == n.ID {
			return n, i, nil
		}
	}

	return n, 0, nil
}

func (a *App) updateNoteBody(id note.NoteID, body string) {
	target := a.Notes
	if a.TrashMode {
		target = a.TrashNotes
	}

	for i := range target {
		if target[i].ID == id {
			target[i].Body = body

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

func (a *App) findTrashNoteIndex(id note.NoteID) int {
	for i, n := range a.TrashNotes {
		if n.ID == id {
			return i
		}
	}

	return -1
}
