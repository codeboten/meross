package api

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDevice(t *testing.T) {
	Convey("Given a device", t, func() {
		d := Device{}
		So(d.getBrokerURL(), ShouldEqual, "tcps://us-iot.meross.com:2001")
		d.Domain = "test-domain"
		So(d.getBrokerURL(), ShouldEqual, "tcps://test-domain:2001")
	})
}
