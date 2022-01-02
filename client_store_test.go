package oauth2gorm

import (
	"context"
	"testing"

	"github.com/go-oauth2/oauth2/v4/models"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClientStore(t *testing.T) {
	cstore := NewClientStore(NewConfig(dsn, dbType, ""))

	Convey("Test client store", t, func() {
		Convey("Test create client", func() {
			info := &models.Client{
				ID:     "1b74413f-f3b8-409f-ac47-e8c062e3472a",
				Secret: "the secret",
				Domain: "http://localhost/",
				UserID: "1_1",
			}

			err := cstore.Create(context.Background(), info)
			So(err, ShouldBeNil)

			cinfo, err := cstore.GetByID(context.Background(), "1b74413f-f3b8-409f-ac47-e8c062e3472a")
			So(err, ShouldBeNil)
			So(cinfo.GetUserID(), ShouldEqual, info.UserID)
		})
	})
}
