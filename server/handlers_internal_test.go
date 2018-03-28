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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/go-spatial/go-wfs/data_provider"
	"github.com/go-spatial/go-wfs/wfs3"
	"github.com/go-spatial/tegola/provider/gpkg"
	"github.com/julienschmidt/httprouter"
)

var testingProvider data_provider.Provider

func init() {
	// Instantiate a provider from the codebase's testing gpkg.
	_, thisFilePath, _, _ := runtime.Caller(0)
	gpkgPath := path.Join(path.Dir(thisFilePath), "..", "test_data/athens-osm-20170921.gpkg")
	gpkgConfig, err := gpkg.AutoConfig(gpkgPath)
	if err != nil {
		panic(err.Error())
	}
	gpkgTiler, err := gpkg.NewTileProvider(gpkgConfig)
	if err != nil {
		panic(err.Error())
	}
	testingProvider = data_provider.Provider{Tiler: gpkgTiler}

	// @See server.go
	// This is the host:port the server expects requests at & is used to create links in responses.
	serveAddress = "server.com"
	// This is the provider the server will use for data
	Provider = testingProvider
}

func TestRoot(t *testing.T) {
	rootUrl := fmt.Sprintf("http://%v/", serveAddress)

	type TestCase struct {
		goContent          interface{}
		overrideContent    interface{}
		contentType        string
		expectedStatusCode int
	}

	testCases := []TestCase{
		// Happy path test case
		TestCase{
			goContent: &wfs3.RootContent{
				Links: []*wfs3.Link{
					&wfs3.Link{
						Href: fmt.Sprintf("http://%v/api", serveAddress),
						Rel:  "service",
					},
					&wfs3.Link{
						Href: fmt.Sprintf("http://%v/conformance", serveAddress),
						Rel:  "conformance",
					},
					&wfs3.Link{
						Href: fmt.Sprintf("http://%v/collections", serveAddress),
						Rel:  "data",
					},
				},
			},
			contentType:        "application/json",
			expectedStatusCode: 200,
		},
		// Schema error, Links type as []string instead of []wfs3.Link
		TestCase{
			goContent:          &HandlerError{Details: "response doesn't match schema"},
			overrideContent:    `{ links: ["http://doesntmatter.com"] }`,
			expectedStatusCode: 500,
		},
	}

	for i, tc := range testCases {
		var expectedBody []byte
		var err error
		// --- Collect expected response body
		switch gc := tc.goContent.(type) {
		case *wfs3.RootContent:
			gc.ContentType(tc.contentType)
			expectedBody, err = json.Marshal(gc)
			if err != nil {
				t.Errorf("Problem marshalling expected content: %v", err)
			}
		case *HandlerError:
			expectedBody, err = json.Marshal(gc)
			if err != nil {
				t.Errorf("Problem marshalling expected content: %v", err)
			}
		default:
			t.Errorf("[%v] Unexpected type in tc.goContent: %T", i, tc.goContent)
		}

		// --- override the content produced in the handler if requested by this test case
		ctx := context.TODO()
		if tc.overrideContent != nil {
			oc, err := json.Marshal(tc.overrideContent)
			if err != nil {
				t.Errorf("[%v] Problem marshalling overrideContent: %v", i, err)
			}
			ctx = context.WithValue(ctx, "overrideContent", oc)
		}

		// --- perform the request & get the response
		responseWriter := httptest.NewRecorder()
		request := httptest.NewRequest("GET", rootUrl, bytes.NewBufferString("")).WithContext(ctx)

		root(responseWriter, request)
		resp := responseWriter.Result()

		// --- check that the results match expected
		if resp.StatusCode != tc.expectedStatusCode {
			t.Errorf("[%v]: status code %v != %v", i, resp.StatusCode, tc.expectedStatusCode)
		}

		body, _ := ioutil.ReadAll(resp.Body)
		if string(body) != string(expectedBody) {
			t.Errorf("\n%v\n--- != ---\n%v", string(body), string(expectedBody))
		}
	}
}

func TestApi(t *testing.T) {
	// TODO: This is pretty circular logic, as the /api endpoint simply returns openapiSpecJson.
	//	Make a better test plan.
	expectedApiContent := wfs3.OpenAPI3SchemaJSON
	expectedStatusCode := 200
	responseWriter := httptest.NewRecorder()
	request := httptest.NewRequest("GET", fmt.Sprintf("http://%v/api", serveAddress), bytes.NewBufferString(""))

	router := httprouter.New()
	router.GET("/api", openapi)
	router.ServeHTTP(responseWriter, request)

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

	expectedBody, err := json.Marshal(wfs3.ConformanceClasses{
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
	conformance(responseWriter, request)
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

func TestCollectionsMetaData(t *testing.T) {
	collectionsUrl := fmt.Sprintf("http://%v/collections", serveAddress)
	cNames, err := testingProvider.CollectionNames()
	if err != nil {
		t.Errorf("Problem getting collection names: %v", err)
	}

	csInfo := wfs3.CollectionsInfo{Links: []*wfs3.Link{}, Collections: []*wfs3.CollectionInfo{}}
	for _, cn := range cNames {
		collectionUrl := fmt.Sprintf("http://%v/collections/%v", serveAddress, cn)
		cInfo := wfs3.CollectionInfo{Name: cn, Links: []*wfs3.Link{&wfs3.Link{Rel: "self", Href: collectionUrl, Type: "application/json"}}}
		cLink := wfs3.Link{Href: cn, Rel: "item", Type: "application/json"}

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
	collectionsMetaData(responseWriter, request)

	resp := responseWriter.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Problem reading response body: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Status code %v != %v", resp.StatusCode, expectedStatus)
	}

	if string(body) != string(expectedContent) {
		reducedOutputError(t, body, expectedContent)
	}
}

func TestSingleCollectionMetaData(t *testing.T) {
	cName := "roads_lines"
	cUrl := fmt.Sprintf("http://%v/collections/%v", serveAddress, cName)
	cInfo := wfs3.CollectionInfo{Name: cName, Links: []*wfs3.Link{&wfs3.Link{Rel: "self", Href: cUrl, Type: "application/json"}}}

	expectedStatus := 200
	expectedContent, err := json.Marshal(cInfo)
	if err != nil {
		t.Errorf("Problem marshalling expected collection info: %v", err)
	}

	responseWriter := httptest.NewRecorder()
	rctx := context.WithValue(context.TODO(), "name", cName)
	request := httptest.NewRequest("GET", cUrl, bytes.NewBufferString("")).WithContext(rctx)

	collectionMetaData(responseWriter, request)
	resp := responseWriter.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Problem reading response body: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Status code %v != %v", resp.StatusCode, expectedStatus)
	}

	if string(body) != string(expectedContent) {
		reducedOutputError(t, body, expectedContent)
	}
}

func TestCollectionFeatures(t *testing.T) {
	t.Errorf("Test not implemented")
}

func TestSingleCollectionFeature(t *testing.T) {
	t.Errorf("Test not implemented")
}

// For large human-readable returns like JSON, limit the output displayed on error to the
//	mismatched line and a few surrounding lines
func reducedOutputError(t *testing.T, body, expectedContent []byte) {
	// Number of lines preceding & following mismatched line to output
	surroundSize := 5
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
			firstLineIdx := maxInt(i-surroundSize, 0)
			lastLineIdxB := minInt(i+surroundSize, len(hrBodyLines))
			lastLineIdxE := minInt(i+surroundSize, len(hrExpectedLines))

			mismatchB := strings.Join(hrBodyLines[firstLineIdx:lastLineIdxB], "\n")
			mismatchE := strings.Join(hrExpectedLines[firstLineIdx:lastLineIdxE], "\n")
			t.Errorf("Result doesn't match expected at line %v, showing %v-%v:\n%v\n--- != ---\n%v\n",
				i, firstLineIdx, lastLineIdxB, mismatchB, mismatchE)
			break
		}
	}
}
