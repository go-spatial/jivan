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

// TODO: The package var serveAddress from server.go is used extensively here.  Update
//	for safe test parallelism.

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-spatial/go-wfs/data_provider"
	"github.com/go-spatial/go-wfs/test_data"
)

var testingProvider data_provider.Provider

func init() {
	gpkgTiler := test_data.GetTestGPKGTiler()
	testingProvider = data_provider.Provider{Tiler: gpkgTiler}

	// @See server.go
	// This is the host:port the server expects requests at & is used to create links in responses.
	serveAddress = "server.com"
	// This is the provider the server will use for data
	Provider = testingProvider
}

func TestRoot(t *testing.T) {
	rc := rootContent{
		Links: []*link{
			&link{
				Href: fmt.Sprintf("http://%v/api", serveAddress),
				Rel:  "service",
			},
			&link{
				Href: fmt.Sprintf("http://%v/conformance", serveAddress),
				Rel:  "conformance",
			},
			&link{
				Href: fmt.Sprintf("http://%v/collections", serveAddress),
				Rel:  "data",
			},
		},
	}
	rc.ContentType("application/json")

	expectedBody, err := json.Marshal(rc)

	expectedStatusCode := 200

	if err != nil {
		t.Errorf("Problem marshalling expected content: %v", err)
	}

	var responseWriter *httptest.ResponseRecorder = httptest.NewRecorder()
	request := httptest.NewRequest("GET", fmt.Sprintf("http://%v/api", serveAddress), bytes.NewBufferString(""))
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

func TestApi(t *testing.T) {
	// TODO: This is pretty circular logic, as the /api endpoint simply returns openapiSpecJson.
	//	Make a better test plan.
	expectedApiContent := OpenAPI3SchemaJSON
	expectedStatusCode := 200
	responseWriter := httptest.NewRecorder()
	request := httptest.NewRequest("GET", fmt.Sprintf("http://%v/api", serveAddress), bytes.NewBufferString(""))
	openapiJson(responseWriter, request)
	resp := responseWriter.Result()

	if resp.StatusCode != expectedStatusCode {
		t.Errorf("status code %v != %v", resp.StatusCode, expectedStatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != string(expectedApiContent) {
		t.Errorf("\n%v\n--- != ---\n%v", string(body), string(expectedApiContent))
	}
}

func TestConformance(t *testing.T) {
	conformanceUrl := fmt.Sprintf("http://%v/conformance", serveAddress)

	expectedBody, err := json.Marshal(conformanceClasses{
		ConformsTo: []string{
			"http://www.opengis.net/spec/wfs-1/3.0/req/core",
			"http://www.opengis.net/spec/wfs-1/3.0/req/geojson",
		},
	})
	if err != nil {
		t.Errorf("problem marshalling expectedBody: %v", err)
	}
	expectedStatusCode := 200

	responseWriter := httptest.NewRecorder()
	request := httptest.NewRequest("GET", conformanceUrl, bytes.NewBufferString(""))
	conformanceJson(responseWriter, request)
	resp := responseWriter.Result()

	if resp.StatusCode != expectedStatusCode {
		t.Errorf("status code %v != %v", resp.StatusCode, expectedStatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Problem reading response: %v", err)
	}

	if string(body) != string(expectedBody) {
		t.Errorf("\n%v\n--- != ---\n%v", string(body), string(expectedBody))
	}
}

func TestCollections(t *testing.T) {
	collectionsUrl := fmt.Sprintf("http://%v/collections", serveAddress)
	cNames, err := testingProvider.CollectionNames()
	if err != nil {
		t.Errorf("Problem getting collection names: %v", err)
	}

	csInfo := collectionsInfo{Links: []*link{}, Collections: []*collectionInfo{}}
	for _, cn := range cNames {
		collectionUrl := fmt.Sprintf("http://%v/collections/%v", serveAddress, cn)
		cInfo := collectionInfo{Name: cn, Links: []*link{&link{Rel: "self", Href: collectionUrl, Type: "application/json"}}}
		cLink := link{Href: cn, Rel: "item", Type: "application/json"}

		csInfo.Links = append(csInfo.Links, &cLink)
		csInfo.Collections = append(csInfo.Collections, &cInfo)
	}

	expectedStatus := 200
	expectedContent, err := json.Marshal(csInfo)
	if err != nil {
		t.Errorf("Problem marshalling expected collections info: %v", err)
	}

	responseWriter := httptest.NewRecorder()
	request := httptest.NewRequest("GET", collectionsUrl, bytes.NewBufferString(""))
	collectionsMetaDataJson(responseWriter, request)

	resp := responseWriter.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Problem reading response body: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Status code %v != %v", resp.StatusCode, expectedStatus)
	}

	if string(body) != string(expectedContent) {
		// Human readable versions of each
		bBuf := bytes.NewBufferString("")
		eBuf := bytes.NewBufferString("")
		json.Indent(bBuf, body, "", "  ")
		json.Indent(eBuf, expectedContent, "", "  ")

		hrBody, err := ioutil.ReadAll(bBuf)
		if err != nil {
			t.Errorf("Problem reading human-friendly body: %v", err)
		}
		hrExpected, err := ioutil.ReadAll(eBuf)
		if err != nil {
			t.Errorf("Problem reading human-friendly expected: %v", err)
		}

		hrBodyLines := strings.Split(string(hrBody), "\n")
		hrExpectedLines := strings.Split(string(hrExpected), "\n")
		maxInt := func(a, b int) int {
			if a > b {
				return a
			}
			return b
		}
		minInt := func(a, b int) int {
			if a < b {
				return a
			}
			return b
		}
		for i, bLine := range hrBodyLines {
			if bLine != hrExpectedLines[i] {
				firstLineIdx := maxInt(i-5, 0)
				lastLineIdxB := minInt(i+5, len(hrBodyLines))
				lastLineIdxE := minInt(i+5, len(hrExpectedLines))

				mismatchB := strings.Join(hrBodyLines[firstLineIdx:lastLineIdxB], "\n")
				mismatchE := strings.Join(hrExpectedLines[firstLineIdx:lastLineIdxE], "\n")
				t.Errorf("Result doesn't match expected at line %v, showing %v-%v:\n%v\n--- != ---\n%v\n",
					i, firstLineIdx, lastLineIdxB, mismatchB, mismatchE)
				break
			}
		}
	}
}
