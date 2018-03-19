///////////////////////////////////////////////////////////////////////////////
//
// The MIT License (MIT)
// Copyright (c) 2018 Jivan Amara
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
// OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE
// USE OR OTHER DEALINGS IN THE SOFTWARE.
//
///////////////////////////////////////////////////////////////////////////////

// go-wfs project handlers_internal_test.go

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
	expectedBody, err := json.Marshal(rootContent{
		Links: []*link{
			NewLink("http://server.com/api", "service", "application/json", "", ""),
			NewLink("http://server.com/conformance", "conformance", "application/json", "", ""),
			NewLink("http://server.com/collections", "data", "application/json", "", ""),
		},
	})
	expectedStatusCode := 200

	if err != nil {
		t.Errorf("Problem marshalling expected content: %v", err.Error())
	}

	var responseWriter *httptest.ResponseRecorder = httptest.NewRecorder()
	request := httptest.NewRequest("GET", "http://server.com/api", bytes.NewBufferString(""))
	serveAddress = "server.com"
	rootJson(responseWriter, request)
	resp := responseWriter.Result()

	if resp.StatusCode != expectedStatusCode {
		t.Errorf("status code %v != %v", resp.StatusCode, expectedStatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != string(expectedBody) {
		t.Errorf("\n%v\n--- != ---\n%v", string(body), string(expectedBody))
	}
}

func TestConformance(t *testing.T) {

}
