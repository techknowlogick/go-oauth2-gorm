package oauth2gorm

import (
	"fmt"
	"io"
	"encoding/json"
	"os"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type ClientStoreItem struct {
	gorm.Model

	Secret      string `gorm:"type:varchar(512)"`
	Domain    string `gorm:"type:varchar(512)"`
	Data   string `gorm:"type:text"`
}

func NewClientStore(config *Config, gcInterval int) *ClientStore {
	var d gorm.Dialector
	switch config.DBType {
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
		return nil
	}
	db, err := gorm.Open(d, defaultConfig)
	if err != nil {
		panic(err)
	}
	// default client pool
	s, err := db.DB()
	if err != nil {
		panic(err)
	}
	s.SetMaxIdleConns(10)
	s.SetMaxOpenConns(100)
	s.SetConnMaxLifetime(time.Hour)

	return NewClientStoreWithDB(config, db)
}

func NewClientStoreWithDB(config *Config, db *gorm.DB) *ClientStore {
	store := &ClientStore{
		db: db,
		tableName: "oauth2_clients",
		stdout: os.Stderr,
	}
	if config.TableName != "" {
		store.tableName = config.TableName
	}

	if !db.Migrator().HasTable(store.tableName) {
		if err := db.Table(store.tableName).Migrator().CreateTable(&TokenStoreItem{}); err != nil {
			panic(err)
		}
	}

	return store
}

type ClientStore struct {
	tableName string
	db *gorm.DB
	stdout io.Writer
}

func (s *ClientStore) toClientInfo(data []byte) (oauth2.ClientInfo, error) {
	var cm models.Client
	err := json.Unmarshal(data, &cm)
	return &cm, err
}

func (s *ClientStore) GetByID(id string) (oauth2.ClientInfo, error) {
	if id == "" {
		return nil, nil
	}

	var item ClientStoreItem
	err := s.db.Table(s.tableName).First(&item, id).Error
	if err != nil {
		return nil, err
	}

	return s.toClientInfo([]byte(item.Data))
}

func (s *ClientStore) Create(info oauth2.ClientInfo) error {
	return s.db.Table(s.tableName).Create(info).Error
}
