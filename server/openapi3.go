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

package server

import (
	"encoding/json"

	"github.com/jban332/kin-openapi/openapi3"
)

var OpenAPI3Schema openapi3.Swagger
var OpenAPI3SchemaJSON []byte

func init() {
	OpenAPI3Schema = openapi3.Swagger{
		OpenAPI: "3.0.0",
		Info: openapi3.Info{
			Title:       "Go-WFS3 API",
			Description: "WFS-3 compatible API",
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
								Content: openapi3.NewContentWithJSONSchema(&rootContentSchema),
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
							// TODO: There isn't an official json schema for openaip3 yet.  This is the best
							//	I could find as of 2018-03-19
							Ref: "https://github.com/googleapis/gnostic/blob/openapi-v3.0.0-rc2/OpenAPIv3/openapi-3.0.json",
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
								Content: openapi3.NewContentWithJSONSchema(&conformanceClassesSchema),
							},
						},
					},
				},
			},
			"/collections": &openapi3.PathItem{
				Summary:     "Feature Collection MetaData",
				Description: "Provides details about the feature collections available from this server",
				Get: &openapi3.Operation{
					OperationID: "getCollections",
					Parameters:  openapi3.Parameters{},
					Responses: openapi3.Responses{
						"200": &openapi3.ResponseRef{
							Value: &openapi3.Response{
								Content: openapi3.NewContentWithJSONSchema(&collectionsInfoSchema),
							},
						},
					},
				},
			},
		},
	}

	schemaJSON, err := json.Marshal(OpenAPI3Schema)
	if err != nil {
		// TODO: log error
		schemaJSON = []byte("{}")
	}

	OpenAPI3SchemaJSON = schemaJSON
}
