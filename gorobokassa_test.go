package gorobokassa

import (
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestUrlGeneration(t *testing.T) {
	Convey("Url", t, func() {
		in := buildRedirectUrl("lel", "lel", 11, 500, "KEK")
		out := "https://auth.robokassa.ru/Merchant/Index.aspx?Desc=KEK&InvId=11&MrchLogin=lel&OutSum=500&SignatureValue=e8c1c7bcacfa991b8612f2759804abd9"
		So(in, ShouldEqual, out)
	})
	Convey("Result", t, func() {
		request := &http.Request{}
		q := url.URL{}
		params := url.Values{}
		params.Add("OutSum", "1200")
		params.Add("InvId", "666")
		params.Add("SignatureValue", "3a3869287aaa475dda04d93280705839")
		q.RawQuery = params.Encode()
		request.URL = &q
		log.Println(q.String())
		So(verifyResult("password", 666, 1200, "3a3869287aaa475dda04d93280705839"), ShouldBeTrue)
		So(verifyResult("password", 666, 1200, "3a3869287aaa475dda04e93280705839"), ShouldBeFalse)
		So(verifyRequest("password", request), ShouldBeTrue)
		So(verifyRequest("test", request), ShouldBeFalse)
	})

	Convey("Client", t, func() {
		c := New("login", "pwd1", "password")
		Convey("Url", func() {
			in := c.Url(110, 2000, "description")
			out := "https://auth.robokassa.ru/Merchant/Index.aspx?Desc=description&InvId=110&MrchLogin=login&OutSum=2000&SignatureValue=1364f38f54e76a0affe62974bfdbde85"
			So(in, ShouldEqual, out)
		})
		Convey("CheckResult", func() {
			request := &http.Request{}
			q := url.URL{}
			params := url.Values{}
			params.Add("OutSum", "1200")
			params.Add("InvId", "666")
			params.Add("SignatureValue", "3a3869287aaa475dda04d93280705839")
			q.RawQuery = params.Encode()
			request.URL = &q
			So(c.CheckResult(request), ShouldBeTrue)
		})
		Convey("CheckSuccess", func() {
			request := &http.Request{}
			q := url.URL{}
			params := url.Values{}
			params.Add("OutSum", "1200")
			params.Add("InvId", "666")
			params.Add("SignatureValue", "1bc5c075a999e194e6d46ba351e4c11e")
			q.RawQuery = params.Encode()
			request.URL = &q
			So(c.CheckSuccess(request), ShouldBeTrue)
		})
		Convey("NotNumber", func() {
			Convey("InvId", func() {
				request := &http.Request{}
				q := url.URL{}
				params := url.Values{}
				params.Add("OutSum", "1200")
				params.Add("InvId", "a666")
				params.Add("SignatureValue", CRC(1200, "a666", c.firstPassword))
				q.RawQuery = params.Encode()
				request.URL = &q
				So(c.CheckSuccess(request), ShouldBeFalse)
			})
			Convey("OutSum", func() {
				request := &http.Request{}
				q := url.URL{}
				params := url.Values{}
				params.Add("OutSum", "12b00")
				params.Add("InvId", "666")
				params.Add("SignatureValue", CRC("12b00", "666", c.firstPassword))
				q.RawQuery = params.Encode()
				request.URL = &q
				So(c.CheckSuccess(request), ShouldBeFalse)
			})
		})
	})
}
