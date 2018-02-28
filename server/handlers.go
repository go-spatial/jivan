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
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-spatial/go-wfs/provider"
	"github.com/go-spatial/tegola/geom/encoding/geojson"
	prv "github.com/go-spatial/tegola/provider"
	//	"github.com/terranodo/tegola/geom/slippy"
)

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

// --- Return the json-encoded OpenAPI 2 spec for the WFS API available on this instance.
func getOpenapiSpec(w http.ResponseWriter, r *http.Request) {
	var jsonSpec []byte
	var err error
	jsonSpec, err = OpenApiSpecJson()

	var status int = 200
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		jsonSpec = []byte("Error in openapi spec")
		status = 500
	}

	w.WriteHeader(status)
	w.Write(jsonSpec)
}

// --- Return the names of feature layers available in current provider
func getCollectionIds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	ftNames, err := Provider.CollectionNames()
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	layersJSON, err := json.Marshal(ftNames)
	if err != nil {
		jsonError(w, err.Error(), 500)
	}

	w.Write(layersJSON)
}

// --- Return the ids of features available in the named collection (layer) for current provider
func getFeatureIds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	reqQuery := r.URL.Query()
	var collectionNames []string = reqQuery["collection"]

	if len(collectionNames) < 1 {
		// No collection specified to filter on indicates all collections
		var err error
		collectionNames, err = Provider.CollectionNames()
		if err != nil {
			jsonError(w, err.Error(), 500)
		}
	}
	sort.Strings(collectionNames)

	fids := make([]provider.FeatureId, 0, 100)
	for _, cn := range collectionNames {
		fs, err := Provider.CollectionFeatures(cn)
		if err != nil {
			jsonError(w, err.Error(), 500)
		}
		for _, f := range fs {
			fids = append(fids, provider.FeatureId{Collection: cn, FeaturePk: f.ID})
		}
	}

	fmt.Printf("getFeatureIds len: %v\n", len(fids))
	idsJSON, err := json.Marshal(fids)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "%v"`, err)))
		return
	}

	w.Write(idsJSON)
}

// --- Return all data, esp. geometry for requested feature
func getFeature(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	reqQuery := r.URL.Query()
	idStr := reqQuery.Get("id")

	split := strings.Split(idStr, "-")
	featurePkStr := split[len(split)-1]
	// strip off the feature id portion of the string plus the "-" separator
	collectionName := idStr[:len(idStr)-(len(featurePkStr)+1)]

	featurePk, err := strconv.ParseUint(featurePkStr, 10, 64)

	featureId := provider.FeatureId{Collection: collectionName, FeaturePk: featurePk}

	featureData, err := Provider.GetFeatures([]provider.FeatureId{featureId})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "%v" }`, err)))
		return
	}
	f := featureData[0]

	gf := geojson.Feature{ID: &f.ID, Geometry: geojson.Geometry{Geometry: f.Geometry}}
	encoding, err := json.Marshal(gf)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.Write(encoding)
}

const DEFAULT_PAGE_SIZE = 10

// --- Provide paged access to data for all features in requested collection
func getCollection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	q := r.URL.Query()
	var pageSize, pageNum uint64
	var collectionName string
	var err error

	qPageSize := q["pageSize"]
	if len(qPageSize) != 1 {
		pageSize = DEFAULT_PAGE_SIZE
	} else {
		pageSize, err = strconv.ParseUint(qPageSize[0], 10, 64)
		if err != nil {
			jsonError(w, err.Error(), 400)
			return
		}
	}

	qPageNum := q["page"]
	if len(qPageNum) != 1 {
		pageNum = 0
	} else {
		pageNum, err = strconv.ParseUint(qPageNum[0], 10, 64)
		if err != nil {
			jsonError(w, err.Error(), 400)
			return
		}
	}

	qCollection := q["collection"]
	if len(qCollection) != 1 {
		jsonError(w, fmt.Sprintf("'collection' is a required parameter of length 1, got: %v", qCollection), 400)
		return
	} else {
		collectionName = qCollection[0]
	}

	log.Printf("Getting page %v (size %v) for '%v'", pageNum, pageSize, collectionName)

	resp, err := json.Marshal(struct {
		pageSize       uint64
		pageNum        uint64
		collectionName string
	}{
		pageSize:       pageSize,
		pageNum:        pageNum,
		collectionName: collectionName,
	})

	// collection features
	cFs, err := Provider.CollectionFeatures(collectionName)

	// First index we're interested in
	startIndex := pageSize * pageNum
	// Last index we're interested in +1
	stopIndex := uint64(math.Min(float64(len(cFs)), float64(startIndex+pageSize)))
	if startIndex > stopIndex {
		startIndex = stopIndex
	}

	// paged features
	pFs := make([]*prv.Feature, 0, stopIndex-startIndex)

	for _, f := range cFs[startIndex:stopIndex] {
		pFs = append(pFs, f)
	}

	// Convert the provider features to geojson features.
	gFs := make([]geojson.Feature, len(pFs))
	for i, pf := range pFs {
		gFs[i] = geojson.Feature{ID: &pf.ID, Geometry: geojson.Geometry{Geometry: pf.Geometry}, Properties: pf.Tags}
	}
	resp, err = json.Marshal(gFs)

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}

// --- Create temporary collection w/ filtered features.
// Returns a collection id for inspecting the resulting features.
func makeFeatureSet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	extentParam := q["extent"]
	collectionParam := q["collection"]

	// Grab any params besides "extent" & "collection" as attribute filters.
	propParams := make(map[string]string, len(q))
	for k, v := range r.URL.Query() {
		if k == "extent" || k == "collection" {
			continue
		}
		propParams[k] = v[0]
		if len(v) > 1 {
			log.Printf("Got multiple values for attribute filter, will only use the first '%v': %v", k, v)
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

	if len(extentParam) > 0 {
		// TODO: filter by extent
		// extent = geojson.Unmarshal(extentParam[0])
		if len(extentParam) > 1 {
			log.Printf("Multiple extent filters, will only use the first '%v'", extentParam)
		}
	}

	fIds, err := Provider.FilterFeatures(nil, collectionNames, propParams)
	newCol, err := Provider.MakeCollection("tempcol", fIds)

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	resp, err := json.Marshal(struct{ Collection string }{Collection: newCol})
	fmt.Printf("newCol / resp: %v / %v", newCol, string(resp))
	if err != nil {
		jsonError(w, err.Error(), 500)
	}
	w.WriteHeader(200)
	w.Write(resp)
}
