package oauth2xorm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
)

var noUpdateContent = "No content found to be updated"

// StoreItem data item
type StoreItem struct {
	gorm.Model
	ID        int64 `gorm:"AUTO_INCREMENT"`
	ExpiredAt int64
	Code      string `gorm:"type:varchar(512)"`
	Access    string `gorm:"type:varchar(512)"`
	Refresh   string `gorm:"type:varchar(512)"`
	Data      string `gorm:"type:text"`
}

// NewConfig create mysql configuration instance
func NewConfig(dsn string, dbType string, tableName string) *Config {
	return &Config{
		DSN:         dsn,
		DBType:      dbType,
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
	db, err := gorm.Open(config.DBType, config.DSN)
	if err != nil {
		panic(err)
	}
	return NewStoreWithDB(config, db, gcInterval)
}

func NewStoreWithDB(config *Config, db *gorm.DB, gcInterval int) *Store {
	store := &Store{
		db:        db,
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

	if err := db.Table(store.tableName).CreateTable(&StoreItem{}).Error; err != nil {
		panic(err)
	}

	go store.gc()
	return store
}

// Store mysql token store
type Store struct {
	tableName string
	db        *gorm.DB
	stdout    io.Writer
	ticker    *time.Ticker
}

// SetStdout set error output
func (s *Store) SetStdout(stdout io.Writer) *Store {
	s.stdout = stdout
	return s
}

// Close close the store
func (s *Store) Close() {
	s.ticker.Stop()
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
		var count int
		if err := s.db.Table(s.tableName).Where("expired_at > ?", now).Count(&count).Error; err != nil {
			s.errorf("[ERROR]:%s\n", err)
			return
		}
		if count > 0 {
			if err := s.db.Table(s.tableName).Where("expired_at > ?", now).Delete(&StoreItem{}).Error; err != nil {
				s.errorf("[ERROR]:%s\n", err)
			}
		}
	}
}

// Create create and store the new token information
func (s *Store) Create(info oauth2.TokenInfo) error {
	jv, err := json.Marshal(info)
	if err != nil {
		return err
	}
	item := &StoreItem{
		Data: string(jv),
	}

	if code := info.GetCode(); code != "" {
		item.Code = code
		item.ExpiredAt = info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()).Unix()
	} else {
		item.Access = info.GetAccess()
		item.ExpiredAt = info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()).Unix()

		if refresh := info.GetRefresh(); refresh != "" {
			item.Refresh = info.GetRefresh()
			item.ExpiredAt = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()).Unix()
		}
	}

	return s.db.Table(s.tableName).Create(item).Error
}

// RemoveByCode delete the authorization code
func (s *Store) RemoveByCode(code string) error {
	return s.db.Table(s.tableName).Where("code = ?", code).Update("code", "").Error
}

// RemoveByAccess use the access token to delete the token information
func (s *Store) RemoveByAccess(access string) error {
	return s.db.Table(s.tableName).Where("access = ?", access).Update("access", "").Error
}

// RemoveByRefresh use the refresh token to delete the token information
func (s *Store) RemoveByRefresh(refresh string) error {
	return s.db.Table(s.tableName).Where("refresh = ?", refresh).Update("refresh", "").Error
}

func (s *Store) toTokenInfo(data string) oauth2.TokenInfo {
	var tm models.Token
	err := json.Unmarshal([]byte(data), &tm)
	if err != nil {
		return nil
	}
	return &tm
}

// GetByCode use the authorization code for token information data
func (s *Store) GetByCode(code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.db.Table(s.tableName).Where("code = ?", code).Find(&item).Error; err != nil {
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}

// GetByAccess use the access token for token information data
func (s *Store) GetByAccess(access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.db.Table(s.tableName).Where("access = ?", access).Find(&item).Error; err != nil {
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}

// GetByRefresh use the refresh token for token information data
func (s *Store) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	var item StoreItem
	if err := s.db.Table(s.tableName).Where("refresh = ?", refresh).Find(&item).Error; err != nil {
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}
