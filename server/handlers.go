package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
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

// --- Return the names of feature layers available in current provider (P)
func getCollectionIds(w http.ResponseWriter, r *http.Request) {
	var fts []string
	fts = P.FeatureTables()
	sort.Strings(fts)

	layersJSON, err := json.Marshal(fts)
	if err != nil {
		panic("TODO")
	}
	w.Header().Set("content-type", "application/json")
	w.Write(layersJSON)
}

// --- Return the ids of features available in the named collection (layer) for current provider (P)
func getFeatureIds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	reqQuery := r.URL.Query()
	var collectionNames []string = reqQuery["collection"]

	// No collection specified to filter on indicates all collections
	if len(collectionNames) < 1 {
		collectionNames = P.FeatureTables()
		sort.Strings(collectionNames)
	}

	var ids []string

	for _, cn := range collectionNames {
		fids, err := P.CollectionFeatureIds(cn)
		sort.Ints(fids)
		if err != nil {
			log.Printf("Problem w/ P.CollectionFeatureIds(%v): %v", cn, err)
			continue
		}
		for _, fid := range fids {
			ids = append(ids, fmt.Sprintf("%v-%v", cn, fid))
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

func getFeature(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	reqQuery := r.URL.Query()
	idStr := reqQuery.Get("id")

	split := strings.Split(idStr, "-")
	featureIdStr := split[len(split)-1]
	// strip off the feature id portion of the string plus the "-" separator
	collectionId := idStr[:len(idStr)-(len(featureIdStr)+1)]

	featureId, err := strconv.Atoi(featureIdStr)

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "invalid 'id' parameter: '%v'"`, idStr)))
		return
	}

	feature, err := P.GetFeature(collectionId, featureId)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "%v" }`, err)))
		return
	}

	w.Write(feature)
}
