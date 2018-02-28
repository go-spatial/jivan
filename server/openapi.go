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

	"github.com/go-openapi/spec"
)

var openapiSpec spec.Swagger
var openapiSpecJson []byte

func OpenApiSpecJson() (result []byte, err error) {
	if openapiSpecJson == nil {
		openapiSpecJson, err = json.Marshal(openapiSpec)
	}
	if err != nil {
		return nil, err
	}

	return openapiSpecJson, nil
}

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
		Paths: map[string]spec.PathItem{},
	}

	// Path entry 0
	p0 := spec.PathItem{}
	p0.Get = new(spec.Operation)
	p0.Get.Description = "Provide names of collections available in API.  Features are grouped by collection."
	p0.Get.Responses = new(spec.Responses)
	p0.Get.Responses.StatusCodeResponses = make(map[int]spec.Response)
	p0.Get.Responses.StatusCodeResponses[200] = spec.Response{}
	openapiSpec.Paths.Paths["/api/collectionNames"] = p0

	r200 := spec.Response{}
	r200.Description = "List of collection names available (JSON)."
	r200.Schema = spec.ArrayProperty(spec.StringProperty())
	r200.Examples = make(map[string]interface{})
	example := []string{"roads", "buildings", "waterways"}
	r200.Examples["application/json"] = example
	p0.Get.Responses.StatusCodeResponses[200] = r200

	openapiSpec.Paths.Paths["/api/collectionNames"] = p0

	// Path entry 1
	p1 := spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Get: &spec.Operation{
				OperationProps: spec.OperationProps{
					Description: "Provides all feature pks available; or those in collection if collection parameter is provided",
					Parameters: []spec.Parameter{
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Description:     "Name identifying collection to retreive features for.",
								Name:            "collection",
								In:              "Query",
								Required:        false,
								Schema:          spec.StringProperty(),
								AllowEmptyValue: true,
							},
						},
					},
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{
								200: spec.Response{
									ResponseProps: spec.ResponseProps{
										Description: "List of feature ids available",
										Schema:      spec.ArrayProperty(spec.Int64Property()),
										Examples: map[string]interface{}{
											"application/json": []uint64{12, 13, 203, 207},
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

	openapiSpec.Paths.Paths["/api/featurePks"] = p1

	// Path entry 2 (/api/feature?pk=<featurePk>)
	p2 := spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Get: &spec.Operation{
				OperationProps: spec.OperationProps{
					Description: "Provides feature data for feature requested.",
					Parameters: []spec.Parameter{
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Description:     "Id of feature to retreive.",
								Name:            "id",
								In:              "Query",
								Required:        true,
								Schema:          spec.StringProperty(),
								AllowEmptyValue: false,
							},
						},
					},
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{
								200: spec.Response{
									ResponseProps: spec.ResponseProps{
										Description: "Feature data for feature requested",
										Schema:      spec.DateProperty(),
										Examples: map[string]interface{}{
											"application/json": `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[23.6946291,37.942376],[23.6946775,37.9421025],[23.6942521,37.9420922]]]},"properties":{}}`,
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

	openapiSpec.Paths.Paths["/api/feature"] = p2

	// Path entry 3 (/api/collection?id=<collectionId>&page=<pageNumber>&pageSize=<pageSize>)
	// TODO: param explode setting for array values
	p3 := spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Get: &spec.Operation{
				OperationProps: spec.OperationProps{
					Description: "Provides paged feature data for features in collection(s) requested.",
					Parameters: []spec.Parameter{
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Name:            "collection",
								Description:     "Limit to features in collection(s) identified by id(s).",
								In:              "Query",
								Required:        true,
								Schema:          spec.ArrayProperty(spec.Int64Property()),
								AllowEmptyValue: false,
							},
						},
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Name:            "page",
								Description:     "Show data only for page indicated (default 0).",
								In:              "Query",
								Required:        false,
								Schema:          spec.Int64Property(),
								AllowEmptyValue: false,
							},
						},
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Name:            "pageSize",
								Description:     "Include this many features per page (default 10)",
								In:              "Query",
								Required:        false,
								Schema:          spec.Int64Property(),
								AllowEmptyValue: false,
							},
						},
					},
					// TODO: Response schema & example
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{
								200: spec.Response{
									ResponseProps: spec.ResponseProps{
										Description: "Feature data for feature requested",
										Schema:      spec.DateProperty(),
										Examples: map[string]interface{}{
											"application/json": `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[23.6946291,37.942376],[23.6946775,37.9421025],[23.6942521,37.9420922]]]},"properties":{}}`,
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

	openapiSpec.Paths.Paths["/api/collection"] = p3

	// Path entry 4 (/api/feature_set?extent=<geomBoundingArea>&<property>=<propertyValue>&collection=<collectionId>)
	// TODO: Should we just make 'extent' a property?
	p4 := spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Get: &spec.Operation{
				OperationProps: spec.OperationProps{
					Description: "Provides the name of a collection data for features matching filters. Consider parameters as filters combined with a logical 'and'.",
					Parameters: []spec.Parameter{
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Name:            "extent",
								Description:     "Include only features partially or fully within this geometry (Not yet implemented).",
								In:              "Query",
								Required:        false,
								Schema:          spec.StringProperty(),
								AllowEmptyValue: false,
							},
						},
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Name:            "<property>",
								Description:     "Include only features that have this property with this value, many different properties are allowed.",
								In:              "Query",
								Required:        false,
								Schema:          spec.StringProperty(),
								AllowEmptyValue: false,
							},
						},
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Name:            "collection",
								Description:     "Limit to features in collection(s) identified by name(s). If no collection is specified, all collections will be included",
								In:              "Query",
								Required:        false,
								Schema:          spec.ArrayProperty(spec.StringProperty()),
								AllowEmptyValue: false,
							},
						},
					},
					// TODO: Response schema & example
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{
								200: spec.Response{
									ResponseProps: spec.ResponseProps{
										Description: "Feature data for feature requested",
										Schema:      spec.DateProperty(),
										Examples: map[string]interface{}{
											"application/json": `{"type":"Feature","geometry":{"type":"Polygon","coordinates":[[[23.6946291,37.942376],[23.6946775,37.9421025],[23.6942521,37.9420922]]]},"properties":{}}`,
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

	openapiSpec.Paths.Paths["/api/feature_set"] = p4
}
