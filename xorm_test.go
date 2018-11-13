package oauth2xorm

import (
	"testing"
	"time"

	"gopkg.in/oauth2.v3/models"
	"github.com/stretchr/testify/assert"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	dbType = "sqlite3"
	dsn = "file::memory:?cache=shared"
)

func TestTokenStore(t *testing.T) {
	// wait gc
	defer time.Sleep(time.Second * 2)

	assert := assert.New(t)

	store := NewStore(NewConfig(dsn, dbType, ""), 1)

	// "Test authorization code store"
	info := &models.Token{
		ClientID:      "1",
		UserID:        "1_1",
		RedirectURI:   "http://localhost/",
		Scope:         "all",
		Code:          "11_11_11",
		CodeCreateAt:  time.Now(),
		CodeExpiresIn: time.Second * 5,
	}

	err := store.Create(info)
	assert.Nil(err)

	cinfo, err := store.GetByCode(info.Code)
	assert.Nil(err)
	assert.Equal(cinfo.GetUserID(), info.UserID, "they should be equal")

	err = store.RemoveByCode(info.Code)
	assert.Nil(err)
	cinfo, err = store.GetByCode(info.Code)
	assert.Nil(err)
	assert.Nil(cinfo)

	// "Test access token store"
	info = &models.Token{
		ClientID:        "1",
		UserID:          "1_1",
		RedirectURI:     "http://localhost/",
		Scope:           "all",
		Access:          "1_1_1",
		AccessCreateAt:  time.Now(),
		AccessExpiresIn: time.Second * 5,
	}
	err = store.Create(info)
	assert.Nil(err)
	ainfo, err := store.GetByAccess(info.GetAccess())
	assert.Nil(err)
	assert.Equal(ainfo.GetUserID(), info.GetUserID(), "they should be equal")

	err = store.RemoveByAccess(info.GetAccess())
	assert.Nil(err)

	ainfo, err = store.GetByAccess(info.GetAccess())
	assert.Nil(err)
	assert.Nil(ainfo)

	// "Test refresh token store"
	info = &models.Token{
		ClientID:         "1",
		UserID:           "1_2",
		RedirectURI:      "http://localhost/",
		Scope:            "all",
		Access:           "1_2_1",
		AccessCreateAt:   time.Now(),
		AccessExpiresIn:  time.Second * 5,
		Refresh:          "1_2_2",
		RefreshCreateAt:  time.Now(),
		RefreshExpiresIn: time.Second * 15,
	}
	err := store.Create(info)
	assert.Nil(err)

	ainfo, err := store.GetByAccess(info.GetAccess())
	assert.Nil(err)\
	assert.Equal(ainfo.GetUserID(), info.GetUserID(), "they should be equal")

	err = store.RemoveByAccess(info.GetAccess())
	assert.Nil(err)

	ainfo, err = store.GetByAccess(info.GetAccess())
	assert.Nil(err)
	assert.Nil(ainfo)

	rinfo, err := store.GetByRefresh(info.GetRefresh())
	assert.Nil(err)
	assert.Equal(rinfo.GetUserID(), info.GetUserID(), "they should be equal")

	err = store.RemoveByRefresh(info.GetRefresh())
	assert.Nil(err)

	rinfo, err = store.GetByRefresh(info.GetRefresh())
	assert.Nil(err)
	assert.Nil(rinfo)

}
