package oauth2gorm

import (
	"testing"

	"github.com/go-oauth2/oauth2/v4/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClientStore(t *testing.T) {
	cstore := NewClientStore(NewConfig(dsn, dbType, ""))

	Convey("Test client store", t, func() {
		Convey("Test create client", func() {
			info := &models.Client{
				ID:     "1",
				Secret: "the secret",
				Domain: "http://localhost/",
				UserID: "1_1",
			}

			err := cstore.Create(info)
			So(err, ShouldBeNil)

			cinfo, err := cstore.GetByID("1")
			So(err, ShouldBeNil)
			So(cinfo.GetUserID(), ShouldEqual, info.UserID)
		})
	})
}
