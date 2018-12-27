package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func getTestLogger() *log.Logger {
	buf := bytes.Buffer{}
	logger := log.New(&buf, "logger: ", log.Lshortfile)
	return logger
}

func getTestServer(jsonResponse string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, jsonResponse)
	}))
}

func TestGetSupportedDevices(t *testing.T) {
	Convey("Given a client request is made without a login", t, func() {
		apiResponse := `
		{"apiStatus":1022,"sysStatus":0,"data":null,"info":"No login","timeStamp":"2018-12-17 07:16:04"}
		`
		ts := getTestServer(apiResponse)
		defer ts.Close()
		api := Client{
			BaseURL:    ts.URL,
			HTTPClient: ts.Client(),
			Logger:     getTestLogger(),
		}
		devices, err := api.GetSupportedDevices()
		So(err, ShouldEqual, errNoLogin)
		So(len(devices), ShouldEqual, 0)
	})
}

func TestLogin(t *testing.T) {
	Convey("Given invalid credentials", t, func() {
		apiResponse := `
		{"apiStatus":1023,"sysStatus":0,"data":null,"info":"Sign check failed","timeStamp":"2018-12-19 14:28:16"}
		`
		ts := getTestServer(apiResponse)
		defer ts.Close()
		api := Client{
			BaseURL:    ts.URL,
			HTTPClient: ts.Client(),
			Logger:     getTestLogger(),
		}
		err := api.Login()
		So(err, ShouldEqual, errSignCheckFailed)
	})
	Convey("Given missing fields", t, func() {
		Convey("Email field missing", func() {
			apiResponse := `
		{"apiStatus":1000,"sysStatus":0,"data":null,"info":"Lack user","timeStamp":"2018-12-19 14:28:16"}
		`
			ts := getTestServer(apiResponse)
			defer ts.Close()
			api := Client{
				BaseURL:    ts.URL,
				HTTPClient: ts.Client(),
				Logger:     getTestLogger(),
			}
			err := api.Login()
			So(err, ShouldEqual, errMissingUser)
		})
		Convey("Password field missing", func() {
			apiResponse := `
		{"apiStatus":1000,"sysStatus":0,"data":null,"info":"Lack password","timeStamp":"2018-12-19 14:28:16"}
		`
			ts := getTestServer(apiResponse)
			defer ts.Close()
			api := Client{
				BaseURL:    ts.URL,
				HTTPClient: ts.Client(),
				Logger:     getTestLogger(),
			}
			err := api.Login()
			So(err, ShouldEqual, errMissingPassword)
		})
	})
	Convey("Given valid credentials", t, func() {
		apiResponse := `
			{"apiStatus":1023,"sysStatus":0,"data":null,"info":"Sign check failed","timeStamp":"2018-12-19 14:28:16"}
			`
		ts := getTestServer(apiResponse)
		defer ts.Close()
		api := Client{
			BaseURL:    ts.URL,
			HTTPClient: ts.Client(),
			Logger:     getTestLogger(),
		}
		api.Login()
	})
}

func TestSignRequest(t *testing.T) {
	Convey("Given a request", t, func() {
		os.Setenv("SECRET", "TEST_SECRET")

		loginParameters := map[string]string{
			"email":    "1234",
			"password": "1234",
		}
		expected := "eyJlbWFpbCI6IjEyMzQiLCJwYXNzd29yZCI6IjEyMzQifQ=="
		So(encodeParameters(loginParameters), ShouldEqual, expected)
		expected = "b0af289ceb39b61ea635dbef2984f749"
		So(signRequest(10, "1234", encodeParameters(loginParameters)), ShouldEqual, expected)
	})
}
