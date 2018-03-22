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
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-spatial/go-wfs/data_provider"
	"github.com/go-spatial/tegola/geom/encoding/geojson"
	prv "github.com/go-spatial/tegola/provider"
	"github.com/julienschmidt/httprouter"
)

// Values returned by content providers (such as root(), conformance()) need to implement this interface
type ContentUpdater interface {
	// Updates any content type references in the struct and contained structs w/ contentType
	ContentType(contentType string)
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

// -- Writes a 404 response
func notFoundError(w http.ResponseWriter) {
	jsonError(w, "not found", 404)
}

func conformanceJson(w http.ResponseWriter, r *http.Request) {
	c := conformance()
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

func rootJson(w http.ResponseWriter, r *http.Request) {
	rootContent := root()
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

// --- Return the json-encoded OpenAPI 3 spec for the WFS API available on this instance.
func openapiJson(w http.ResponseWriter, r *http.Request) {
	status := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(OpenAPI3SchemaJSON)
}

// --- Return the names of feature layers available in current provider
func collectionNames(w http.ResponseWriter, r *http.Request) {
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
func featurePks(w http.ResponseWriter, r *http.Request) {
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

	fids := make([]data_provider.FeatureId, 0, 100)
	for _, cn := range collectionNames {
		fs, err := Provider.CollectionFeatures(cn, nil)
		if err != nil {
			jsonError(w, err.Error(), 500)
		}
		for _, f := range fs {
			fids = append(fids, data_provider.FeatureId{Collection: cn, FeaturePk: f.ID})
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

	featureId := data_provider.FeatureId{Collection: collectionName, FeaturePk: featurePk}

	featureData, err := Provider.GetFeatures([]data_provider.FeatureId{featureId})
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

var collectionPathRegexp = regexp.MustCompile(`^/collection/(\w+)$`)

const DEFAULT_PAGE_SIZE = 10

func collectionMetaDataJson(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	ct := "application/json"

	cName := params.ByName("name")

	var mdI interface{}
	var err error
	if cName != "" {
		mdI, err = collectionMetaData(cName, &Provider)
	} else {
		mdI, err = collectionsMetaData(&Provider)
	}

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	var mdJson []byte
	switch md := mdI.(type) {
	case *collectionInfo:
		md.ContentType(ct)
		mdJson, err = json.Marshal(md)
	case *collectionsInfo:
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
	// fId := params.ByName("feature_id")

	w.Header().Set("content-type", ct)

	q := r.URL.Query()
	var pageSize, pageNum uint64
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

	log.Printf("Getting page %v (size %v) for '%v'", pageNum, pageSize, cName)

	resp, err := json.Marshal(struct {
		pageSize       uint64
		pageNum        uint64
		collectionName string
	}{
		pageSize:       pageSize,
		pageNum:        pageNum,
		collectionName: cName,
	})

	// collection features
	cFs, err := Provider.CollectionFeatures(cName, nil)

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

	// Wrap the features up in a FeatureCollection
	fc := geojson.FeatureCollection{
		Features: gFs,
	}
	resp, err = json.Marshal(fc)

	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	w.WriteHeader(200)
	w.Write(resp)
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
