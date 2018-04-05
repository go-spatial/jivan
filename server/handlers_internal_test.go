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

	"github.com/go-spatial/go-wfs/config"
	"github.com/go-spatial/go-wfs/data_provider"
	"github.com/go-spatial/go-wfs/wfs3"
	"github.com/go-spatial/tegola/geom"
	"github.com/go-spatial/tegola/geom/encoding/geojson"
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

	// This is the provider the server will use for data
	Provider = testingProvider
}

func TestServeAddress(t *testing.T) {
	type TestCase struct {
		requestHost          string
		serverConfigAddress  string
		expectedServeAddress string
	}

	testCases := []TestCase{
		{
			requestHost:          "someplace.com",
			serverConfigAddress:  "",
			expectedServeAddress: "someplace.com",
		},
		{
			requestHost:          "someplace.com",
			serverConfigAddress:  "otherplace.com",
			expectedServeAddress: "otherplace.com",
		},
	}

	originalServerAddress := config.Configuration.Server.Address
	for i, tc := range testCases {
		url := fmt.Sprintf("http://%v", tc.requestHost)
		req := httptest.NewRequest("GET", url, bytes.NewReader([]byte{}))
		if tc.serverConfigAddress != "" {
			config.Configuration.Server.Address = tc.serverConfigAddress
			// Restore the config change so other tests aren't affected.
			defer func(osa string) { config.Configuration.Server.Address = osa }(originalServerAddress)
		}
		sa := serveAddress(req)
		if sa != tc.expectedServeAddress {
			t.Errorf("[%v] serve address %v != %v", i, sa, tc.expectedServeAddress)
		}
	}
}

func TestRoot(t *testing.T) {
	serveAddress := "test.com"
	rootUrl := fmt.Sprintf("http://%v/", serveAddress)

	type TestCase struct {
		goContent          interface{}
		overrideContent    interface{}
		contentType        string
		expectedETag       string
		expectedStatusCode int
	}

	testCases := []TestCase{
		// Happy path test case
		{
			goContent: &wfs3.RootContent{
				Links: []*wfs3.Link{
					{
						Href: fmt.Sprintf("http://%v/", serveAddress),
						Rel:  "self",
					},
					{
						Href: fmt.Sprintf("http://%v/api", serveAddress),
						Rel:  "service",
					},
					{
						Href: fmt.Sprintf("http://%v/conformance", serveAddress),
						Rel:  "conformance",
					},
					{
						Href: fmt.Sprintf("http://%v/collections", serveAddress),
						Rel:  "data",
					},
				},
			},
			contentType:        JSONContentType,
			expectedETag:       "746573742e636f6dcbf29ce484222325",
			expectedStatusCode: 200,
		},
		// Schema error, Links type as []string instead of []wfs3.Link
		{
			goContent:          &HandlerError{Details: "response doesn't match schema"},
			overrideContent:    `{ links: ["http://doesntmatter.com"] }`,
			expectedStatusCode: 500,
		},
	}

	for i, tc := range testCases {
		var expectedContent []byte
		var err error
		// --- Collect expected response body
		switch gc := tc.goContent.(type) {
		case *wfs3.RootContent:
			gc.ContentType(tc.contentType)
			expectedContent, err = json.Marshal(gc)
			if err != nil {
				t.Errorf("Problem marshalling expected content: %v", err)
			}
		case *HandlerError:
			expectedContent, err = json.Marshal(gc)
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

		if tc.expectedETag != "" && (resp.Header.Get("ETag") != tc.expectedETag) {
			t.Errorf("[%v]: ETag %v != %v", i, resp.Header.Get("ETag"), tc.expectedETag)
		}

		body, _ := ioutil.ReadAll(resp.Body)
		if string(body) != string(expectedContent) {
			t.Errorf("[%v] response body doesn't match expected", i)
			reducedOutputError(t, body, expectedContent)
		}
	}
}

func TestApi(t *testing.T) {
	// TODO: This is pretty circular logic, as the /api endpoint simply returns openapiSpecJson.
	//	Make a better test plan.

	serveAddress := "unittest.net"

	type TestCase struct {
		goContent          interface{}
		overrideContent    interface{}
		contentType        string
		expectedETag       string
		expectedStatusCode int
	}

	testCases := []TestCase{
		{
			goContent:          wfs3.OpenAPI3Schema(),
			overrideContent:    nil,
			contentType:        JSONContentType,
			expectedETag:       "3b6ca0c9c15e1720",
			expectedStatusCode: 200,
		},
	}

	for i, tc := range testCases {
		var expectedContent []byte
		var err error
		if tc.contentType == JSONContentType {
			expectedContent, err = json.Marshal(tc.goContent)
			if err != nil {
				t.Errorf("[%v] problem marshalling tc.goContent to JSON: %v", i, err)
				return
			}
		} else {
			t.Errorf("[%v] unsupported content type: '%v'", i, tc.contentType)
			return
		}

		responseWriter := httptest.NewRecorder()
		rctx := context.WithValue(context.TODO(), "overrideContent", tc.overrideContent)
		request := httptest.NewRequest(
			"GET", fmt.Sprintf("http://%v/api", serveAddress), bytes.NewBufferString("")).WithContext(rctx)
		openapi(responseWriter, request)
		resp := responseWriter.Result()

		if resp.StatusCode != tc.expectedStatusCode {
			t.Errorf("[%v] status code %v != %v", i, resp.StatusCode, tc.expectedStatusCode)
		}

		if tc.expectedETag != "" && (resp.Header.Get("ETag") != tc.expectedETag) {
			t.Errorf("[%v] ETag %v != %v", i, resp.Header.Get("ETag"), tc.expectedETag)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		if string(body) != string(expectedContent) {
			t.Errorf("[%v] response content doesn't match expected:", i)
			reducedOutputError(t, body, expectedContent)
		}
	}
}

func TestConformance(t *testing.T) {
	serveAddress := "tdd.uk"
	conformanceUrl := fmt.Sprintf("http://%v/conformance", serveAddress)

	type TestCase struct {
		goContent          interface{}
		overrideContent    interface{}
		contentType        string
		expectedETag       string
		expectedStatusCode int
	}

	testCases := []TestCase{
		{
			goContent: wfs3.ConformanceClasses{
				ConformsTo: []string{
					"http://www.opengis.net/spec/wfs-1/3.0/req/core",
					"http://www.opengis.net/spec/wfs-1/3.0/req/geojson",
				},
			},
			overrideContent:    nil,
			contentType:        JSONContentType,
			expectedETag:       "4385e7a21a681d7d",
			expectedStatusCode: 200,
		},
	}

	for i, tc := range testCases {
		var expectedContent []byte
		var err error
		if tc.contentType == JSONContentType {
			expectedContent, err = json.Marshal(tc.goContent)
			if err != nil {
				t.Errorf("[%v] problem marshalling expected content to json: %v", i, err)
				return
			}
		} else {
			t.Errorf("[%v] unexpected content type: %v", i, tc.contentType)
			return
		}

		responseWriter := httptest.NewRecorder()
		rctx := context.WithValue(context.TODO(), "overrideContent", tc.overrideContent)
		request := httptest.NewRequest("GET", conformanceUrl, bytes.NewBufferString("")).WithContext(rctx)
		conformance(responseWriter, request)
		resp := responseWriter.Result()

		if resp.StatusCode != tc.expectedStatusCode {
			t.Errorf("status code %v != %v", resp.StatusCode, tc.expectedStatusCode)
		}

		if resp.Header.Get("ETag") != "" && (resp.Header.Get("ETag") != tc.expectedETag) {
			t.Errorf("[%v] ETag %v != %v", i, resp.Header.Get("ETag"), tc.expectedETag)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Problem reading response: %v", err)
		}

		if string(body) != string(expectedContent) {
			t.Errorf("[%v] response content doesn't match expected:")
			reducedOutputError(t, body, expectedContent)
		}
	}
}

func TestCollectionsMetaData(t *testing.T) {
	serveAddress := "extratesting.org:77"
	// Build the expected result
	collectionsUrl := fmt.Sprintf("http://%v/collections", serveAddress)
	cNames, err := testingProvider.CollectionNames()
	if err != nil {
		t.Errorf("Problem getting collection names: %v", err)
	}

	csInfo := wfs3.CollectionsInfo{Links: []*wfs3.Link{}, Collections: []*wfs3.CollectionInfo{}}
	for _, cn := range cNames {
		collectionUrl := fmt.Sprintf("http://%v/collections/%v", serveAddress, cn)
		cInfo := wfs3.CollectionInfo{Name: cn, Links: []*wfs3.Link{{Rel: "self", Href: collectionUrl, Type: "application/json"}}}
		cLink := wfs3.Link{Href: collectionUrl, Rel: "item", Type: "application/json"}

		csInfo.Links = append(csInfo.Links, &cLink)
		csInfo.Collections = append(csInfo.Collections, &cInfo)
	}

	type TestCase struct {
		goContent          interface{}
		overrideContent    interface{}
		contentType        string
		expectedStatusCode int
	}

	testCases := []TestCase{
		{
			goContent:          csInfo,
			overrideContent:    nil,
			contentType:        JSONContentType,
			expectedStatusCode: 200,
		},
	}

	for i, tc := range testCases {
		var expectedContent []byte
		var err error
		if tc.contentType == JSONContentType {
			expectedContent, err = json.Marshal(csInfo)
			if err != nil {
				t.Errorf("[%v] problem marshalling expected collections info to json: %v", i, err)
				return
			}
		} else {
			t.Errorf("[%v] unsupported content type: %v", i, tc.contentType)
			return
		}

		responseWriter := httptest.NewRecorder()
		rctx := context.WithValue(context.TODO(), "overrideContent", tc.overrideContent)
		request := httptest.NewRequest("GET", collectionsUrl, bytes.NewBufferString("")).WithContext(rctx)
		collectionsMetaData(responseWriter, request)

		resp := responseWriter.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("[%v] Problem reading response body: %v", i, err)
		}

		if resp.StatusCode != tc.expectedStatusCode {
			t.Errorf("[%v] Status code %v != %v", i, resp.StatusCode, tc.expectedStatusCode)
		}

		if string(body) != string(expectedContent) {
			t.Errorf("[%v] response content doesn't match expected", i)
			reducedOutputError(t, body, expectedContent)
		}
	}
}

func TestSingleCollectionMetaData(t *testing.T) {
	serveAddress := "testthis.com"

	type TestCase struct {
		goContent          interface{}
		contentOverride    interface{}
		contentType        string
		expectedStatusCode int
		urlParams          map[string]string
	}

	testCases := []TestCase{
		{
			goContent: wfs3.CollectionInfo{
				Name: "roads_lines",
				Links: []*wfs3.Link{
					{
						Rel:  "self",
						Href: fmt.Sprintf("http://%v/collections/%v", serveAddress, "roads_lines"),
						Type: JSONContentType,
					},
				},
			},
			contentOverride:    nil,
			contentType:        JSONContentType,
			expectedStatusCode: 200,
			urlParams:          map[string]string{"name": "roads_lines"},
		},
	}

	for i, tc := range testCases {
		url := fmt.Sprintf("http://%v/collections/%v", serveAddress, tc.urlParams["name"])

		var expectedContent []byte
		var err error
		if tc.contentType == JSONContentType {
			expectedContent, err = json.Marshal(tc.goContent)
			if err != nil {
				t.Errorf("[%v] Problem marshalling expected collection info: %v", i, err)
				return
			}
		} else {
			t.Errorf("[%v] Unexpected content type: %v", err, tc.contentType)
			return
		}

		responseWriter := httptest.NewRecorder()
		hrParams := make(httprouter.Params, 0, len(tc.urlParams))
		for k, v := range tc.urlParams {
			hrParams = append(hrParams, httprouter.Param{Key: k, Value: v})
		}

		request := httptest.NewRequest("GET", url, bytes.NewBufferString(""))
		rctx := context.WithValue(request.Context(), httprouter.ParamsKey, hrParams)
		rctx = context.WithValue(rctx, "contentOverride", tc.contentOverride)
		request = request.WithContext(rctx)

		collectionMetaData(responseWriter, request)
		resp := responseWriter.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("[%v] Problem reading response body: %v", err)
		}
		if resp.StatusCode != tc.expectedStatusCode {
			t.Errorf("[%v] Status code %v != %v", resp.StatusCode, tc.expectedStatusCode)
		}
		if string(body) != string(expectedContent) {
			t.Errorf("[%v] result content doesn't match expected", i)
			reducedOutputError(t, body, expectedContent)
		}
	}
}

func uint64ptr(i uint64) *uint64 {
	return &i
}

func TestCollectionFeatures(t *testing.T) {
	serveAddress := "test.com"

	type TestCase struct {
		goContent          interface{}
		contentOverride    interface{}
		contentType        string
		expectedStatusCode int
		urlParams          map[string]string
	}

	testCases := []TestCase{
		{
			goContent: geojson.FeatureCollection{
				Features: []geojson.Feature{
					{
						ID: uint64ptr(1),
						Geometry: geojson.Geometry{
							Geometry: geom.Polygon{
								{
									{23.7389198, 37.8860416},
									{23.7391532, 37.8861252},
									{23.7391844, 37.8860708},
									{23.7392575, 37.8860969},
									{23.7392784, 37.8860605},
									{23.7393112, 37.8860723},
									{23.7393746, 37.885962},
									{23.7393413, 37.8859501},
									{23.7395238, 37.8856327},
									{23.7396799, 37.8856886},
									{23.739719, 37.8856206},
									{23.739576, 37.8855694},
									{23.739825, 37.8851363},
									{23.7397731, 37.8851177},
									{23.7398198, 37.8850365},
									{23.7395535, 37.8849411},
									{23.7395002, 37.8850338},
									{23.739454, 37.8850173},
									{23.7389237, 37.8859395},
									{23.7389692, 37.8859558},
									{23.7389198, 37.8860416},
								},
							},
						},
						Properties: map[string]interface{}{
							"aeroway":    "terminal",
							"building":   "yes",
							"name":       "Ανατολικός Αερολιμένας",
							"osm_way_id": "191315051",
						},
					},
					{
						ID: uint64ptr(2),
						Geometry: geojson.Geometry{
							Geometry: geom.Polygon{
								{
									{23.7398922, 37.886111},
									{23.7403791, 37.8862854},
									{23.7407282, 37.8856782},
									{23.7405694, 37.8856213},
									{23.7403906, 37.8855572},
									{23.7402413, 37.8855038},
									{23.7398922, 37.886111},
								},
							},
						},
						Properties: map[string]interface{}{
							"aeroway":    "terminal",
							"building":   "yes",
							"osm_way_id": "191315114",
						},
					},
					{
						ID: uint64ptr(3),
						Geometry: geojson.Geometry{
							Geometry: geom.Polygon{
								{
									{23.7407222, 37.8849804},
									{23.740901, 37.8850445},
									{23.7410637, 37.8851028},
									{23.7414177, 37.884487},
									{23.7404886, 37.8841542},
									{23.7401345, 37.88477},
									{23.7407222, 37.8849804},
								},
							},
						},
						Properties: map[string]interface{}{
							"aeroway":    "terminal",
							"building":   "yes",
							"osm_way_id": "191315119",
						},
					},
					{
						ID: uint64ptr(4),
						Geometry: geojson.Geometry{
							Geometry: geom.Polygon{
								{
									{23.7393297, 37.8862976},
									{23.7392296, 37.8862617},
									{23.7392581, 37.8862122},
									{23.7385715, 37.8859662},
									{23.7384902, 37.8861076},
									{23.7391751, 37.8863529},
									{23.7391999, 37.8863097},
									{23.7393018, 37.8863462},
									{23.7393297, 37.8862976},
								},
							},
						},
						Properties: map[string]interface{}{
							"aeroway":    "terminal",
							"building":   "yes",
							"osm_way_id": "191315126",
						},
					},
					{
						ID: uint64ptr(5),
						Geometry: geojson.Geometry{
							Geometry: geom.Polygon{
								{
									{23.7400581, 37.8850307},
									{23.7400919, 37.884972},
									{23.7399529, 37.8849222},
									{23.739979, 37.8848768},
									{23.739275, 37.8846247},
									{23.7391938, 37.884766},
									{23.73991, 37.8850225},
									{23.7399314, 37.8849853},
									{23.7400581, 37.8850307},
								},
							},
						},
						Properties: map[string]interface{}{
							"aeroway":    "terminal",
							"building":   "yes",
							"osm_way_id": "191315130",
						},
					},
					{
						ID: uint64ptr(6),
						Geometry: geojson.Geometry{
							Geometry: geom.Polygon{
								{
									{23.739719, 37.8856206},
									{23.7396799, 37.8856886},
									{23.739478, 37.8860396},
									{23.7398555, 37.8861748},
									{23.7398922, 37.886111},
									{23.7402413, 37.8855038},
									{23.7402659, 37.8854609},
									{23.7402042, 37.8854388},
									{23.7398885, 37.8853257},
									{23.739719, 37.8856206},
								},
							},
						},
						Properties: map[string]interface{}{
							"aeroway":    "terminal",
							"building":   "yes",
							"osm_way_id": "191315133",
						},
					},
					{
						ID: uint64ptr(7),
						Geometry: geojson.Geometry{
							Geometry: geom.Polygon{
								{
									{23.7340727, 37.8954438},
									{23.735682, 37.892599},
									{23.735682, 37.8924297},
									{23.7355962, 37.8922857},
									{23.7354245, 37.8921503},
									{23.7335577, 37.8915322},
									{23.7318947, 37.8946225},
									{23.7340727, 37.8954438},
								},
							},
						},
						Properties: map[string]interface{}{
							"aeroway":    "apron",
							"osm_way_id": "232164874",
						},
					},
					{
						ID: uint64ptr(8),
						Geometry: geojson.Geometry{
							Geometry: geom.Polygon{
								{
									{23.6698795, 37.9390531},
									{23.6698992, 37.9390386},
									{23.6699119, 37.9390199},
									{23.6699162, 37.9389989},
									{23.6699117, 37.938978},
									{23.6698987, 37.9389593},
									{23.6698788, 37.938945},
									{23.6698541, 37.9389366},
									{23.6698272, 37.9389349},
									{23.6698011, 37.9389403},
									{23.6697787, 37.938952},
									{23.6697622, 37.9389688},
									{23.6697536, 37.9389889},
									{23.6697537, 37.9390102},
									{23.6697626, 37.9390302},
									{23.6697793, 37.9390469},
									{23.6698019, 37.9390585},
									{23.669828, 37.9390636},
									{23.6698549, 37.9390617},
									{23.6698795, 37.9390531},
								},
							},
						},
						Properties: map[string]interface{}{
							"aeroway":    "helipad",
							"osm_way_id": "265713911",
							"source":     "bing",
						},
					},
				},
			},
			contentOverride:    nil,
			contentType:        JSONContentType,
			expectedStatusCode: 200,
			urlParams: map[string]string{
				"name": "aviation_polygons",
			},
		},
	}

	for i, tc := range testCases {
		url := fmt.Sprintf("http://%v/collections/%v/items", serveAddress, tc.urlParams["name"])

		var expectedContent []byte
		var err error
		if tc.contentType == JSONContentType {
			expectedContent, err = json.Marshal(tc.goContent)
			if err != nil {
				t.Errorf("[%v] problem marshalling expected content: %v", i, err)
				return
			}
		} else {
			t.Errorf("[%v] unsupported content type for expected content: %v", i, tc.contentType)
			return
		}

		responseWriter := httptest.NewRecorder()
		request := httptest.NewRequest("GET", url, bytes.NewBufferString(""))
		rctx := request.Context()
		rctx = context.WithValue(rctx, "contentOverride", tc.contentOverride)
		hrParams := make(httprouter.Params, 0, len(tc.urlParams))
		for k, v := range tc.urlParams {
			hrp := httprouter.Param{Key: k, Value: v}
			hrParams = append(hrParams, hrp)
		}
		rctx = context.WithValue(rctx, httprouter.ParamsKey, hrParams)
		request = request.WithContext(rctx)

		collectionData(responseWriter, request)
		resp := responseWriter.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("[%v] problem reading response body: %v", i, err)
		}

		if string(body) != string(expectedContent) {
			t.Errorf("[%v] result doesn't match expected", i)
			// bBuf := bytes.NewBufferString("")
			// json.Indent(bBuf, body, "", "  ")
			// fmt.Println(bBuf)
			t.Logf("result:\n%v\n", string(body))
			reducedOutputError(t, body, expectedContent)
		}
	}
}

func TestSingleCollectionFeature(t *testing.T) {
	serveAddress := "tdd.net"

	type TestCase struct {
		goContent          interface{}
		contentOverride    interface{}
		contentType        string
		expectedStatusCode int
		urlParams          map[string]string
	}

	var i18 uint64 = 18
	testCases := []TestCase{
		{
			goContent: geojson.Feature{
				ID: &i18,
				Geometry: geojson.Geometry{
					Geometry: geom.LineString{
						{23.708656, 37.9137612},
						{23.7086007, 37.9140051},
						{23.708592, 37.9140435},
						{23.7085454, 37.914249},
					},
				},
				Properties: map[string]interface{}{
					"highway": "secondary_link",
					"osm_id":  "4380983",
					"z_index": "6",
				},
			},
			contentOverride:    nil,
			contentType:        JSONContentType,
			expectedStatusCode: 200,
			urlParams: map[string]string{
				"name":       "roads_lines",
				"feature_id": "18",
			},
		},
	}

	for i, tc := range testCases {
		url := fmt.Sprintf("http://%v/collections/%v/items/%v",
			serveAddress, tc.urlParams["name"], tc.urlParams["feature_id"])

		var expectedContent []byte
		var err error
		if tc.contentType == JSONContentType {
			expectedContent, err = json.Marshal(tc.goContent)
			if err != nil {
				t.Errorf("[%v] problem marshalling expected content: %v", i, err)
				return
			}
		} else {
			t.Errorf("[%v] unsupported content type for expected content: %v", i, tc.contentType)
			return
		}

		responseWriter := httptest.NewRecorder()
		request := httptest.NewRequest("GET", url, bytes.NewBufferString(""))
		rctx := request.Context()
		rctx = context.WithValue(rctx, "contentOverride", tc.contentOverride)
		hrParams := make(httprouter.Params, 0, len(tc.urlParams))
		for k, v := range tc.urlParams {
			hrp := httprouter.Param{Key: k, Value: v}
			hrParams = append(hrParams, hrp)
		}
		rctx = context.WithValue(rctx, httprouter.ParamsKey, hrParams)
		request = request.WithContext(rctx)

		collectionData(responseWriter, request)
		resp := responseWriter.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("[%v] problem reading response body: %v", i, err)
		}

		if string(body) != string(expectedContent) {
			t.Errorf("[%v] result doesn't match expected", i)
			// bBuf := bytes.NewBufferString("")
			// json.Indent(bBuf, body, "", "  ")
			// fmt.Println(bBuf)

			reducedOutputError(t, body, expectedContent)
		}
	}
}

// For large human-readable returns like JSON, limit the output displayed on error to the
//	mismatched line and a few surrounding lines
func reducedOutputError(t *testing.T, body, expectedContent []byte) {
	// Number of lines to output before and after mismatched line
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
