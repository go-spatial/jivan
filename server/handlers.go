package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-openapi/spec"
)

var openapiSpec spec.Swagger
var openapiSpecText string

func init() {
	openapiSpec.ID = "Go-WFS"
	openapiSpec.Swagger = "2.0"
	openapiSpec.Info = &spec.Info{}
	openapiSpec.Info.Title = "tegola-wfs"
	openapiSpec.Info.Description = "Feature query service, providing features in GeoJSON format."
	openapiSpec.Info.TermsOfService = ""
	//	openapiSpec.Info.Contact = ...
	openapiSpec.Info.License = &spec.License{
		Name: "Unchosen License",
		URL:  "",
	} // TODO: Choose a license
	openapiSpec.Info.Version = "0.0.0"

	openapiSpec.Paths = &spec.Paths{
		Paths: map[string]spec.PathItem{
			"/api/layers": spec.PathItem{},
		},
	}
	p0 := spec.PathItem{}
	p0.Get = new(spec.Operation)
	p0.Get.Description = "Provide names of layers available in API.  Features are grouped by layer."
	p0.Get.Responses = new(spec.Responses)
	p0.Get.Responses.StatusCodeResponses = make(map[int]spec.Response)
	p0.Get.Responses.StatusCodeResponses[200] = spec.Response{}
	openapiSpec.Paths.Paths["/api/layers"] = p0

	r200 := spec.Response{}
	r200.Description = "List of collection (layer) names available (JSON)."
	r200.Schema = spec.ArrayProperty(spec.StringProperty())
	r200.Examples = make(map[string]interface{})
	example := []string{"roads", "buildings", "waterways"}
	r200.Examples["application/json"] = example
	p0.Get.Responses.StatusCodeResponses[200] = r200

	//	openapiSpec.Paths.Paths["/api/layers"] = p0
	//	json, e = openapiSpec.Paths.MarshalJSON()

	//	openapiSpecText = `{
	//	    "openapi": "3.0.0",
	//	    "info": {
	//	        "title": "tegola-wfs",
	//	        "description": "Feature query service, providing features in GeoJSON format.",
	//	        "contact": "TODO: What's a good contact email?",
	//	        "version": "0.0.0"
	//	    },
	//	    "paths": {
	//	        "/api/layers": {
	//	            "get": {
	//	                "summary": "Provide layer names available in API.  Features are grouped by layer.",
	//	                "operationId": "Why do we need this?",
	//	                "tags": "Why do we need this?",
	//	                "responses": {
	//	                    "200": {
	//	                        "description": "List of layer names available.",
	//	                        "content": {
	//	                            "application/json": {
	//	                                "schema": {
	//	                                    "type": "array",
	//	                                    "items": {
	//	                                        "type": "string"
	//	                                    }
	//	                                }
	//	                            }
	//	                        }
	//	                    }
	//	                }
	//	            }
	//	        },
	//	        "/api/layer/{layerName}": {
	//	            "get": {
	//	                "summary": "Provides the names of features available in layer.",
	//	                "operationId": "needed?",
	//	                "tags": "needed?",
	//	                "parameters": [
	//	                    {
	//	                        "name": "layerName",
	//	                        "in": "path",
	//	                        "type": "string"
	//	                    }
	//	                ],
	//	                "responses": {
	//	                    "200": {
	//	                        "description": "List of features available for layer.",
	//	                        "content": {
	//	                            "application/json": {
	//	                                "schema": {
	//	                                    "type": "array",
	//	                                    "items": {
	//	                                        "type": "string"
	//	                                    }
	//	                                }
	//	                            }
	//	                        }
	//	                    },
	//	                    "400": {
	//	                        "description": "layerName not recognized"
	//	                    }
	//	                }
	//	            }
	//	        },
	//	        "/api/feature/{featureName}": {
	//	            "get": {
	//	                "summary": "Provides the feature named including geometry in GeoJSON format.",
	//	                "operationId": "needed?",
	//	                "tags": "needed?",
	//	                "parameters": [
	//	                    {
	//	                        "name": "featureName",
	//	                        "in": "path",
	//	                        "type": "string"
	//	                    }
	//	                ],
	//	                "responses": {
	//	                    "200": {
	//	                        "description": "Feature object w/ name & geometry",
	//	                        "content": {
	//	                            "application/json": {
	//	                                "schema": {
	//	                                    "properties": {
	//	                                        "name": "string",
	//	                                        "geom": "string",
	//	                                        "__comment__": "geom is in GeoJSON format"
	//	                                    }
	//	                                }
	//	                            }
	//	                        }
	//	                    },
	//	                    "400": {
	//	                        "description": "layerName not recognized"
	//	                    }
	//	                }
	//	            }
	//	        }
	//	    }
	//	}
	//	`
	//		err := json.Unmarshal([]byte(openapiSpecText), &openapiSpec)
	//	if err != nil {
	//		panic(fmt.Sprintf("Problem unmarshalling openAPI spec text: %v", err))
	//	}
}

// --- Return the json-encoded OpenAPI 2 spec for the WFS API available on this instance.
func getOpenapiSpec(w http.ResponseWriter, r *http.Request) {
	jsonString, err := json.MarshalIndent(openapiSpec, "", "    ")
	var status int = 200
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		jsonString = []byte("Error in openapi spec")
		status = 500
	}

	w.WriteHeader(status)
	w.Write(jsonString)
}

// --- Return the names of feature layers available in current provider (P)
func getLayers(w http.ResponseWriter, r *http.Request) {
	layersJSON, err := json.Marshal(P.FeatureTables())
	if err != nil {
		panic("TODO")
	}
	w.Header().Set("content-type", "application/json")
	w.Write(layersJSON)
}

// --- Return the ids of features available in the named collection (layer) for current provider (P)
func getLayerFeatures(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	reqQuery := r.URL.Query()
	layerName := reqQuery.Get("name")

	if len(layerName) < 1 {
		w.WriteHeader(400)
		w.Write([]byte(`{ "detail": "'name' parameter required but not provided"}`))
		return
	}

	ids, err := P.CollectionFeatureIds(layerName)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "%v" }`, err)))
		return
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
	reqQuery := r.URL.Query()
	collName := reqQuery.Get("collection")
	idStr := reqQuery.Get("id")
	featureId, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "invalid 'id' parameter: '%v'"`, idStr)))
	}

	w.Header().Set("content-type", "application/json")
	if len(collName) < 1 {
		w.WriteHeader(400)
		w.Write([]byte(`{ "detail": "'collection' parameter required but not provided"`))
		return
	}

	feature, err := P.GetFeature(collName, featureId)

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(`{ "detail": "%v" }`, err)))
		return
	}

	w.Write(feature)
}
