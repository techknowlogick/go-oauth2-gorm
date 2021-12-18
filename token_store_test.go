package oauth2gorm

import (
	"context"
	"testing"
	"time"

	"github.com/go-oauth2/oauth2/v4/models"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	dbType = SQLite
	dsn    = "file::memory:?cache=shared"
)

func TestTokenStore(t *testing.T) {
	// wait gc
	defer time.Sleep(time.Second * 2)

	config, dialector := NewConfig(dsn, dbType, "")

	store := NewTokenStore(config, dialector, 1)

	Convey("Test token store", t, func() {
		Convey("Test authorization code store", func() {
			info := &models.Token{
				ClientID:      "1",
				UserID:        "1_1",
				RedirectURI:   "http://localhost/",
				Scope:         "all",
				Code:          "11_11_11",
				CodeCreateAt:  time.Now(),
				CodeExpiresIn: time.Second * 5,
			}
			err := store.Create(context.Background(), info)
			So(err, ShouldBeNil)

			cinfo, err := store.GetByCode(context.Background(), info.Code)
			So(err, ShouldBeNil)
			So(cinfo.GetUserID(), ShouldEqual, info.UserID)

			err = store.RemoveByCode(context.Background(), info.Code)
			So(err, ShouldBeNil)

			cinfo, err = store.GetByCode(context.Background(), info.Code)
			So(err, ShouldBeNil)
			So(cinfo, ShouldBeNil)
		})

		Convey("Test access token store", func() {
			info := &models.Token{
				ClientID:        "1",
				UserID:          "1_1",
				RedirectURI:     "http://localhost/",
				Scope:           "all",
				Access:          "1_1_1",
				AccessCreateAt:  time.Now(),
				AccessExpiresIn: time.Second * 5,
			}
			err := store.Create(context.Background(), info)
			So(err, ShouldBeNil)

			ainfo, err := store.GetByAccess(context.Background(), info.GetAccess())
			So(err, ShouldBeNil)
			So(ainfo.GetUserID(), ShouldEqual, info.GetUserID())

			err = store.RemoveByAccess(context.Background(), info.GetAccess())
			So(err, ShouldBeNil)

			ainfo, err = store.GetByAccess(context.Background(), info.GetAccess())
			So(err, ShouldBeNil)
			So(ainfo, ShouldBeNil)
		})

		Convey("Test refresh token store", func() {
			info := &models.Token{
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
			err := store.Create(context.Background(), info)
			So(err, ShouldBeNil)

			ainfo, err := store.GetByAccess(context.Background(), info.GetAccess())
			So(err, ShouldBeNil)
			So(ainfo.GetUserID(), ShouldEqual, info.GetUserID())

			err = store.RemoveByAccess(context.Background(), info.GetAccess())
			So(err, ShouldBeNil)

			ainfo, err = store.GetByAccess(context.Background(), info.GetAccess())
			So(err, ShouldBeNil)
			So(ainfo, ShouldBeNil)

			rinfo, err := store.GetByRefresh(context.Background(), info.GetRefresh())
			So(err, ShouldBeNil)
			So(rinfo.GetUserID(), ShouldEqual, info.GetUserID())

			err = store.RemoveByRefresh(context.Background(), info.GetRefresh())
			So(err, ShouldBeNil)

			rinfo, err = store.GetByRefresh(context.Background(), info.GetRefresh())
			So(err, ShouldBeNil)
			So(rinfo, ShouldBeNil)
		})
	})
}
