package lago

import (
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
	CertFile       string
	KeyFile        string
	Plugins        map[string]toml.Primitive
}

type DBType string

const (
	DBTypeSqlite   = DBType("Sqlite")
	DBTypePostgres = DBType("Postgres")
)

func LoadConfigFromFile(path string) (LagoConfig, error) {
	var config LagoConfig
	md, err := toml.DecodeFile(path, &config)
	if err != nil {
		return config, err
	}

	for key, cfgPointer := range RegistryConfig.All() {
		if prim, ok := config.Plugins[key]; ok {
			err = md.PrimitiveDecode(prim, cfgPointer)
			if err != nil {
				return config, err
			}
			cfgPointer.PostConfig()
		}
	}

	return config, nil
}
