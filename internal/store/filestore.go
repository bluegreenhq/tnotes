package store

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"sort"
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
	ErrFolderAlreadyExists = errors.New("folder already exists")
	ErrFolderNotFound      = errors.New("folder not found")
	ErrSystemFolder        = errors.New("cannot operate on system folder")
)

const (
	dirPerm  = 0o750
	notesDir = "Notes"
)

// FileStore はファイルシステムベースのStore実装。
type FileStore struct {
	dir   string
	index map[note.NoteID]note.Metadata
}

// NewFileStore は指定ディレクトリでFileStoreを生成する。
// ディレクトリが存在しなければ作成する。index.jsonがあれば読み込む。
func NewFileStore(dir string) (*FileStore, error) {
	err := os.MkdirAll(dir, dirPerm)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fs := &FileStore{
		dir:   dir,
		index: make(map[note.NoteID]note.Metadata),
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

// isTrashPath は Path が .trash/ プレフィックスかを返す。
func isTrashPath(path string) bool {
	return strings.HasPrefix(path, note.TrashDir+string(filepath.Separator))
}

// List はインデックスから通常ノート一覧を返す（Bodyは空）。
// Path が .trash/ プレフィックスのエントリを除外する。
func (fs *FileStore) List() ([]note.Note, error) {
	notes := make([]note.Note, 0, len(fs.index))
	for _, m := range fs.index {
		if !isTrashPath(m.Path) {
			notes = append(notes, note.FromMetadata(m))
		}
	}

	return notes, nil
}

// Load はノートファイルを読んでNoteを返す。
func (fs *FileStore) Load(id note.NoteID) (note.Note, error) {
	meta, ok := fs.index[id]
	if !ok {
		return note.Note{}, errors.WithDetail(ErrNoteNotFound, string(id))
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
	var relPath string

	if meta, ok := fs.index[n.ID]; ok {
		// 既存ノートの場合、元のパスを維持
		relPath = meta.Path
	} else if n.Path != "" {
		// 新規ノートでApp層がパスを設定済みの場合
		relPath = n.Path
	} else {
		// フォールバック
		relPath = NotePath(notesDir, n.CreatedAt, n.ID)
	}

	absPath := filepath.Join(fs.dir, relPath)

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

// Delete はノートを完全に削除する（ゴミ箱には移動しない）。
func (fs *FileStore) Delete(id note.NoteID) error {
	unlock, err := fs.lockAndReload()
	if err != nil {
		return err
	}
	defer unlock()

	meta, ok := fs.index[id]
	if !ok {
		return errors.WithDetail(ErrNoteNotFound, string(id))
	}

	absPath := filepath.Join(fs.dir, meta.Path)

	err = os.Remove(absPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.WithStack(err)
	}

	delete(fs.index, id)

	err = fs.saveIndex()
	if err != nil {
		fs.index[id] = meta

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
	dstRelPath := filepath.Join(note.TrashDir, pathWithoutNotes)
	dstPath := filepath.Join(fs.dir, dstRelPath)

	err = os.MkdirAll(filepath.Dir(dstPath), dirPerm)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Rename(srcPath, dstPath)
	if err != nil {
		return errors.WithStack(err)
	}

	// Update in-memory state: Path を .trash/... に変更
	oldMeta := meta
	meta.Path = dstRelPath
	fs.index[id] = meta

	// Save index; roll back on failure
	err = fs.saveIndex()
	if err != nil {
		fs.index[id] = oldMeta
		_ = os.Rename(dstPath, srcPath)

		return err
	}

	return nil
}

// ListTrashed はゴミ箱内のノート一覧を返す（Bodyは空）。
// index から Path が .trash/ プレフィックスのエントリをフィルタして返す。
func (fs *FileStore) ListTrashed() ([]note.Note, error) {
	notes := make([]note.Note, 0)

	for _, m := range fs.index {
		if isTrashPath(m.Path) {
			notes = append(notes, note.FromMetadata(m))
		}
	}

	return notes, nil
}

// PurgeTrash はゴミ箱内の全ノートを完全削除する。削除した件数を返す。
func (fs *FileStore) PurgeTrash() (int, error) {
	unlock, err := fs.lockAndReload()
	if err != nil {
		return 0, err
	}
	defer unlock()

	// ゴミ箱ノートを収集
	trashIDs := make([]note.NoteID, 0)

	for id, m := range fs.index {
		if isTrashPath(m.Path) {
			trashIDs = append(trashIDs, id)
		}
	}

	count := len(trashIDs)
	if count == 0 {
		return 0, nil
	}

	// ゴミ箱内の各ノートファイルを削除
	for _, id := range trashIDs {
		path := filepath.Join(fs.dir, fs.index[id].Path)

		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return 0, errors.WithStack(err)
		}

		delete(fs.index, id)
	}

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

// ListFolders はデータDir直下のユーザー定義フォルダ名一覧をアルファベット順で返す。
// Notes, .trash, ドットディレクトリ、ファイルは除外する。
func (fs *FileStore) ListFolders() ([]string, error) {
	entries, err := os.ReadDir(fs.dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	folders := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		name := e.Name()
		if fs.isSystemDir(name) {
			continue
		}

		folders = append(folders, name)
	}

	sort.Strings(folders)

	return folders, nil
}

// CreateFolder はユーザー定義フォルダを作成する。
func (fs *FileStore) CreateFolder(name string) error {
	if fs.isSystemDir(name) {
		return errors.WithDetail(ErrSystemFolder, name)
	}

	path := filepath.Join(fs.dir, name)

	_, err := os.Stat(path)
	if err == nil {
		return errors.WithDetail(ErrFolderAlreadyExists, name)
	}

	if !os.IsNotExist(err) {
		return errors.WithStack(err)
	}

	return errors.WithStack(os.MkdirAll(path, dirPerm))
}

// DeleteFolder はユーザー定義フォルダを削除する。
// 空サブディレクトリを再帰的に削除してからフォルダ自体を削除する。
// 中のノートの移動は呼び出し側の責務。
func (fs *FileStore) DeleteFolder(name string) error {
	if fs.isSystemDir(name) {
		return errors.WithDetail(ErrSystemFolder, name)
	}

	path := filepath.Join(fs.dir, name)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.WithDetail(ErrFolderNotFound, name)
	}

	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(removeEmptyDirs(path))
}

// RenameFolder はユーザー定義フォルダをリネームする。
// ディレクトリのリネームとindex内のノートパスを更新する。
func (fs *FileStore) RenameFolder(oldName, newName string) error {
	oldPath, newPath, err := fs.validateRename(oldName, newName)
	if err != nil {
		return err
	}

	err = os.Rename(oldPath, newPath)
	if err != nil {
		return errors.WithStack(err)
	}

	// ロック取得 → index再読み込み → パス更新 → 書き込み → ロック解放
	unlock, err := lockFile(fs.dir)
	if err != nil {
		_ = os.Rename(newPath, oldPath)

		return err
	}
	defer unlock()

	err = fs.loadIndex()
	if err != nil {
		_ = os.Rename(newPath, oldPath)

		return err
	}

	oldPrefix := oldName + string(filepath.Separator)
	newPrefix := newName + string(filepath.Separator)

	for id, meta := range fs.index {
		if after, ok := strings.CutPrefix(meta.Path, oldPrefix); ok {
			meta.Path = newPrefix + after
			fs.index[id] = meta
		}
	}

	err = fs.saveIndex()
	if err != nil {
		_ = os.Rename(newPath, oldPath)

		return err
	}

	return nil
}

// MoveNote はノートを別のフォルダに移動する。
// ファイルの移動とindex内のパスを更新する。
// Trash 内ノート（.trash/ パス）も移動可能。
func (fs *FileStore) MoveNote(id note.NoteID, destFolder string) error {
	unlock, err := fs.lockAndReload()
	if err != nil {
		return err
	}
	defer unlock()

	meta, ok := fs.index[id]
	if !ok {
		return errors.WithDetail(ErrNoteNotFound, string(id))
	}

	const pathParts = 2

	parts := strings.SplitN(meta.Path, string(filepath.Separator), pathParts)
	if len(parts) < pathParts {
		return errors.WithDetail(ErrNoteNotFound, string(id))
	}

	subPath := parts[1] // "20260404/id.md"

	newRelPath := filepath.Join(destFolder, subPath)
	srcPath := filepath.Join(fs.dir, meta.Path)
	dstPath := filepath.Join(fs.dir, newRelPath)

	err = os.MkdirAll(filepath.Dir(dstPath), dirPerm)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.Rename(srcPath, dstPath)
	if err != nil {
		return errors.WithStack(err)
	}

	meta.Path = newRelPath
	fs.index[id] = meta

	err = fs.saveIndex()
	if err != nil {
		_ = os.Rename(dstPath, srcPath)

		return err
	}

	return nil
}

// Duplicate はノートを複製する。新しいIDで同じフォルダに保存し、複製されたNoteを返す。
// Body、Pinned、CreatedAt、UpdatedAt をコピー元からコピーする。
func (fs *FileStore) Duplicate(id note.NoteID) (note.Note, error) {
	meta, ok := fs.index[id]
	if !ok {
		return note.Note{}, errors.WithDetail(ErrNoteNotFound, string(id))
	}

	src, err := fs.Load(id)
	if err != nil {
		return note.Note{}, err
	}

	src.Path = meta.Path

	dup, err := note.New(src.CreatedAt)
	if err != nil {
		return note.Note{}, err
	}

	dup.Body = src.Body
	dup.Pinned = src.Pinned
	dup.Metadata.Title = src.Metadata.Title
	dup.Metadata.Preview = src.Metadata.Preview
	dup.CreatedAt = src.CreatedAt
	dup.UpdatedAt = src.UpdatedAt
	dup.Path = NotePath(src.Folder(), src.CreatedAt, dup.ID)

	err = fs.Save(dup)
	if err != nil {
		return note.Note{}, err
	}

	return dup, nil
}

func (fs *FileStore) validateRename(oldName, newName string) (string, string, error) {
	if fs.isSystemDir(oldName) {
		return "", "", errors.WithDetail(ErrSystemFolder, oldName)
	}

	if fs.isSystemDir(newName) {
		return "", "", errors.WithDetail(ErrSystemFolder, newName)
	}

	oldPath := filepath.Join(fs.dir, oldName)

	_, err := os.Stat(oldPath)
	if os.IsNotExist(err) {
		return "", "", errors.WithDetail(ErrFolderNotFound, oldName)
	}

	if err != nil {
		return "", "", errors.WithStack(err)
	}

	newPath := filepath.Join(fs.dir, newName)

	_, err = os.Stat(newPath)
	if err == nil {
		return "", "", errors.WithDetail(ErrFolderAlreadyExists, newName)
	}

	if !os.IsNotExist(err) {
		return "", "", errors.WithStack(err)
	}

	return oldPath, newPath, nil
}

// removeEmptyDirs はディレクトリ内の空サブディレクトリを再帰的に削除してから、
// 自身も削除する。ファイルが残っている場合はエラーを返す。
func removeEmptyDirs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			err := removeEmptyDirs(filepath.Join(dir, e.Name()))
			if err != nil {
				return err
			}
		}
	}

	return errors.WithStack(os.Remove(dir))
}

func (fs *FileStore) isSystemDir(name string) bool {
	return name == notesDir || strings.HasPrefix(name, ".")
}

// NotePath はフォルダ名、作成日時、IDからノートの相対パスを生成する。
func NotePath(folder string, createdAt time.Time, id note.NoteID) string {
	return filepath.Join(folder, createdAt.Format("20060102"), string(id)+".md")
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
		fs.index[note.NoteID(id)] = metadataFromEntry(id, entry)
	}

	return nil
}

func (fs *FileStore) saveIndex() error {
	idx := indexData{
		Notes: make(map[string]indexEntry, len(fs.index)),
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
