package oauth2gorm

import (
	"log"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBType int8

// Config gorm configuration
type Config struct {
	DSN         string
	DBType      DBType
	TableName   string
	MaxLifetime time.Duration
}

const (
	MySQL = iota
	PostgreSQL
	SQLite
	SQLServer
)

var defaultConfig = &gorm.Config{
	Logger: logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // slow SQL
			LogLevel:      logger.Info, // log level
			Colorful:      true,        // color
		},
	),
}

func NewConfig(dsn string, dbType DBType, tableName string) *Config {
	return &Config{
		DSN:         dsn,
		DBType:      dbType,
		TableName:   tableName,
		MaxLifetime: time.Hour * 2,
	}
}
