package lago

import (
	"github.com/BurntSushi/toml"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

type LagoConfig struct {
	Debug bool
	DBType DBType
	SqliteConfig *sqlite.Config
	PostgresConfig *postgres.Config
	Address string
	CertFile string
	KeyFile string
}


type DBType string

const (
	DBTypeSqlite = DBType("Sqlite")
	DBTypePostgres = DBType("Postgres")
)


func LoadConfigFromFile(path string) (LagoConfig, error) {
	var config LagoConfig
	_, err := toml.DecodeFile(path, &config)
	return config, err
}
