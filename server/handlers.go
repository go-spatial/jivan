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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-spatial/go-wfs/wfs3"
	"github.com/go-spatial/tegola/geom/encoding/geojson"
	"github.com/julienschmidt/httprouter"
)

const DEFAULT_PAGE_SIZE = 10

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

func rootJson(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	rootContent := wfs3.Root(serveAddress)
	ct := "application/json"
	rootContent.ContentType(ct)
	rJson, err := json.Marshal(rootContent)
	w.Header().Set("Content-Type", ct)

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(rJson)
}

func conformanceJson(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c := wfs3.Conformance()
	result, err := json.Marshal(c)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		jsonError(w, "problem marshaling conformance declaration to json", 500)
		return
	} else {
		w.WriteHeader(200)
		w.Write([]byte(result))
	}
}

// --- Return the json-encoded OpenAPI 3 spec for the WFS API available on this instance.
func openapiJson(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	status := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(wfs3.OpenAPI3SchemaJSON)
}

func collectionMetaDataJson(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	ct := "application/json"

	cName := params.ByName("name")

	var mdI interface{}
	var err error
	if cName != "" {
		mdI, err = wfs3.CollectionMetaData(cName, &Provider, serveAddress)
	} else {
		mdI, err = wfs3.CollectionsMetaData(&Provider, serveAddress)
	}

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var mdJson []byte
	switch md := mdI.(type) {
	case *wfs3.CollectionInfo:
		md.ContentType(ct)
		mdJson, err = json.Marshal(md)
	case *wfs3.CollectionsInfo:
		md.ContentType(ct)
		mdJson, err = json.Marshal(md)
	default:
		msg := fmt.Sprintf("Got an unexpected metadata type: %T, %v", md, md)
		jsonError(w, msg, 500)
		return
	}

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", ct)
	w.WriteHeader(200)
	w.Write(mdJson)
}

// --- Provide paged access to data for all features at /collections/{name}/items/{feature_id}
func collectionDataJson(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	ct := "application/json"
	cName := params.ByName("name")
	fIdStr := params.ByName("feature_id")
	var fId uint64
	var err error
	if fIdStr != "" {
		cid, err := strconv.Atoi(fIdStr)
		if err != nil {
			jsonError(w, "Invalid feature_id: "+fIdStr, 400)
		}
		fId = uint64(cid)
	}

	q := r.URL.Query()
	var pageSize, pageNum uint

	qPageSize := q["pageSize"]
	if len(qPageSize) != 1 {
		pageSize = DEFAULT_PAGE_SIZE
	} else {
		ps, err := strconv.ParseUint(qPageSize[0], 10, 64)
		if err != nil {
			jsonError(w, err.Error(), 400)
			return
		}
		pageSize = uint(ps)
	}

	qPageNum := q["page"]
	if len(qPageNum) != 1 {
		pageNum = 0
	} else {
		pn, err := strconv.ParseUint(qPageNum[0], 10, 64)
		if err != nil {
			jsonError(w, err.Error(), 400)
			return
		}
		pageNum = uint(pn)
	}

	log.Printf("Getting page %v (size %v) for '%v'", pageNum, pageSize, cName)

	var data interface{}
	// If a feature_id was provided, get a single feature, otherwise get a feature collection
	//	containing all of the collection's features
	if fIdStr != "" {
		data, err = wfs3.Feature(cName, fId, &Provider)
	} else {
		// First index we're interested in
		startIdx := pageSize * pageNum
		// Last index we're interested in +1
		stopIdx := startIdx + pageSize

		data, err = wfs3.FeatureCollection(cName, startIdx, stopIdx, &Provider)
	}

	if err != nil {
		msg := fmt.Sprintf("Problem collecting feature data: %v", err)
		jsonError(w, msg, 500)
		return
	}

	var dataJson []byte
	switch d := data.(type) {
	case *geojson.Feature:
		dataJson, err = json.Marshal(d)
	case *geojson.FeatureCollection:
		dataJson, err = json.Marshal(d)
	default:
		msg := fmt.Sprintf("Unexpected feature data type: %T, %v", data, data)
		jsonError(w, msg, 500)
		return
	}

	if err != nil {
		msg := fmt.Sprintf("Problem marshalling feature data: %v", err)
		jsonError(w, msg, 500)
	}

	w.Header().Set("Content-Type", ct)
	w.WriteHeader(200)
	w.Write(dataJson)
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
			jsonError(w, err.Error(), 500)
		}
	}

	var extent [2][2]float64
	if len(extentParam) > 0 {
		// lat/lon bounding box arranged as [<minx>, <miny>, <maxx>, <maxy>]
		var llbbox [4]float64
		err := json.Unmarshal([]byte(extentParam[0]), &llbbox)
		if err != nil {
			jsonError(w, fmt.Sprintf("unable to unmarshal extent (%v) due to error: %v", extentParam[0], err), 400)
			return
		}
		extent = [2][2]float64{{llbbox[0], llbbox[1]}, {llbbox[2], llbbox[3]}}
		// TODO: filter by extent
		if len(extentParam) > 1 {
			log.Printf("Multiple extent filters, will only use the first '%v'", extentParam)
		}
	}

	fIds, err := Provider.FilterFeatures(&extent, collectionNames, propParams)
	newCol, err := Provider.MakeCollection("tempcol", fIds)

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	resp, err := json.Marshal(struct {
		Collection   string
		FeatureCount int
	}{Collection: newCol, FeatureCount: len(fIds)})
	if err != nil {
		jsonError(w, err.Error(), 500)
	}
	w.WriteHeader(200)
	w.Write(resp)
}
