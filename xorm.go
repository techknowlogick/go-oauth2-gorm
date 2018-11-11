package oauth2xorm

import (
  "time"
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
func NewConfig(x *xorm.Engine) *Config {
	return &Config{
		X:           x,
		MaxLifetime: time.Hour * 2,
	}
}

// Config xorm configuration
type Config struct {
	X            *xorm.Engine
	MaxLifetime  time.Duration
}

// NewStore create mysql store instance,
func NewStore(config *Config, tableName string, gcInterval int) *Store {

}
