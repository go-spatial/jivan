package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestTesting(t *testing.T) {
	t.Logf("It's working")
}

//    /api (relation type 'service')
//    /conformance (relation type 'conformance')
//    /collections (relation type 'data')

func TestRoot(t *testing.T) {
	// Convert to strings.Builder when support for go versions < 1.10 is no longer needed
	var responseWriter *httptest.ResponseRecorder = httptest.NewRecorder()
	request := httptest.NewRequest("GET", "http://server.com/api", bytes.NewBufferString(""))
	serveAddress = "server.com"
	rootJson(responseWriter, request)
	expectedBody, err := json.Marshal(rootContent{
		Links: []*link{
			NewLink("http://server.com/api", "service", "application/json", "", ""),
			NewLink("http://server.com/conformance", "conformance", "application/json", "", ""),
			NewLink("http://server.com/collections", "data", "application/json", "", ""),
		},
	})

	if err != nil {
		t.Errorf("Problem marshalling expected content: %v", err.Error())
	}

	body, _ := ioutil.ReadAll(responseWriter.Body)
	if string(body) != string(expectedBody) {
		t.Errorf("\n%v\n--- != ---\n%v", string(body), string(expectedBody))
	}
}
