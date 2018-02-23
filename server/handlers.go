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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"

	"strconv"
	"strings"

	"github.com/terranodo/tegola/geom/encoding/geojson"
	"github.com/terranodo/tegola/geom/slippy"
	"github.com/terranodo/tegola/provider"
)

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
	featureTableInfo, err := Provider.Layers()
	if err != nil {
		panic("TODO")
	}

	ftNames := make([]string, len(featureTableInfo))
	for i, fti := range featureTableInfo {
		ftNames[i] = fti.Name()
	}
	sort.Strings(ftNames)

	layersJSON, err := json.Marshal(ftNames)
	if err != nil {
		panic("TODO")
	}
	w.Header().Set("content-type", "application/json")
	w.Write(layersJSON)
}

// --- Return the ids of features available in the named collection (layer) for current provider
func getFeatureIds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	reqQuery := r.URL.Query()
	var collectionNames []string = reqQuery["collection"]

	// No collection specified to filter on indicates all collections
	var ftNames []string
	if len(collectionNames) < 1 {
		featureTableInfo, err := Provider.Layers()
		if err != nil {
			panic("TODO")
		}
		ftNames = make([]string, len(featureTableInfo))
		for i, fti := range featureTableInfo {
			ftNames[i] = fti.Name()
		}
		sort.Strings(ftNames)
	} else {
		ftNames = collectionNames
	}
	var ids []string

	ctx := context.TODO()
	fids := []uint64{}
	collectFid := func(f *provider.Feature) error {
		fids = append(fids, f.ID)
		return nil
	}
	for _, ftn := range ftNames {
		tile := slippy.Tile{}
		err := Provider.TileFeatures(ctx, ftn, &tile, collectFid)
		sort.Slice(fids, func(i, j int) bool { return fids[i] < fids[j] })
		if err != nil {
			log.Printf("Problem collecting feature ids for '%v': %v", ftn, err)
			continue
		}
		for _, fid := range fids {
			ids = append(ids, fmt.Sprintf("%v-%v", ftn, fid))
		}
	}
	idsJSON, err := json.Marshal(ids)
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
	featureIdStr := split[len(split)-1]
	// strip off the feature id portion of the string plus the "-" separator
	collectionId := idStr[:len(idStr)-(len(featureIdStr)+1)]

	featureId, err := strconv.ParseUint(featureIdStr, 10, 64)

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "invalid 'id' parameter: '%v'"`, idStr)))
		return
	}

	// This scans the features in the indicated collection and grabs the one with 'featureId'
	// With the current Tiler interface this is the method to filter features.
	// TODO: Update Tiler interface to allow filtering

	var desiredFeature *provider.Feature
	collectGeom := func(f *provider.Feature) error {
		if f.ID == featureId {
			desiredFeature = f
			return provider.ErrCanceled
		}
		return nil
	}

	ctx := context.TODO()
	err = Provider.TileFeatures(ctx, collectionId, &slippy.Tile{}, collectGeom)

	if err != nil && err != provider.ErrCanceled {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "%v" }`, err)))
		return
	}

	gf := geojson.Feature{ID: &desiredFeature.ID, Geometry: geojson.Geometry{Geometry: desiredFeature.Geometry}}
	encoding, err := json.Marshal(gf)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.Write(encoding)
}

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

// --- Provide paged access to data for all features in requested collection
func getCollection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	q := r.URL.Query()
	var pageSize, pageNum uint64
	var collectionId string
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

	qCollectionId := q["collection"]
	if len(qCollectionId) != 1 {
		jsonError(w, "'collection' is a required parameter", 400)
		return
	} else {
		collectionId = qCollectionId[0]
	}

	resp, err := json.Marshal(struct {
		pageSize     uint64
		pageNum      uint64
		collectionId string
	}{
		pageSize:     pageSize,
		pageNum:      pageNum,
		collectionId: collectionId,
	})

	// Get collection features
	// Index of the feature currently passed to getFeatures()
	var idx uint64
	// First index we're interested in
	startIndex := pageSize * pageNum
	// Last index we're interested in +1
	stopIndex := startIndex + pageSize
	pFs := make([]*provider.Feature, 0, stopIndex-startIndex)

	getFeatures := func(f *provider.Feature) error {
		if idx >= startIndex && idx < stopIndex {
			pFs = append(pFs, f)
		} else if idx >= stopIndex {
			return provider.ErrCanceled
		}
		idx += 1
		return nil
	}

	Provider.TileFeatures(context.TODO(), collectionId, &slippy.Tile{}, getFeatures)

	// Convert the provider features to geojson features.
	gFs := make([]geojson.Feature, len(pFs))
	for i, pf := range pFs {
		gFs[i] = geojson.Feature{ID: &pf.ID, Geometry: geojson.Geometry{Geometry: pf.Geometry}}
	}
	resp, err = json.Marshal(gFs)

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
}
