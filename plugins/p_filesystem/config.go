package p_filesystem

import (
	"log"

	"github.com/lariv-in/lago/lago"
)

type FilesystemConfig struct {
	StorageBackend string `toml:"storageBackend"`
	LocalDir       string `toml:"localDir"`
}

var Config = &FilesystemConfig{
	StorageBackend: "local",
	LocalDir:       "filesystem",
}

func (c *FilesystemConfig) PostConfig() {
	switch c.StorageBackend {
	case "", "local":
		Store = &LocalFilestore{BaseDir: c.LocalDir}
	default:
		log.Panicf("unsupported filesystem storage backend %q", c.StorageBackend)
	}
}

func init() {
	lago.RegistryConfig.Register("p_filesystem", Config)
	Config.PostConfig()
}
