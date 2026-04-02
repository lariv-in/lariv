package p_filesystem

import (
	"context"
	"log"

	"github.com/lariv-in/lago/lago"
)

type StorageBackend string

const (
	StorageBackendLocal StorageBackend = "local"
	StorageBackendGCS   StorageBackend = "gcs"
)

type FilesystemConfig struct {
	StorageBackend StorageBackend `toml:"storageBackend"`
	LocalDir       string         `toml:"localDir"`
	// GCS: bucket name (required when storageBackend is "gcs").
	GCSBucket string `toml:"gcsBucket"`
	// GCS: path to service account JSON key file. Empty uses Application Default Credentials.
	GCSCredentialsFile string `toml:"gcsCredentialsFile"`
	// GCS: object key prefix (default "lago/"). Normalized to end with "/".
	GCSPrefix string `toml:"gcsPrefix"`
}

var Config = &FilesystemConfig{
	StorageBackend: StorageBackendLocal,
	LocalDir:       "filesystem",
}

func (c *FilesystemConfig) PostConfig() {
	switch c.StorageBackend {
	case "", StorageBackendLocal:
		Store = &LocalFilestore{BaseDir: c.LocalDir}
	case StorageBackendGCS:
		if c.GCSBucket == "" {
			log.Panicf("filesystem storageBackend %q requires gcsBucket in config", c.StorageBackend)
		}
		fs, err := NewGCSFilestore(context.Background(), c.GCSBucket, c.GCSCredentialsFile, c.GCSPrefix)
		if err != nil {
			log.Panicf("failed to initialize GCS filestore: %v", err)
		}
		Store = fs
	default:
		log.Panicf("unsupported filesystem storage backend %q", c.StorageBackend)
	}
}

func init() {
	lago.RegistryConfig.Register("p_filesystem", Config)
	Config.PostConfig()
}
