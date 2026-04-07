package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/note"
)

var (
	ErrNoteNotFound        = errors.New("note not found")
	ErrTrashedNoteNotFound = errors.New("trashed note not found")
)

const dirPerm = 0o750

// FileStore はファイルシステムベースのStore実装。
type FileStore struct {
	dir        string
	index      map[note.NoteID]note.Metadata
	trashIndex map[note.NoteID]trashMetadata
}

// NewFileStore は指定ディレクトリでFileStoreを生成する。
// ディレクトリが存在しなければ作成する。index.jsonがあれば読み込む。
func NewFileStore(dir string) (*FileStore, error) {
	err := os.MkdirAll(dir, dirPerm)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fs := &FileStore{
		dir:        dir,
		index:      make(map[note.NoteID]note.Metadata),
		trashIndex: make(map[note.NoteID]trashMetadata),
	}

	err = fs.loadIndex()
	if err != nil {
		return nil, err
	}

	err = fs.loadTrashIndex()
	if err != nil {
		return nil, err
	}

	return fs, nil
}

// List はインデックスからノート一覧を返す（Bodyは空）。
func (fs *FileStore) List() ([]note.Note, error) {
	notes := make([]note.Note, 0, len(fs.index))
	for _, m := range fs.index {
		notes = append(notes, note.FromMetadata(m))
	}

	return notes, nil
}

// Load はノートファイルを読んでNoteを返す。
func (fs *FileStore) Load(id note.NoteID) (note.Note, error) {
	meta, ok := fs.index[id]
	if !ok {
		// ゴミ箱も確認
		trashMeta, ok2 := fs.trashIndex[id]
		if !ok2 {
			return note.Note{}, errors.WithDetail(ErrNoteNotFound, string(id))
		}

		meta = trashMeta.Metadata
	}

	path := filepath.Join(fs.dir, meta.Path)

	data, err := os.ReadFile(path)
	if err != nil {
		return note.Note{}, errors.WithStack(err)
	}

	return parseNoteFile(string(data))
}

// Save はノートをファイルに書き出し、インデックスを更新する。
func (fs *FileStore) Save(n note.Note) error {
	dateDir := n.CreatedAt.Format("20060102")
	relPath := filepath.Join(dateDir, string(n.ID)+".md")
	absPath := filepath.Join(fs.dir, relPath)

	// 既存ノートの場合、元のパスを維持
	if meta, ok := fs.index[n.ID]; ok {
		relPath = meta.Path
		absPath = filepath.Join(fs.dir, relPath)
	}

	// ディレクトリ作成
	err := os.MkdirAll(filepath.Dir(absPath), dirPerm)
	if err != nil {
		return errors.WithStack(err)
	}

	// ノートファイルをアトミック書き込み
	content := marshalNoteFile(n)

	err = atomicWrite(absPath, []byte(content))
	if err != nil {
		return errors.WithStack(err)
	}

	// インデックス更新
	fs.index[n.ID] = note.Metadata{
		ID:        n.ID,
		Title:     n.Title(),
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		Path:      relPath,
	}

	err = fs.saveIndex()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Trash はノートをゴミ箱に移動する。
func (fs *FileStore) Trash(id note.NoteID) error {
	meta, ok := fs.index[id]
	if !ok {
		return errors.WithDetail(ErrNoteNotFound, string(id))
	}

	srcPath := filepath.Join(fs.dir, meta.Path)
	dstRelPath := filepath.Join(".trash", meta.Path)
	dstPath := filepath.Join(fs.dir, dstRelPath)

	err := os.MkdirAll(filepath.Dir(dstPath), dirPerm)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Rename(srcPath, dstPath)
	if err != nil {
		return errors.WithStack(err)
	}

	// Update in-memory state
	trashMeta := trashMetadata{
		Metadata: note.Metadata{
			ID:        meta.ID,
			Title:     meta.Title,
			CreatedAt: meta.CreatedAt,
			UpdatedAt: meta.UpdatedAt,
			Path:      dstRelPath,
		},
		OriginalPath: meta.Path,
	}
	fs.trashIndex[id] = trashMeta
	delete(fs.index, id)

	// Save both indexes; roll back on failure
	err = fs.saveTrashIndex()
	if err != nil {
		// Roll back: restore in-memory state and move file back
		fs.index[id] = meta
		delete(fs.trashIndex, id)

		_ = os.Rename(dstPath, srcPath)

		return err
	}

	err = fs.saveIndex()
	if err != nil {
		// Roll back: restore in-memory state and move file back
		fs.index[id] = meta
		delete(fs.trashIndex, id)

		_ = os.Rename(dstPath, srcPath)
		_ = fs.saveTrashIndex() // best-effort rollback of trash index file

		return err
	}

	return nil
}

// ListTrashed はゴミ箱内のノート一覧を返す（Bodyは空）。
func (fs *FileStore) ListTrashed() ([]note.Note, error) {
	notes := make([]note.Note, 0, len(fs.trashIndex))
	for _, m := range fs.trashIndex {
		notes = append(notes, note.FromMetadata(m.Metadata))
	}

	return notes, nil
}

// Restore はゴミ箱からノートを復元する。
func (fs *FileStore) Restore(id note.NoteID) error {
	meta, ok := fs.trashIndex[id]
	if !ok {
		return errors.WithDetail(ErrTrashedNoteNotFound, string(id))
	}

	srcPath := filepath.Join(fs.dir, meta.Path)
	dstPath := filepath.Join(fs.dir, meta.OriginalPath)

	err := os.MkdirAll(filepath.Dir(dstPath), dirPerm)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Rename(srcPath, dstPath)
	if err != nil {
		return errors.WithStack(err)
	}

	// Update in-memory state
	restoredMeta := note.Metadata{
		ID:        meta.ID,
		Title:     meta.Title,
		CreatedAt: meta.CreatedAt,
		UpdatedAt: meta.UpdatedAt,
		Path:      meta.OriginalPath,
	}
	fs.index[id] = restoredMeta
	delete(fs.trashIndex, id)

	// Save both indexes; roll back on failure
	err = fs.saveIndex()
	if err != nil {
		// Roll back
		delete(fs.index, id)
		fs.trashIndex[id] = meta
		_ = os.Rename(dstPath, srcPath)

		return err
	}

	err = fs.saveTrashIndex()
	if err != nil {
		// Roll back
		delete(fs.index, id)
		fs.trashIndex[id] = meta
		_ = os.Rename(dstPath, srcPath)
		_ = fs.saveIndex()

		return err
	}

	return nil
}

// DataDir はデータディレクトリのパスを返す。
func (fs *FileStore) DataDir() string {
	return fs.dir
}

func (fs *FileStore) trashDir() string {
	return filepath.Join(fs.dir, ".trash")
}

func (fs *FileStore) loadIndex() error {
	path := filepath.Join(fs.dir, indexFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return errors.WithStack(err)
	}

	var idx indexData

	err = json.Unmarshal(data, &idx)
	if err != nil {
		return errors.WithStack(err)
	}

	for id, entry := range idx.Notes {
		nid := note.NoteID(id)
		createdAt, _ := parseTime(entry.CreatedAt)
		updatedAt, _ := parseTime(entry.UpdatedAt)
		fs.index[nid] = note.Metadata{
			ID:        nid,
			Title:     entry.Title,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Path:      entry.Path,
		}
	}

	return nil
}

func (fs *FileStore) saveIndex() error {
	idx := indexData{Notes: make(map[string]indexEntry, len(fs.index))}
	for id, meta := range fs.index {
		idx.Notes[string(id)] = indexEntry{
			Title:     meta.Title,
			CreatedAt: meta.CreatedAt.Format(timeFormat),
			UpdatedAt: meta.UpdatedAt.Format(timeFormat),
			Path:      meta.Path,
		}
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return errors.WithStack(err)
	}

	return atomicWrite(filepath.Join(fs.dir, indexFile), data)
}

func (fs *FileStore) loadTrashIndex() error {
	path := filepath.Join(fs.trashDir(), indexFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return errors.WithStack(err)
	}

	var idx trashIndexData

	err = json.Unmarshal(data, &idx)
	if err != nil {
		return errors.WithStack(err)
	}

	for id, entry := range idx.Notes {
		nid := note.NoteID(id)
		createdAt, _ := parseTime(entry.CreatedAt)
		updatedAt, _ := parseTime(entry.UpdatedAt)
		fs.trashIndex[nid] = trashMetadata{
			Metadata: note.Metadata{
				ID:        nid,
				Title:     entry.Title,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Path:      entry.Path,
			},
			OriginalPath: entry.OriginalPath,
		}
	}

	return nil
}

func (fs *FileStore) saveTrashIndex() error {
	trashDir := fs.trashDir()

	err := os.MkdirAll(trashDir, dirPerm)
	if err != nil {
		return errors.WithStack(err)
	}

	idx := trashIndexData{Notes: make(map[string]trashIndexEntry, len(fs.trashIndex))}
	for id, meta := range fs.trashIndex {
		idx.Notes[string(id)] = trashIndexEntry{
			Title:        meta.Title,
			CreatedAt:    meta.CreatedAt.Format(timeFormat),
			UpdatedAt:    meta.UpdatedAt.Format(timeFormat),
			Path:         meta.Path,
			OriginalPath: meta.OriginalPath,
		}
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return errors.WithStack(err)
	}

	return atomicWrite(filepath.Join(trashDir, indexFile), data)
}
