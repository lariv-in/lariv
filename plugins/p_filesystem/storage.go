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

type FileDownload struct {
	Filename    string
	ContentType string
	Size        int64
	Reader      io.ReadCloser
}

type Filestore interface {
	Save(file *multipart.FileHeader) (string, error)
	Open(path string, name string) (*FileDownload, error)
	Delete(path string) error
}

var Store Filestore = &LocalFilestore{}

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

	dir, err := s.ensureBaseDir()
	if err != nil {
		return "", err
	}

	dst, err := os.CreateTemp(dir, "upload-*"+filepath.Ext(file.Filename))
	if err != nil {
		slog.Error("failed creating local filestore temp file", "dir", dir, "filename", file.Filename, "error", err)
		return "", err
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		_ = os.Remove(dst.Name())
		slog.Error("failed copying uploaded file to local filestore", "path", dst.Name(), "filename", file.Filename, "error", err)
		return "", err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(dst.Name())
		slog.Error("failed closing local filestore temp file", "path", dst.Name(), "error", err)
		return "", err
	}

	return dst.Name(), nil
}

func (s *LocalFilestore) Open(path string, name string) (*FileDownload, error) {
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
