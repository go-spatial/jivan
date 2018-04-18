///////////////////////////////////////////////////////////////////////////////
//
// The MIT License (MIT)
// Copyright (c) 2018 Jivan Amara
// Copyright (c) 2018 Tom Kralidis
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

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-spatial/go-wfs/config"
	"github.com/go-spatial/go-wfs/data_provider"
	"github.com/go-spatial/go-wfs/wfs3"
	"github.com/go-spatial/tegola/geom"
	"github.com/julienschmidt/httprouter"
)

// This is the default max number of features to return for feature collection reqeusts
const DEFAULT_RESULT_LIMIT = 10

const (
	JSONContentType = "application/json"
	HTMLContentType = "text/html" // Not yet supported
)

const (
	HTTPStatusOk          = 200
	HTTPStatusNotModified = 304
	HTTPStatusServerError = 500
	HTTPStatusClientError = 400

	HTTPMethodGET  = "GET"
	HTTPMethodHEAD = "HEAD"
)

type HandlerError struct {
	Details string `json:"detail"`
}

func serveAddress(r *http.Request) string {
	psh := config.Configuration.Server.URLHostPort
	if psh == "" {
		psh = r.URL.Host
	}

	psh = strings.TrimRight(psh, "/")

	psh = fmt.Sprintf("%v://%v%v", config.Configuration.Server.URLScheme, psh, strings.TrimRight(config.Configuration.Server.URLBasePath, "/"))

	return psh
}

// contentType() returns the Content-Type string that will be used for the response to this request.
// This Content-Type will be chosen in order of increasing priority from:
// request Content-Type, request Accept
// If the type chosen from the request isn't supported, defaultContentType will be used.
// TODO: Move defaultContentType to configuration.
func supportedContentType(ct string) bool {
	supportedContentTypes := []string{JSONContentType}
	typeSupported := false
	for _, sct := range supportedContentTypes {
		if ct == sct {
			typeSupported = true
			break
		}
	}
	return typeSupported
}

func contentType(r *http.Request) string {
	defaultContentType := JSONContentType
	useType := ""
	ctType := r.Header.Get("Content-Type")
	acceptTypes := r.Header.Get("Accept")

	if supportedContentType(ctType) {
		useType = ctType
	}

	// TODO: Parse acceptTypes properly
	acceptTypes = acceptTypes

	if !supportedContentType(useType) {
		useType = defaultContentType
	}

	return useType
}

// Sets response 'status', and writes a json-encoded object with property "detail" having value "msg".
func jsonError(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)

	result, err := json.Marshal(struct {
		Detail string `json:"detail"`
	}{
		Detail: msg,
	})

	if err != nil {
		w.Write([]byte(fmt.Sprintf("problem marshaling error: %v", msg)))
	} else {
		w.Write(result)
	}
}

// Serves the root content for WFS3.
func root(w http.ResponseWriter, r *http.Request) {
	rPath := "/"
	// This allows tests to set the result to whatever they want.
	overrideContent := r.Context().Value("overrideContent")

	rootContent, contentId := wfs3.Root(serveAddress(r), false)
	w.Header().Set("ETag", contentId)
	if r.Method == HTTPMethodHEAD {
		if r.Header.Get("ETag") == contentId {
			w.WriteHeader(HTTPStatusNotModified)
		} else {
			w.WriteHeader(HTTPStatusOk)
		}
		return
	}

	ct := contentType(r)
	rootContent.ContentType(ct)

	var encodedContent []byte
	var err error
	if ct == JSONContentType {
		encodedContent, err = json.Marshal(rootContent)
	} else {
		jsonError(w, "Content-Type: '"+ct+"' not supported.", HTTPStatusServerError)
		return
	}

	if err != nil {
		jsonError(w, err.Error(), HTTPStatusServerError)
		return
	}

	w.Header().Set("Content-Type", ct)

	if overrideContent != nil {
		encodedContent = overrideContent.([]byte)
	}
	respBodyRC := ioutil.NopCloser(bytes.NewReader(encodedContent))
	err = wfs3.ValidateJSONResponse(r, rPath, HTTPStatusOk, w.Header(), respBodyRC)
	if err != nil {
		log.Printf("%v", err)
		jsonError(w, "response doesn't match schema", HTTPStatusServerError)
		return
	}

	w.WriteHeader(HTTPStatusOk)
	w.Write(encodedContent)
}

func conformance(w http.ResponseWriter, r *http.Request) {
	cPath := "/conformance"
	// This allows tests to set the result to whatever they want.
	overrideContent := r.Context().Value("overrideContent")

	ct := contentType(r)
	c, contentId := wfs3.Conformance()
	w.Header().Set("ETag", contentId)
	if r.Method == HTTPMethodHEAD {
		if r.Header.Get("ETag") == contentId {
			w.WriteHeader(HTTPStatusNotModified)
		} else {
			w.WriteHeader(HTTPStatusOk)
		}
		return
	}

	var encodedContent []byte
	var err error
	if ct == JSONContentType {
		encodedContent, err = json.Marshal(c)
	} else {
		jsonError(w, "Content-Type: ''"+ct+"'' not supported.", HTTPStatusServerError)
		return
	}

	if err != nil {
		msg := fmt.Sprintf("problem marshaling conformance declaration to %v: %v", ct, err.Error())
		jsonError(w, msg, HTTPStatusServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if overrideContent != nil {
		encodedContent = overrideContent.([]byte)
	}
	respBodyRC := ioutil.NopCloser(bytes.NewReader(encodedContent))
	err = wfs3.ValidateJSONResponse(r, cPath, HTTPStatusOk, w.Header(), respBodyRC)
	if err != nil {
		log.Printf(fmt.Sprintf("%v", err))
		jsonError(w, "response doesn't match schema", HTTPStatusServerError)
		return
	}

	w.WriteHeader(HTTPStatusOk)
	w.Write(encodedContent)
}

// --- Return the json-encoded OpenAPI 3 spec for the WFS API available on this instance.
func openapi(w http.ResponseWriter, r *http.Request) {
	// --- TODO: Disabled due to #34
	// oapiPath := "/api"
	// This allows tests to set the result to whatever they want.
	overrideContent := r.Context().Value("overrideContent")

	ct := contentType(r)

	if ct != JSONContentType {
		jsonError(w, "Content-Type: ''"+ct+"'' not supported.", HTTPStatusServerError)
		return
	}
	encodedContent, contentId := wfs3.OpenAPI3SchemaEncoded(ct)
	w.Header().Set("ETag", contentId)

	if r.Method == HTTPMethodHEAD {
		if r.Header.Get("ETag") == contentId {
			w.WriteHeader(HTTPStatusNotModified)
		} else {
			w.WriteHeader(HTTPStatusOk)
		}
		return
	}

	w.Header().Set("Content-Type", ct)

	if overrideContent != nil {
		encodedContent = overrideContent.([]byte)
	}

	// TODO: As of 2018-04-05 I can't find a reliable openapi3 document schema.  When one is published use if for validation here.
	// if ct == JSONContentType {
	// 	err := wfs3.ValidateJSONResponseAgainstJSONSchema(encodedContent, jsonSchema)
	// 	if err != nil {
	// 		log.Printf(fmt.Sprintf("%v", err))
	// 		jsonError(w, "response doesn't match schema", HTTPStatusServerError)
	// 		return
	// 	}
	// } else {
	// 	msg := fmt.Sprintf("unsupported content type: %v", ct)
	// 	log.Printf(msg)
	// 	jsonError(w, msg, HTTPStatusClientError)
	// }

	w.WriteHeader(HTTPStatusOk)
	w.Write(encodedContent)
}

func collectionMetaData(w http.ResponseWriter, r *http.Request) {
	cmdPath := "/collections/{name}"
	overrideContent := r.Context().Value("overrideContent")

	ct := contentType(r)
	ps := httprouter.ParamsFromContext(r.Context())

	cName := ps.ByName("name")
	if cName == "" {
		jsonError(w, "No {name} provided", HTTPStatusClientError)
		return
	}

	md, contentId, err := wfs3.CollectionMetaData(cName, &Provider, serveAddress(r), false)
	if err != nil {
		jsonError(w, err.Error(), HTTPStatusServerError)
		return
	}

	w.Header().Set("ETag", contentId)
	if r.Method == HTTPMethodHEAD {
		if r.Header.Get("ETag") == contentId {
			w.WriteHeader(HTTPStatusNotModified)
		} else {
			w.WriteHeader(HTTPStatusOk)
		}
		return
	}

	var encodedContent []byte
	if ct == JSONContentType {
		md.ContentType(ct)
		encodedContent, err = json.Marshal(md)
	} else {
		jsonError(w, "Content-Type: ''"+ct+"'' not supported.", HTTPStatusServerError)
		return
	}

	if err != nil {
		jsonError(w, err.Error(), HTTPStatusServerError)
		return
	}

	w.Header().Set("Content-Type", ct)

	if overrideContent != nil {
		encodedContent = overrideContent.([]byte)
	}

	respBodyRC := ioutil.NopCloser(bytes.NewReader(encodedContent))
	err = wfs3.ValidateJSONResponse(r, cmdPath, HTTPStatusOk, w.Header(), respBodyRC)
	if err != nil {
		log.Printf(fmt.Sprintf("%v", err))
		jsonError(w, "response doesn't match schema", HTTPStatusServerError)
		return
	}

	w.WriteHeader(HTTPStatusOk)
	w.Write(encodedContent)
}

func collectionsMetaData(w http.ResponseWriter, r *http.Request) {
	cmdPath := "/collections"
	overrideContent := r.Context().Value("overrideContent")

	ct := contentType(r)
	md, contentId, err := wfs3.CollectionsMetaData(&Provider, serveAddress(r), false)
	if err != nil {
		jsonError(w, err.Error(), HTTPStatusServerError)
		return
	}

	w.Header().Set("ETag", contentId)
	if r.Method == HTTPMethodHEAD {
		if r.Header.Get("ETag") == contentId {
			w.WriteHeader(HTTPStatusNotModified)
		} else {
			w.WriteHeader(HTTPStatusOk)
		}
		return
	}

	var encodedContent []byte
	if ct == JSONContentType {
		md.ContentType(JSONContentType)
		encodedContent, err = json.Marshal(md)
	} else {
		jsonError(w, "Content-Type: ''"+ct+"'' not supported.", HTTPStatusServerError)
		return
	}

	if err != nil {
		jsonError(w, err.Error(), HTTPStatusServerError)
		return
	}

	w.Header().Set("Content-Type", ct)

	if overrideContent != nil {
		encodedContent = overrideContent.([]byte)
	}

	respBodyRC := ioutil.NopCloser(bytes.NewReader(encodedContent))
	err = wfs3.ValidateJSONResponse(r, cmdPath, HTTPStatusOk, w.Header(), respBodyRC)
	if err != nil {
		log.Printf(fmt.Sprintf("%v", err))
		jsonError(w, "response doesn't match schema", HTTPStatusServerError)
		return
	}

	w.WriteHeader(HTTPStatusOk)
	w.Write(encodedContent)
}

// --- Provide paged access to data for all features at /collections/{name}/items/{feature_id}
func collectionData(w http.ResponseWriter, r *http.Request) {
	ct := contentType(r)
	overrideContent := r.Context().Value("overrideContent")

	urlParams := httprouter.ParamsFromContext(r.Context())
	cName := urlParams.ByName("name")
	fidStr := urlParams.ByName("feature_id")
	var fid uint64
	var err error
	if fidStr != "" {
		cid, err := strconv.Atoi(fidStr)
		if err != nil {
			jsonError(w, "Invalid feature_id: "+fidStr, HTTPStatusClientError)
		}
		fid = uint64(cid)
	}

	q := r.URL.Query()
	var limit, pageNum uint
	var timeprops map[string]string

	qPageSize := q["limit"]
	if len(qPageSize) != 1 {
		limit = DEFAULT_RESULT_LIMIT
	} else {
		ps, err := strconv.ParseUint(qPageSize[0], 10, 64)
		if err != nil {
			jsonError(w, err.Error(), HTTPStatusClientError)
			return
		}
		limit = uint(ps)
	}

	qPageNum := q["page"]
	if len(qPageNum) != 1 {
		pageNum = 0
	} else {
		pn, err := strconv.ParseUint(qPageNum[0], 10, 64)
		if err != nil {
			jsonError(w, err.Error(), HTTPStatusClientError)
			return
		}
		pageNum = uint(pn)
	}

	qTime := q["time"]
	if len(qTime) > 0 {
		if len(qTime) > 1 {
			jsonError(w, "'time' parameter used more than once'", 400)
			return
		}
		ts := strings.Split(qTime[0], "/")
		timeprops = make(map[string]string)
		if len(ts) == 1 {
			timeprops["timestamp"] = ts[0]
		} else if len(ts) == 2 {
			timeprops["start_time"] = ts[0]
			timeprops["stop_time"] = ts[1]
		} else {
			jsonError(w, "'time' parameter contains more than two time values ('/' separator)", HTTPStatusClientError)
			return
		}
	}

	var data interface{}
	var jsonSchema string
	// Hex string hash of content
	var contentId string
	// Indicates if there is more data available from stopIdx onward
	var more bool
	// If a feature_id was provided, get a single feature, otherwise get a feature collection
	//	containing all of the collection's features
	if fidStr != "" {
		data, contentId, err = wfs3.FeatureData(cName, fid, &Provider, false)
		jsonSchema = wfs3.FeatureJSONSchema
	} else {
		// First index we're interested in
		startIdx := limit * pageNum
		// Last index we're interested in +1
		stopIdx := startIdx + limit

		data, more, contentId, err = wfs3.FeatureCollectionData(cName, startIdx, stopIdx, timeprops, &Provider, false)
		jsonSchema = wfs3.FeatureCollectionJSONSchema
	}

	if err != nil {
		var sc int
		var msg string
		switch e := err.(type) {
		case *data_provider.BadTimeString:
			msg = e.Error()
			sc = HTTPStatusClientError
		default:
			msg = fmt.Sprintf("Problem collecting feature data: %v", e)
			sc = HTTPStatusServerError
		}
		jsonError(w, msg, sc)
		return
	}

	w.Header().Set("ETag", contentId)
	if r.Method == HTTPMethodHEAD {
		if r.Header.Get("ETag") == contentId {
			w.WriteHeader(HTTPStatusNotModified)
		} else {
			w.WriteHeader(HTTPStatusOk)
		}
		return
	}

	var encodedContent []byte
	switch d := data.(type) {
	case *wfs3.Feature:
		if ct == JSONContentType {
			// Generate self link
			d.Self = r.URL.String()
			d.Collection = fmt.Sprintf("http://%v/collections/%v", r.URL.Host, cName)
			encodedContent, err = json.Marshal(d)
		} else {
			jsonError(w, "Content-Type: ''"+ct+"'' not supported.", HTTPStatusServerError)
			return
		}
	case *wfs3.FeatureCollection:
		// Generate self, previous, and next links
		self := fmt.Sprintf("http://%v%v?page=%v&limit=%v", r.URL.Host, r.URL.Path, pageNum, limit)
		var prev string
		var next string
		if pageNum > 0 {
			prev = fmt.Sprintf("http://%v%v?page=%v&limit=%v", r.URL.Host, r.URL.Path, pageNum-1, limit)
		}
		if more {
			next = fmt.Sprintf("http://%v%v?page=%v&limit=%v", r.URL.Host, r.URL.Path, pageNum+1, limit)
		}

		if ct == JSONContentType {
			d.Self = self
			d.Prev = prev
			d.Next = next
			encodedContent, err = json.Marshal(d)
		} else {
			jsonError(w, "Content-Type: ''"+ct+"'' not supported.", HTTPStatusServerError)
			return
		}
	default:
		msg := fmt.Sprintf("Unexpected feature data type: %T, %v", data, data)
		jsonError(w, msg, HTTPStatusServerError)
		return
	}

	if err != nil {
		msg := fmt.Sprintf("Problem marshalling feature data: %v", err)
		jsonError(w, msg, HTTPStatusServerError)
	}

	w.Header().Set("Content-Type", ct)

	if overrideContent != nil {
		encodedContent = overrideContent.([]byte)
	}

	if ct == JSONContentType {
		err = wfs3.ValidateJSONResponseAgainstJSONSchema(encodedContent, jsonSchema)
		if err != nil {
			log.Printf(fmt.Sprintf("%v", err))
			jsonError(w, "response doesn't match schema", HTTPStatusServerError)
			return
		}
	} else {
		msg := fmt.Sprintf("unsupported content type: %v", ct)
		log.Printf(msg)
		jsonError(w, msg, HTTPStatusClientError)
	}

	w.WriteHeader(HTTPStatusOk)
	w.Write(encodedContent)
}

// --- Create temporary collection w/ filtered features.
// Returns a collection id for inspecting the resulting features.
func filteredFeatures(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	extentParam := q["extent"]
	collectionParam := q["collection"]

	// Grab any params besides "extent" & "collection" as property filters.
	propParams := make(map[string]string, len(q))
	for k, v := range r.URL.Query() {
		if k == "extent" || k == "collection" {
			continue
		}
		propParams[k] = v[0]
		if len(v) > 1 {
			log.Printf("Got multiple values for property filter, will only use the first '%v': %v", k, v)
		}
	}

	var collectionNames []string
	if len(collectionParam) > 0 {
		collectionNames = collectionParam
	} else {
		var err error
		collectionNames, err = Provider.CollectionNames()
		if err != nil {
			jsonError(w, err.Error(), HTTPStatusServerError)
		}
	}

	var extent geom.Extent
	if len(extentParam) > 0 {
		// lat/lon bounding box arranged as [<minx>, <miny>, <maxx>, <maxy>]
		var llbbox [4]float64
		err := json.Unmarshal([]byte(extentParam[0]), &llbbox)
		if err != nil {
			jsonError(w, fmt.Sprintf("unable to unmarshal extent (%v) due to error: %v", extentParam[0], err), HTTPStatusClientError)
			return
		}
		extent = geom.Extent{llbbox[0], llbbox[1], llbbox[2], llbbox[3]}
		// TODO: filter by extent
		if len(extentParam) > 1 {
			log.Printf("Multiple extent filters, will only use the first '%v'", extentParam)
		}
	}

	fids, err := Provider.FilterFeatures(&extent, collectionNames, propParams)
	newCol, err := Provider.MakeCollection("tempcol", fids)

	if err != nil {
		jsonError(w, err.Error(), HTTPStatusServerError)
		return
	}

	resp, err := json.Marshal(struct {
		Collection   string
		FeatureCount int
	}{Collection: newCol, FeatureCount: len(fids)})
	if err != nil {
		jsonError(w, err.Error(), HTTPStatusServerError)
	}
	w.WriteHeader(HTTPStatusOk)
	w.Write(resp)
}
