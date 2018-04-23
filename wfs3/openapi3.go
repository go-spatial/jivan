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

// go-wfs project openapi3.go

package wfs3

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"

	"github.com/go-spatial/go-wfs/config"
	"github.com/jban332/kin-openapi/openapi3"
)

var openAPI3Schema *openapi3.Swagger
var openAPI3SchemaContentId string
var openAPI3SchemaJSON []byte

func OpenAPI3Schema() *openapi3.Swagger {
	if openAPI3Schema == nil {
		GenerateOpenAPIDocument()
	}
	return openAPI3Schema
}

func OpenAPI3SchemaEncoded(encoding string) (encodedContent []byte, contentId string) {
	if openAPI3SchemaJSON == nil {
		GenerateOpenAPIDocument()
	}

	if encoding == "application/json" {
		return openAPI3SchemaJSON, openAPI3SchemaContentId
	} else {
		msg := fmt.Sprintf("Encoding not supported: %v", encoding)
		panic(msg)
	}
}

func GenerateOpenAPIDocument() {
	openAPI3Schema = &openapi3.Swagger{
		OpenAPI: "3.0.0",
		Info: openapi3.Info{
			Title:       config.Configuration.Metadata.Identification.Title,
			Description: config.Configuration.Metadata.Identification.Description,
			Version:     "0.0.1",
			License: &openapi3.License{
				Name: "MIT",
				URL:  "http://opensource.org/licenses/MIT",
			},
		},
		Paths: openapi3.Paths{
			"/": &openapi3.PathItem{
				Summary:     "top-level endpoints available",
				Description: "Root of API, all metadata & services are beneath these links",
				Get: &openapi3.Operation{
					OperationID: "getRoot",
					Parameters:  openapi3.Parameters{},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Ref: "",
							Value: &openapi3.Response{
								Content: openapi3.NewContentWithJSONSchema(&RootContentSchema),
							},
						},
					},
				},
			},
			"/api": &openapi3.PathItem{
				Summary:     "api definition",
				Description: "OpenAPI 3.0 definition of this WFS 3.0 service",
				Get: &openapi3.Operation{
					OperationID: "getAPI",
					Parameters:  openapi3.Parameters{},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							// TODO: There isn't an official json schema for openaip3 yet.
							// The best I can do as of 2018-03-30 is a json schema schema
							Ref: "http://json-schema.org/draft-07/schema",
						},
					},
				},
			},
			"/conformance": &openapi3.PathItem{
				Summary:     "Conformance classes",
				Description: "Functionality requirements this api conforms to.",
				Get: &openapi3.Operation{
					OperationID: "getConformance",
					Parameters:  openapi3.Parameters{},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Value: &openapi3.Response{
								Content: openapi3.NewContentWithJSONSchema(&ConformanceClassesSchema),
							},
						},
					},
				},
			},
			"/collections": &openapi3.PathItem{
				Summary:     "Feature collection metadata",
				Description: "Provides details about all feature collections served",
				Get: &openapi3.Operation{
					OperationID: "getCollectionsMetaData",
					Parameters: openapi3.Parameters{
						&openapi3.ParameterRef{
							Value: &openapi3.Parameter{
								Description:     "Name of collection to retrieve metadata for.",
								Name:            "name",
								In:              "path",
								Required:        false,
								Schema:          &openapi3.SchemaRef{Value: openapi3.NewStringSchema()},
								AllowEmptyValue: true,
							},
						},
					},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Value: &openapi3.Response{
								// TODO: openapi3.NewContentWithJSONSchema() would help, but is broken
								Content: openapi3.Content{
									"application/json": &openapi3.ContentType{
										Schema: &openapi3.SchemaRef{
											Value: &CollectionsInfoSchema,
										},
									},
								},
							},
						},
					},
				},
			},
			"/collections/{name}": &openapi3.PathItem{
				Summary:     "Feature collection metadata",
				Description: "Provides details about the feature collection named",
				Get: &openapi3.Operation{
					OperationID: "getCollectionMetaData",
					Parameters: openapi3.Parameters{
						&openapi3.ParameterRef{
							Value: &openapi3.Parameter{
								Description:     "Name of collection to retrieve metadata for.",
								Name:            "name",
								In:              "path",
								Required:        true,
								Schema:          &openapi3.SchemaRef{Value: openapi3.NewStringSchema()},
								AllowEmptyValue: false,
							},
						},
					},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Value: &openapi3.Response{
								// TODO: openapi3.NewContentWithJSONSchema() would help, but is broken
								//
								Content: openapi3.Content{
									"application/json": &openapi3.ContentType{
										Schema: &openapi3.SchemaRef{
											Value: &CollectionInfoSchema,
										},
									},
								},
							},
						},
					},
				},
			},
			"/collections/{name}/items": &openapi3.PathItem{
				Summary:     "Feature data for collection",
				Description: "Provides paged access to data for all features in collection",
				Get: &openapi3.Operation{
					OperationID: "getCollectionFeatures",
					Parameters: openapi3.Parameters{
						&openapi3.ParameterRef{
							Value: &openapi3.Parameter{
								Name:            "name",
								Description:     "Name of collection to retrieve data for.",
								In:              "path",
								Required:        true,
								Schema:          &openapi3.SchemaRef{Value: openapi3.NewStringSchema()},
								AllowEmptyValue: false,
							},
						},
						&openapi3.ParameterRef{
							Value: &openapi3.Parameter{
								Name:        "limit",
								Description: "Maximum number of results to return.",
								In:          "query",
								Required:    false,
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type:    "integer",
										Min:     func(i int) *float64 { f64 := float64(i); return &f64 }(1),
										Max:     func(i int) *float64 { f64 := float64(i); return &f64 }(config.MaxFeatureLimit),
										Default: config.DefaultFeatureLimit,
									},
								},
								AllowEmptyValue: true,
							},
						},
					},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Value: &openapi3.Response{
								Content: openapi3.Content{
									"application/json": &openapi3.ContentType{
										Schema: &openapi3.SchemaRef{
											Ref: "http://geojson.org/schema/FeatureCollection.json",
										},
									},
								},
							},
						},
					},
				},
			},
			"/collections/{name}/items/{feature_id}": &openapi3.PathItem{
				Summary:     "Single feature data from collection",
				Description: "Provides access to a single feature identitfied by {feature_id} from the named collection",
				Get: &openapi3.Operation{
					OperationID: "getCollectionFeature",
					Parameters: openapi3.Parameters{
						&openapi3.ParameterRef{
							Value: &openapi3.Parameter{
								Name:            "name",
								Description:     "Name of collection to retrieve data for.",
								In:              "path",
								Required:        true,
								Schema:          &openapi3.SchemaRef{Value: openapi3.NewStringSchema()},
								AllowEmptyValue: false,
							},
						},
						&openapi3.ParameterRef{
							Value: &openapi3.Parameter{
								Name:            "feature_id",
								Description:     "Id of feature in collection to retrieve data for.",
								In:              "path",
								Required:        true,
								Schema:          &openapi3.SchemaRef{Value: openapi3.NewStringSchema()},
								AllowEmptyValue: false,
							},
						},
					},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Value: &openapi3.Response{
								Content: openapi3.Content{
									"application/json": &openapi3.ContentType{
										Schema: &openapi3.SchemaRef{
											Ref: "http://geojson.org/schema/Feature.json",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	schemaJSON, err := json.Marshal(openAPI3Schema)
	if err != nil {
		log.Printf("Problem marshalling openapi3 schema: %v", err)
	}

	openAPI3SchemaJSON = schemaJSON

	hasher := fnv.New64()
	hasher.Write(openAPI3SchemaJSON)
	openAPI3SchemaContentId = fmt.Sprintf("%x", hasher.Sum64())
}
