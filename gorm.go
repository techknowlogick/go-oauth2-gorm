package oauth2gorm

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
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

func NewConfig(dsn string, dbType DBType, tableName string) (*Config, gorm.Dialector) {
	config := &Config{
		DSN:         dsn,
		DBType:      dbType,
		TableName:   tableName,
		MaxLifetime: time.Hour * 2,
	}

	var d gorm.Dialector

	switch dbType {
	case MySQL:
		d = mysql.New(mysql.Config{
			DSN: config.DSN,
		})
	case PostgreSQL:
		d = postgres.New(postgres.Config{
			DSN: config.DSN,
		})
	case SQLite:
		d = sqlite.Open(config.DSN)
	case SQLServer:
		d = sqlserver.Open(config.DSN)
	default:
		fmt.Println("unsupported databases")
		d = nil
	}
	return config, d
}
