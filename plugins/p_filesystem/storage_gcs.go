package p_filesystem

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// GCSFilestore stores files in a Google Cloud Storage bucket. The string returned
// by Save / SaveFromReader is the object name within the configured bucket.
type GCSFilestore struct {
	client *storage.Client
	bucket *storage.BucketHandle
	prefix string
}

// NewGCSFilestore builds a filestore backed by GCS. If credentialsFile is empty,
// Application Default Credentials are used (metadata server on GCP, or
// GOOGLE_APPLICATION_CREDENTIALS / gcloud auth elsewhere). prefix is normalized
// to a non-empty path ending with "/".
func NewGCSFilestore(ctx context.Context, bucketName, credentialsFile, prefix string) (*GCSFilestore, error) {
	if bucketName == "" {
		return nil, errors.New("gcs bucket name is required")
	}
	prefix = normalizeGCSPrefix(prefix)
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, credentialsFile))
	}
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		slog.Error("failed creating GCS storage client", "bucket", bucketName, "error", err)
		return nil, err
	}
	return &GCSFilestore{
		client: client,
		bucket: client.Bucket(bucketName),
		prefix: prefix,
	}, nil
}

func normalizeGCSPrefix(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "lago/"
	}
	p = strings.TrimPrefix(p, "/")
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func (s *GCSFilestore) newObjectKey(ext string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		slog.Error("failed generating random GCS object key", "error", err)
		return "", err
	}
	return s.prefix + hex.EncodeToString(b[:]) + ext, nil
}

func (s *GCSFilestore) SaveFromReader(r io.Reader, ext string) (string, error) {
	key, err := s.newObjectKey(ext)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	w := s.bucket.Object(key).NewWriter(ctx)
	w.ContentType = mime.TypeByExtension(ext)
	if w.ContentType == "" {
		w.ContentType = "application/octet-stream"
	}
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		slog.Error("failed uploading to GCS", "key", key, "error", err)
		_ = s.bucket.Object(key).Delete(ctx)
		return "", err
	}
	if err := w.Close(); err != nil {
		slog.Error("failed closing GCS writer", "key", key, "error", err)
		_ = s.bucket.Object(key).Delete(ctx)
		return "", err
	}
	return key, nil
}

func (s *GCSFilestore) Save(file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", nil
	}
	src, err := file.Open()
	if err != nil {
		slog.Error("failed opening uploaded file for GCS", "filename", file.Filename, "error", err)
		return "", err
	}
	defer src.Close()
	return s.SaveFromReader(src, filepath.Ext(file.Filename))
}

func (s *GCSFilestore) Open(path, name string) (*FileDownload, error) {
	if path == "" {
		return nil, os.ErrNotExist
	}
	ctx := context.Background()
	obj := s.bucket.Object(path)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		slog.Error("failed reading GCS object attrs", "key", path, "error", err)
		return nil, err
	}
	r, err := obj.NewReader(ctx)
	if err != nil {
		slog.Error("failed opening GCS object reader", "key", path, "error", err)
		return nil, err
	}
	contentType := attrs.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(name))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return &FileDownload{
		Filename:    name,
		ContentType: contentType,
		Size:        attrs.Size,
		Reader:      r,
	}, nil
}

func (s *GCSFilestore) Delete(path string) error {
	if path == "" {
		return nil
	}
	ctx := context.Background()
	if err := s.bucket.Object(path).Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			slog.Warn("GCS object already missing during delete", "key", path)
			return nil
		}
		slog.Error("failed deleting GCS object", "key", path, "error", err)
		return err
	}
	return nil
}

func (s *GCSFilestore) StoredSize(path string) (int64, error) {
	if path == "" {
		return 0, nil
	}
	ctx := context.Background()
	attrs, err := s.bucket.Object(path).Attrs(ctx)
	if err != nil {
		return 0, err
	}
	return attrs.Size, nil
}
