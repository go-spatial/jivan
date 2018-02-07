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
	openapiSpec.Paths.Paths["/api/collectionIds"] = p0

	r200 := spec.Response{}
	r200.Description = "List of collection names available (JSON)."
	r200.Schema = spec.ArrayProperty(spec.StringProperty())
	r200.Examples = make(map[string]interface{})
	example := []string{"roads", "buildings", "waterways"}
	r200.Examples["application/json"] = example
	p0.Get.Responses.StatusCodeResponses[200] = r200

	openapiSpec.Paths.Paths["/api/collectionIds"] = p0

	// Path entry 1
	p1 := spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Get: &spec.Operation{
				OperationProps: spec.OperationProps{
					Description: "Provides all feature ids available; or those in collection if collection parameter is provided",
					Parameters: []spec.Parameter{
						spec.Parameter{
							ParamProps: spec.ParamProps{
								Description:     "Name identifying layer to retreive features for.",
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

	openapiSpec.Paths.Paths["/api/featureIds"] = p1

	// Path entry 2 (/api/feature?id=<featureId>)
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

	//  p3 = ...
	//	openapiSpec.Paths.Paths["/api/features"] = p3
}
