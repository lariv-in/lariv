package p_filesystem

import (
	"errors"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/lariv-in/lariv"
	"gorm.io/gorm"
)

// dbFileInfo implements fs.FileInfo for VNode.
type dbFileInfo struct {
	name    string
	size    int64
	isDir   bool
	modTime time.Time
}

func (fi dbFileInfo) Name() string { return fi.name }
func (fi dbFileInfo) Size() int64  { return fi.size }
func (fi dbFileInfo) Mode() fs.FileMode {
	if fi.isDir {
		return fs.ModeDir | 0o555
	}
	return 0o444
}
func (fi dbFileInfo) ModTime() time.Time { return fi.modTime }
func (fi dbFileInfo) IsDir() bool        { return fi.isDir }
func (fi dbFileInfo) Sys() any           { return nil }

// dbDirEntry implements fs.DirEntry for VNode.
type dbDirEntry struct {
	info dbFileInfo
}

func (de dbDirEntry) Name() string               { return de.info.name }
func (de dbDirEntry) IsDir() bool                { return de.info.isDir }
func (de dbDirEntry) Type() fs.FileMode          { return de.info.Mode().Type() }
func (de dbDirEntry) Info() (fs.FileInfo, error) { return de.info, nil }

// dbFile implements fs.File and fs.ReadDirFile for database-backed files and directories.
type dbFile struct {
	node           *VNode
	info           dbFileInfo
	reader         io.ReadCloser
	db             *gorm.DB
	readDirEntries []fs.DirEntry
	readDirOffset  int
}

func (f *dbFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *dbFile) Read(b []byte) (int, error) {
	if f.info.isDir {
		return 0, &fs.PathError{Op: "read", Path: f.info.name, Err: errors.New("is a directory")}
	}
	if f.reader == nil {
		return 0, io.EOF
	}
	return f.reader.Read(b)
}

func (f *dbFile) Close() error {
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

func (f *dbFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if !f.info.isDir {
		return nil, &fs.PathError{Op: "readdir", Path: f.info.name, Err: errors.New("not a directory")}
	}
	if f.readDirEntries == nil {
		var children []VNode
		query := f.db.Order("is_directory DESC, name ASC")
		if f.node == nil {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", f.node.ID)
		}
		if err := query.Find(&children).Error; err != nil {
			return nil, err
		}

		f.readDirEntries = make([]fs.DirEntry, len(children))
		for i, child := range children {
			var size int64
			if !child.IsDirectory {
				sz, _ := child.GetFileSize()
				size = int64(sz)
			}
			f.readDirEntries[i] = dbDirEntry{
				info: dbFileInfo{
					name:    child.Name,
					size:    size,
					isDir:   child.IsDirectory,
					modTime: child.UpdatedAt,
				},
			}
		}
	}

	if n <= 0 {
		entries := f.readDirEntries[f.readDirOffset:]
		f.readDirOffset = len(f.readDirEntries)
		return entries, nil
	}

	if f.readDirOffset >= len(f.readDirEntries) {
		return nil, io.EOF
	}

	end := f.readDirOffset + n
	if end > len(f.readDirEntries) {
		end = len(f.readDirEntries)
	}
	entries := f.readDirEntries[f.readDirOffset:end]
	f.readDirOffset = end
	return entries, nil
}

// dbFilesystem implements lariv.UsefulFilesystem reading from the database VNodes.
type dbFilesystem struct {
	db *gorm.DB
}

func (dfs *dbFilesystem) Open(name string) (fs.File, error) {
	node, err := resolveVNode(dfs.db, name)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	if node == nil {
		// Root directory
		return &dbFile{
			node: nil,
			info: dbFileInfo{
				name:    ".",
				size:    0,
				isDir:   true,
				modTime: time.Time{},
			},
			db: dfs.db,
		}, nil
	}

	if node.IsDirectory {
		return &dbFile{
			node: node,
			info: dbFileInfo{
				name:    node.Name,
				size:    0,
				isDir:   true,
				modTime: node.UpdatedAt,
			},
			db: dfs.db,
		}, nil
	}

	download, err := node.OpenDownload()
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	size, _ := node.GetFileSize()
	return &dbFile{
		node: node,
		info: dbFileInfo{
			name:    node.Name,
			size:    int64(size),
			isDir:   false,
			modTime: node.UpdatedAt,
		},
		reader: download.Reader,
		db:     dfs.db,
	}, nil
}

func (dfs *dbFilesystem) ReadDir(name string) ([]fs.DirEntry, error) {
	file, err := dfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dirFile, ok := file.(fs.ReadDirFile)
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: errors.New("not a directory")}
	}

	return dirFile.ReadDir(-1)
}

func (dfs *dbFilesystem) ReadFile(name string) ([]byte, error) {
	file, err := dfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// resolveVNode traverses standard fs paths like "a/b/c" starting from database root.
func resolveVNode(db *gorm.DB, path string) (*VNode, error) {
	if !fs.ValidPath(path) {
		return nil, fs.ErrInvalid
	}
	if path == "." || path == "" {
		return nil, nil
	}
	parts := strings.Split(path, "/")
	var current *VNode
	for _, part := range parts {
		var next VNode
		query := db.Where("name = ?", part)
		if current == nil {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", current.ID)
		}
		err := query.First(&next).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fs.ErrNotExist
			}
			return nil, err
		}
		current = &next
	}
	return current, nil
}

// NewDatabaseFilesystem returns a lariv.UsefulFilesystem implementation reading from the database.
func NewDatabaseFilesystem(db *gorm.DB) lariv.UsefulFilesystem {
	return &dbFilesystem{db: db}
}
