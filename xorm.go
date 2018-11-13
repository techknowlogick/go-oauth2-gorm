package oauth2xorm

import (
  "fmt"
  "time"
  "os"
  "io"
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
	x, err := xorm.NewEngine(config.DBType, config.DSN)
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
	err = x.Sync2(new(StoreItem))
	if err != nil {
		panic(err)
	}

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

func (s *Store) errorf(format string, args ...interface{}) {
	if s.stdout != nil {
		buf := fmt.Sprintf(format, args...)
		s.stdout.Write([]byte(buf))
	}
}

func (s *Store) gc() {
	for range s.ticker.C {
		now := time.Now().Unix()
		counts, err := s.db.Where("expired_at > ?", now).Count(&StoreItem)
		if err != nil {
			s.errorf("[ERROR]:%s", err.Error())
			return
		} else if n > 0 {
			_, err = s.db.Where("expired_at > ?", now).Delete(&StoreItem)
			if err != nil {
				s.errorf("[ERROR]:%s", err.Error())
			}
		}
	}
}