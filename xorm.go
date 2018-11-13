package oauth2xorm

import (
  "fmt"
  "time"
  "os"
  "io"
  "github.com/json-iterator/go"
  "github.com/go-xorm/xorm"
  "gopkg.in/oauth2.v3"
  "gopkg.in/oauth2.v3/models"
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
		counts, err := s.db.Where("expired_at > ?", now).Count(&StoreItem{})
		if err != nil {
			s.errorf("[ERROR]:%s", err.Error())
			return
		} else if counts > 0 {
			_, err = s.db.Where("expired_at > ?", now).Delete(&StoreItem{})
			if err != nil {
				s.errorf("[ERROR]:%s", err.Error())
			}
		}
	}
}

// Create create and store the new token information
func (s *Store) Create(info oauth2.TokenInfo) error {
	buf, _ := jsoniter.Marshal(info)
	item := &StoreItem{
		Data: string(buf),
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

	_, err :=  s.db.Insert(&item)
	return err
}

// RemoveByCode delete the authorization code
func (s *Store) RemoveByCode(code string) error {
	_, err := s.db.Update(&StoreItem{}, &StoreItem{Code:""})
	return err
}

// RemoveByAccess use the access token to delete the token information
func (s *Store) RemoveByAccess(access string) error {
	_, err := s.db.Update(&StoreItem{}, &StoreItem{Access:""})
	return err
}

// RemoveByRefresh use the refresh token to delete the token information
func (s *Store) RemoveByRefresh(refresh string) error {
	_, err := s.db.Update(&StoreItem{}, &StoreItem{Refresh:""})
	return err
}

func (s *Store) toTokenInfo(data string) oauth2.TokenInfo {
	var tm models.Token
	jsoniter.Unmarshal([]byte(data), &tm)
	return &tm
}

// TODO: update below to xorm

// GetByCode use the authorization code for token information data
func (s *Store) GetByCode(code string) (oauth2.TokenInfo, error) {
	if code == "" {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE code=? LIMIT 1", s.tableName)
	var item StoreItem
	err := s.db.SelectOne(&item, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}

// GetByAccess use the access token for token information data
func (s *Store) GetByAccess(access string) (oauth2.TokenInfo, error) {
	if access == "" {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE access=? LIMIT 1", s.tableName)
	var item StoreItem
	err := s.db.SelectOne(&item, query, access)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}

// GetByRefresh use the refresh token for token information data
func (s *Store) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	if refresh == "" {
		return nil, nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE refresh=? LIMIT 1", s.tableName)
	var item StoreItem
	err := s.db.SelectOne(&item, query, refresh)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s.toTokenInfo(item.Data), nil
}