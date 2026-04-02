package p_filesystem

import (
	"errors"
	"io"
	"mime/multipart"
	"os"

	"cloud.google.com/go/storage"
)

type FileDownload struct {
	Filename    string
	ContentType string
	Size        int64
	Reader      io.ReadCloser
}

// Filestore persists uploaded files and serves them back by the opaque path
// string returned from Save / SaveFromReader (local filesystem path or GCS object key).
type Filestore interface {
	Save(file *multipart.FileHeader) (string, error)
	SaveFromReader(r io.Reader, ext string) (string, error)
	Open(path, name string) (*FileDownload, error)
	Delete(path string) error
	StoredSize(path string) (int64, error)
}

// Store is the active backend; it is set only from FilesystemConfig.PostConfig
// (see config.go: init and lago.LoadConfigFromFile after TOML decode). It is nil
// until the first PostConfig run.
var Store Filestore

// IsStoredFileMissing reports whether err indicates the backing blob is absent
// (local file or GCS object).
func IsStoredFileMissing(err error) bool {
	if err == nil {
		return false
	}
	if os.IsNotExist(err) {
		return true
	}
	return errors.Is(err, storage.ErrObjectNotExist)
}
