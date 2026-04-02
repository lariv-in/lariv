package p_filesystem

import (
	"io"
	"log"
	"log/slog"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalFilestore struct {
	BaseDir string
}

func (s *LocalFilestore) resolvedBaseDir() string {
	baseDir := s.BaseDir
	if baseDir == "" {
		baseDir = "filesystem"
	}
	if filepath.IsAbs(baseDir) {
		return baseDir
	}
	exe, err := os.Executable()
	if err != nil {
		log.Panicf("failed to resolve executable for filesystem storage: %v", err)
	}
	return filepath.Join(filepath.Dir(exe), baseDir)
}

func (s *LocalFilestore) ensureBaseDir() (string, error) {
	dir := s.resolvedBaseDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Error("failed to create local filestore directory", "dir", dir, "error", err)
		return "", err
	}
	return dir, nil
}

// SaveFromReader persists data from an io.Reader into the local filestore,
// using the given file extension for the temp file name. Returns the stored
// path on disk. This is the shared core used by Save and available to
// generators or any other code that has raw bytes instead of a multipart upload.
func (s *LocalFilestore) SaveFromReader(r io.Reader, ext string) (string, error) {
	dir, err := s.ensureBaseDir()
	if err != nil {
		return "", err
	}

	dst, err := os.CreateTemp(dir, "store-*"+ext)
	if err != nil {
		slog.Error("failed creating local filestore temp file", "dir", dir, "error", err)
		return "", err
	}

	if _, err := io.Copy(dst, r); err != nil {
		_ = dst.Close()
		_ = os.Remove(dst.Name())
		slog.Error("failed copying data to local filestore", "path", dst.Name(), "error", err)
		return "", err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(dst.Name())
		slog.Error("failed closing local filestore temp file", "path", dst.Name(), "error", err)
		return "", err
	}

	return dst.Name(), nil
}

func (s *LocalFilestore) Save(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", nil
	}

	src, err := file.Open()
	if err != nil {
		slog.Error("failed opening uploaded file", "filename", file.Filename, "error", err)
		return "", err
	}
	defer src.Close()

	return s.SaveFromReader(src, filepath.Ext(file.Filename))
}

func (s *LocalFilestore) Open(path, name string) (*FileDownload, error) {
	file, err := os.Open(path)
	if err != nil {
		slog.Error("failed opening local filestore path", "path", path, "error", err)
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	contentType := mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &FileDownload{
		Filename:    name,
		ContentType: contentType,
		Size:        info.Size(),
		Reader:      file,
	}, nil
}

func (s *LocalFilestore) Delete(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			slog.Warn("local filestore path already missing during delete", "path", path)
			return nil
		}
		return err
	}
	return nil
}

func (s *LocalFilestore) StoredSize(path string) (int64, error) {
	if path == "" {
		return 0, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
