package lago

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

type LagoConfig struct {
	Debug          bool
	DBType         DBType
	SqliteConfig   *sqlite.Config
	PostgresConfig *postgres.Config
	Address        string
	UDS            string
	GeneratorOrder []string
	TrustedOrigins []string
	Plugins        map[string]toml.Primitive
}

type DBType string

const (
	DBTypeSqlite   = DBType("Sqlite")
	DBTypePostgres = DBType("Postgres")
)

func LoadConfigFromFile(path string) (LagoConfig, error) {
	var config LagoConfig
	if path == "" {
		return config, fmt.Errorf("config path is empty")
	}

	resolvedPath := path
	if !filepath.IsAbs(resolvedPath) {
		exe, err := os.Executable()
		if err != nil {
			slog.Error("failed resolving executable path for config file", "err", err, "configPath", path)
			return config, err
		}
		resolvedPath = filepath.Join(filepath.Dir(exe), resolvedPath)
	}

	md, err := toml.DecodeFile(resolvedPath, &config)
	if err != nil {
		slog.Error("failed decoding config file", "err", err, "configPath", path, "resolvedPath", resolvedPath)
		return config, err
	}

	for key, cfgPointer := range RegistryConfig.All() {
		if prim, ok := config.Plugins[key]; ok {
			err = md.PrimitiveDecode(prim, cfgPointer)
			if err != nil {
				slog.Error("failed decoding plugin config", "err", err, "plugin", key)
				return config, err
			}
			cfgPointer.PostConfig()
		}
	}

	return config, nil
}
