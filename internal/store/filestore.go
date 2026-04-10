package store

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/bluegreenhq/tnotes/internal/note"
)

var (
	ErrNoteNotFound        = errors.New("note not found")
	ErrTrashedNoteNotFound = errors.New("trashed note not found")
	ErrDataDirNotEmpty     = errors.New("data directory is not empty")
	ErrInvalidZipPath      = errors.New("invalid path in zip")
)

const (
	dirPerm  = 0o750
	notesDir = "Notes"
	trashDir = ".trash"
)

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

	err = os.MkdirAll(fs.notesDirPath(), dirPerm)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = fs.loadIndex()
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
	relPath := filepath.Join(notesDir, dateDir, string(n.ID)+".md")
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

	// ロック取得 → index再読み込み → マージ → 書き込み → ロック解放
	unlock, err := lockFile(fs.dir)
	if err != nil {
		return err
	}
	defer unlock()

	err = fs.loadIndex()
	if err != nil {
		return err
	}

	fs.index[n.ID] = note.Metadata{
		ID:        n.ID,
		Title:     n.Title(),
		Preview:   n.Preview(),
		Pinned:    n.Pinned,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		Path:      relPath,
	}

	err = fs.saveIndex()
	if err != nil {
		return err
	}

	return nil
}

// Trash はノートをゴミ箱に移動する。
func (fs *FileStore) Trash(id note.NoteID) error {
	unlock, err := fs.lockAndReload()
	if err != nil {
		return err
	}
	defer unlock()

	meta, ok := fs.index[id]
	if !ok {
		return errors.WithDetail(ErrNoteNotFound, string(id))
	}

	srcPath := filepath.Join(fs.dir, meta.Path)
	pathWithoutNotes := strings.TrimPrefix(meta.Path, notesDir+string(filepath.Separator))
	dstRelPath := filepath.Join(trashDir, pathWithoutNotes)
	dstPath := filepath.Join(fs.dir, dstRelPath)

	err = os.MkdirAll(filepath.Dir(dstPath), dirPerm)
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
			Preview:   meta.Preview,
			Pinned:    meta.Pinned,
			CreatedAt: meta.CreatedAt,
			UpdatedAt: meta.UpdatedAt,
			Path:      dstRelPath,
		},
		OriginalPath: meta.Path,
	}
	fs.trashIndex[id] = trashMeta
	delete(fs.index, id)

	// Save unified index; roll back on failure
	err = fs.saveIndex()
	if err != nil {
		fs.index[id] = meta
		delete(fs.trashIndex, id)

		_ = os.Rename(dstPath, srcPath)

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
	unlock, err := fs.lockAndReload()
	if err != nil {
		return err
	}
	defer unlock()

	meta, ok := fs.trashIndex[id]
	if !ok {
		return errors.WithDetail(ErrTrashedNoteNotFound, string(id))
	}

	srcPath := filepath.Join(fs.dir, meta.Path)
	dstPath := filepath.Join(fs.dir, meta.OriginalPath)

	err = os.MkdirAll(filepath.Dir(dstPath), dirPerm)
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
		Preview:   meta.Preview,
		Pinned:    meta.Pinned,
		CreatedAt: meta.CreatedAt,
		UpdatedAt: meta.UpdatedAt,
		Path:      meta.OriginalPath,
	}
	fs.index[id] = restoredMeta
	delete(fs.trashIndex, id)

	// Save unified index; roll back on failure
	err = fs.saveIndex()
	if err != nil {
		delete(fs.index, id)
		fs.trashIndex[id] = meta
		_ = os.Rename(dstPath, srcPath)

		return err
	}

	return nil
}

// PurgeTrash はゴミ箱内の全ノートを完全削除する。削除した件数を返す。
func (fs *FileStore) PurgeTrash() (int, error) {
	unlock, err := fs.lockAndReload()
	if err != nil {
		return 0, err
	}
	defer unlock()

	count := len(fs.trashIndex)
	if count == 0 {
		return 0, nil
	}

	// ゴミ箱内の各ノートファイルを削除
	for _, meta := range fs.trashIndex {
		path := filepath.Join(fs.dir, meta.Path)

		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return 0, errors.WithStack(err)
		}
	}

	// trashIndex をクリアして保存
	clear(fs.trashIndex)

	err = fs.saveIndex()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// DataDir はデータディレクトリのパスを返す。
func (fs *FileStore) DataDir() string {
	return fs.dir
}

// IndexModTime はindex.jsonの最終更新日時を返す。
func (fs *FileStore) IndexModTime() (time.Time, error) {
	path := filepath.Join(fs.dir, IndexFile)

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return time.Time{}, nil
	}

	if err != nil {
		return time.Time{}, errors.WithStack(err)
	}

	return info.ModTime(), nil
}

// Reload はindex.jsonを再読み込みしてインメモリ状態を更新する。
func (fs *FileStore) Reload() error {
	return fs.loadIndex()
}

// Export はデータディレクトリの内容を指定された io.Writer に zip 形式で書き出す。
func (fs *FileStore) Export(w io.Writer) error {
	zw := zip.NewWriter(w)

	err := filepath.WalkDir(fs.dir, func(path string, d iofs.DirEntry, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}

		if d.IsDir() {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".tmp-") {
			return nil
		}

		return fs.addFileToZip(zw, path)
	})
	if err != nil {
		return errors.WithStack(err)
	}

	err = zw.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Import は zip 形式のデータを io.Reader から読み込み、データディレクトリに展開する。
// データディレクトリが空でない場合はエラーを返す。
func (fs *FileStore) Import(r io.Reader) error {
	if fs.HasData() {
		return ErrDataDirNotEmpty
	}

	zr, err := readZipFromReader(r)
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if strings.Contains(f.Name, "..") {
			return errors.WithDetail(ErrInvalidZipPath, f.Name)
		}

		destPath := filepath.Join(fs.dir, filepath.Clean(f.Name))

		mkErr := os.MkdirAll(filepath.Dir(destPath), dirPerm)
		if mkErr != nil {
			return errors.WithStack(mkErr)
		}

		writeErr := extractZipFile(f, destPath)
		if writeErr != nil {
			return errors.WithStack(writeErr)
		}
	}

	return fs.loadIndex()
}

// HasData はデータが存在するかを返す（index.json の存在チェック）。
func (fs *FileStore) HasData() bool {
	_, err := os.Stat(filepath.Join(fs.dir, IndexFile))

	return err == nil
}

// lockAndReload はロックを取得し、indexを再読み込みする。
// 戻り値の関数を呼ぶとロックを解放する。
func (fs *FileStore) lockAndReload() (func(), error) {
	unlock, err := lockFile(fs.dir)
	if err != nil {
		return nil, err
	}

	err = fs.loadIndex()
	if err != nil {
		unlock()

		return nil, err
	}

	return unlock, nil
}

func (fs *FileStore) notesDirPath() string {
	return filepath.Join(fs.dir, notesDir)
}

func (fs *FileStore) loadIndex() error {
	path := filepath.Join(fs.dir, IndexFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		clear(fs.index)
		clear(fs.trashIndex)

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

	clear(fs.index)

	for id, entry := range idx.Notes {
		nid := note.NoteID(id)
		createdAt, _ := parseTime(entry.CreatedAt)
		updatedAt, _ := parseTime(entry.UpdatedAt)
		fs.index[nid] = note.Metadata{
			ID:        nid,
			Title:     entry.Title,
			Preview:   entry.Preview,
			Pinned:    entry.Pinned,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Path:      entry.Path,
		}
	}

	clear(fs.trashIndex)

	for id, entry := range idx.Trash {
		nid := note.NoteID(id)
		createdAt, _ := parseTime(entry.CreatedAt)
		updatedAt, _ := parseTime(entry.UpdatedAt)
		fs.trashIndex[nid] = trashMetadata{
			Metadata: note.Metadata{
				ID:        nid,
				Title:     entry.Title,
				Preview:   entry.Preview,
				Pinned:    entry.Pinned,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				Path:      entry.Path,
			},
			OriginalPath: entry.OriginalPath,
		}
	}

	return nil
}

func (fs *FileStore) saveIndex() error {
	idx := indexData{
		Notes: make(map[string]indexEntry, len(fs.index)),
		Trash: make(map[string]trashIndexEntry, len(fs.trashIndex)),
	}

	for id, meta := range fs.index {
		idx.Notes[string(id)] = indexEntry{
			Title:     meta.Title,
			Preview:   meta.Preview,
			Pinned:    meta.Pinned,
			CreatedAt: meta.CreatedAt.Format(timeFormat),
			UpdatedAt: meta.UpdatedAt.Format(timeFormat),
			Path:      meta.Path,
		}
	}

	for id, meta := range fs.trashIndex {
		idx.Trash[string(id)] = trashIndexEntry{
			Title:        meta.Title,
			Preview:      meta.Preview,
			Pinned:       meta.Pinned,
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

	return atomicWrite(filepath.Join(fs.dir, IndexFile), data)
}

func (fs *FileStore) addFileToZip(zw *zip.Writer, path string) error {
	relPath, err := filepath.Rel(fs.dir, path)
	if err != nil {
		return errors.WithStack(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return errors.WithStack(err)
	}

	zf, err := zw.Create(relPath)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = zf.Write(data)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func readZipFromReader(r io.Reader) (*zip.Reader, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return zr, nil
}

func extractZipFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return errors.WithStack(err)
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer out.Close()

	_, err = io.Copy(out, io.LimitReader(rc, int64(f.UncompressedSize64))) //nolint:gosec // zip内ファイルサイズはユーザー管理のバックアップデータ
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
