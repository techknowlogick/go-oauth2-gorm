package oauth2xorm

import (
  "time"
  "os"
  "github.com/go-xorm/xorm"
)

// StoreItem data item
type StoreItem struct {
    ID        int64  `xorm:"pk autoincr"`
	ExpiredAt time.Time
	Code      string `xorm:"varchar(512)"`
	Access    string `xorm:"varchar(512)"`
	Refresh   string `xorm:"varchar(512)"`
	Data      string `xorm:"text"`
}

// NewConfig create mysql configuration instance
func NewConfig(dsn string, dbType string, tableName string) *Config {
	return &Config{
		DSN:         dsn,
		DBType:		 dbType,
		TableName:   tableName,
		MaxLifetime: time.Hour * 2,
	}
}

// Config xorm configuration
type Config struct {
	DSN         string
	DBType      string
	TableName   string
	MaxLifetime time.Duration
}

// NewStore create mysql store instance,
func NewStore(config *Config, gcInterval int) *Store {
	var err error
	x, err = xorm.NewEngine(config.DBType config.DSN)
	if err != nil {
		panic(err)
	}
	store := &Store{
		db:        x,
		tableName: "oauth2_token",
		stdout:    os.Stderr,
	}
	if config.TableName != "" {
		store.tableName = config.TableName
	}
	interval := 600
	if gcInterval > 0 {
		interval = gcInterval
	}
	store.ticker = time.NewTicker(time.Second * time.Duration(interval))

	// TODO: create table if not exist

	go store.gc()
	return store
}

// Store mysql token store
type Store struct {
	tableName string
	db        *xorm.Engine
	stdout    io.Writer
	ticker    *time.Ticker
}