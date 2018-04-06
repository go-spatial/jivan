///////////////////////////////////////////////////////////////////////////////
//
// The MIT License (MIT)
// Copyright (c) 2018 Jivan Amara
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

// go-wfs project wfs3_schema.go

package wfs3

import "github.com/jban332/kin-openapi/openapi3"

// --- @See http://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/root.yaml
//	for rootContentSchema Definition
// What the endpoint at "/" returns
type RootContent struct {
	Links []*Link `json:"links"`
}

func (rc RootContent) ContentType(contentType string) RootContent {
	for _, l := range rc.Links {
		l.ContentType(contentType)
	}
	return rc
}

var RootContentSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"links"},
	Properties: map[string]*openapi3.SchemaRef{
		"links": {
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &LinkSchema,
				},
			},
		},
	},
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/bbox.yaml
//	for bbox schema
// maxItems is needed for setting the bbox array MaxItems in the below Schema literal.
var maxItems int64 = 4

type Bbox struct {
	Crs  string    `json:"crs"`
	Bbox []float64 `json:"bbox"`
}

var BboxSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"bbox"},
	Properties: map[string]*openapi3.SchemaRef{
		"crs": {
			// TODO: This is supposed to have an enum & default based on: http://www.opengis.net/def/crs/OGC/1.3/CRS84
			Value: openapi3.NewStringSchema(),
		},
		"bbox": {
			Value: &openapi3.Schema{
				Type:     "array",
				MinItems: 4,
				MaxItems: &maxItems,
				Items:    openapi3.NewSchemaRef("", openapi3.NewFloat64Schema().WithMin(-180).WithMax(180)),
			},
		},
	},
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/link.yaml
//  for link schema
type Link struct {
	Href     string `json:"href"`
	Rel      string `json:"rel"`
	Type     string `json:"type"`
	Hreflang string `json:"hreflang"`
	Title    string `json:"title"`
}

var LinkSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"href"},
	Properties: map[string]*openapi3.SchemaRef{
		"href": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"rel": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"type": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"hreflang": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"title": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
	},
}

func (l *Link) ContentType(contentType string) {
	l.Type = contentType
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/collectionInfo.yaml
//  for collectionInfo schema
type CollectionInfo struct {
	Name        string   `json:"name"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Links       []*Link  `json:"links"`
	Extent      *Bbox    `json:"extent,omitempty"`
	Crs         []string `json:"crs,omitempty"`
}

func (ci *CollectionInfo) ContentType(contentType string) {
	for _, l := range ci.Links {
		l.ContentType(contentType)
	}
}

var CollectionInfoSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"name", "links"},
	Properties: map[string]*openapi3.SchemaRef{
		"name": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"title": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"description": {
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"links": {
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &LinkSchema,
				},
			},
		},
		"extent": {
			Value: &BboxSchema,
		},
		"crs": {
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: "string",
					},
				},
			},
		},
	},
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/content.yaml
//  for collectionsInfo schema.
type CollectionsInfo struct {
	Links       []*Link           `json:"links"`
	Collections []*CollectionInfo `json:"collections"`
}

func (csi *CollectionsInfo) ContentType(contentType string) {
	for _, l := range csi.Links {
		l.ContentType(contentType)
	}
	for _, c := range csi.Collections {
		c.ContentType(contentType)
	}
}

var CollectionsInfoSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"links", "collections"},
	Properties: map[string]*openapi3.SchemaRef{
		"links": {
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &LinkSchema,
				},
			},
		},
		"collections": {
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &CollectionInfoSchema,
				},
			},
		},
	},
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/req-classes.yaml
//  for ConformanceClasses schema
type ConformanceClasses struct {
	ConformsTo []string `json:"conformsTo"`
}

var ConformanceClassesSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"conformsTo"},
	Properties: map[string]*openapi3.SchemaRef{
		"conformsTo": {
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: "string",
					},
				},
			},
		},
	},
}

var GeoJsonFeatureSchema openapi3.Schema = openapi3.Schema{
	Properties: map[string]*openapi3.SchemaRef{
		"bbox": {
			Value: openapi3.NewArraySchema().WithItems(openapi3.NewFloat64Schema()).WithMinItems(4),
		},
		"geometry": {
			Value: openapi3.NewOneOfSchema(
				&openapi3.Schema{Type: "null"},
				&openapi3.Schema{
					Properties: map[string]*openapi3.SchemaRef{
						"bbox": {
							Value: openapi3.NewArraySchema().WithItems(openapi3.NewFloat64Schema()).WithMinItems(4),
						},
						"coordinates": {
							Value: openapi3.NewArraySchema().WithItems(openapi3.NewFloat64Schema()).WithMinItems(2),
						},
						"type": {
							Value: &openapi3.Schema{
								Enum: []interface{}{"Point"},
								Type: "string",
							},
						},
					},
					Required: []string{"type", "coordinates"},
					// Title:    "GeoJSON Point",
					Type: "object",
				}),
		},
	},
}

// "$id": "http://geojson.org/schema/Feature.json",
// "$schema": "http://json-schema.org/draft-07/schema#",
// 	"properties": {
// 		"geometry": {
// 			"oneOf": [
// 				{
// 					"type": "null"
// 				},
// 				{
// 					"properties": {
// 						"bbox": {
// 							"items": {
// 								"type": "number"
// 							},
// 							"minItems": 4,
// 							"type": "array"
// 						},
// 						"coordinates": {
// 							"items": {
// 								"type": "number"
// 							},
// 							"minItems": 2,
// 							"type": "array"
// 						},
// 						"type": {
// 							"enum": [
// 								"Point"
// 							],
// 							"type": "string"
// 						}
// 					},
// 					"required": [
// 						"type",
// 						"coordinates"
// 					],
// 					"title": "GeoJSON Point",
// 					"type": "object"
// 				},
// 				{
// 					"properties": {
// 						"bbox": {
// 							"items": {
// 								"type": "number"
// 							},
// 							"minItems": 4,
// 							"type": "array"
// 						},
// 						"coordinates": {
// 							"items": {
// 								"items": {
// 									"type": "number"
// 								},
// 								"minItems": 2,
// 								"type": "array"
// 							},
// 							"minItems": 2,
// 							"type": "array"
// 						},
// 						"type": {
// 							"enum": [
// 								"LineString"
// 							],
// 							"type": "string"
// 						}
// 					},
// 					"required": [
// 						"type",
// 						"coordinates"
// 					],
// 					"title": "GeoJSON LineString",
// 					"type": "object"
// 				},
// 				{
// 					"properties": {
// 						"bbox": {
// 							"items": {
// 								"type": "number"
// 							},
// 							"minItems": 4,
// 							"type": "array"
// 						},
// 						"coordinates": {
// 							"items": {
// 								"items": {
// 									"items": {
// 										"type": "number"
// 									},
// 									"minItems": 2,
// 									"type": "array"
// 								},
// 								"minItems": 4,
// 								"type": "array"
// 							},
// 							"type": "array"
// 						},
// 						"type": {
// 							"enum": [
// 								"Polygon"
// 							],
// 							"type": "string"
// 						}
// 					},
// 					"required": [
// 						"type",
// 						"coordinates"
// 					],
// 					"title": "GeoJSON Polygon",
// 					"type": "object"
// 				},
// 				{
// 					"properties": {
// 						"bbox": {
// 							"items": {
// 								"type": "number"
// 							},
// 							"minItems": 4,
// 							"type": "array"
// 						},
// 						"coordinates": {
// 							"items": {
// 								"items": {
// 									"type": "number"
// 								},
// 								"minItems": 2,
// 								"type": "array"
// 							},
// 							"type": "array"
// 						},
// 						"type": {
// 							"enum": [
// 								"MultiPoint"
// 							],
// 							"type": "string"
// 						}
// 					},
// 					"required": [
// 						"type",
// 						"coordinates"
// 					],
// 					"title": "GeoJSON MultiPoint",
// 					"type": "object"
// 				},
// 				{
// 					"properties": {
// 						"bbox": {
// 							"items": {
// 								"type": "number"
// 							},
// 							"minItems": 4,
// 							"type": "array"
// 						},
// 						"coordinates": {
// 							"items": {
// 								"items": {
// 									"items": {
// 										"type": "number"
// 									},
// 									"minItems": 2,
// 									"type": "array"
// 								},
// 								"minItems": 2,
// 								"type": "array"
// 							},
// 							"type": "array"
// 						},
// 						"type": {
// 							"enum": [
// 								"MultiLineString"
// 							],
// 							"type": "string"
// 						}
// 					},
// 					"required": [
// 						"type",
// 						"coordinates"
// 					],
// 					"title": "GeoJSON MultiLineString",
// 					"type": "object"
// 				},
// 				{
// 					"properties": {
// 						"bbox": {
// 							"items": {
// 								"type": "number"
// 							},
// 							"minItems": 4,
// 							"type": "array"
// 						},
// 						"coordinates": {
// 							"items": {
// 								"items": {
// 									"items": {
// 										"items": {
// 											"type": "number"
// 										},
// 										"minItems": 2,
// 										"type": "array"
// 									},
// 									"minItems": 4,
// 									"type": "array"
// 								},
// 								"type": "array"
// 							},
// 							"type": "array"
// 						},
// 						"type": {
// 							"enum": [
// 								"MultiPolygon"
// 							],
// 							"type": "string"
// 						}
// 					},
// 					"required": [
// 						"type",
// 						"coordinates"
// 					],
// 					"title": "GeoJSON MultiPolygon",
// 					"type": "object"
// 				},
// 				{
// 					"properties": {
// 						"bbox": {
// 							"items": {
// 								"type": "number"
// 							},
// 							"minItems": 4,
// 							"type": "array"
// 						},
// 						"geometries": {
// 							"items": {
// 								"oneOf": [
// 									{
// 										"properties": {
// 											"bbox": {
// 												"items": {
// 													"type": "number"
// 												},
// 												"minItems": 4,
// 												"type": "array"
// 											},
// 											"coordinates": {
// 												"items": {
// 													"type": "number"
// 												},
// 												"minItems": 2,
// 												"type": "array"
// 											},
// 											"type": {
// 												"enum": [
// 													"Point"
// 												],
// 												"type": "string"
// 											}
// 										},
// 										"required": [
// 											"type",
// 											"coordinates"
// 										],
// 										"title": "GeoJSON Point",
// 										"type": "object"
// 									},
// 									{
// 										"properties": {
// 											"bbox": {
// 												"items": {
// 													"type": "number"
// 												},
// 												"minItems": 4,
// 												"type": "array"
// 											},
// 											"coordinates": {
// 												"items": {
// 													"items": {
// 														"type": "number"
// 													},
// 													"minItems": 2,
// 													"type": "array"
// 												},
// 												"minItems": 2,
// 												"type": "array"
// 											},
// 											"type": {
// 												"enum": [
// 													"LineString"
// 												],
// 												"type": "string"
// 											}
// 										},
// 										"required": [
// 											"type",
// 											"coordinates"
// 										],
// 										"title": "GeoJSON LineString",
// 										"type": "object"
// 									},
// 									{
// 										"properties": {
// 											"bbox": {
// 												"items": {
// 													"type": "number"
// 												},
// 												"minItems": 4,
// 												"type": "array"
// 											},
// 											"coordinates": {
// 												"items": {
// 													"items": {
// 														"items": {
// 															"type": "number"
// 														},
// 														"minItems": 2,
// 														"type": "array"
// 													},
// 													"minItems": 4,
// 													"type": "array"
// 												},
// 												"type": "array"
// 											},
// 											"type": {
// 												"enum": [
// 													"Polygon"
// 												],
// 												"type": "string"
// 											}
// 										},
// 										"required": [
// 											"type",
// 											"coordinates"
// 										],
// 										"title": "GeoJSON Polygon",
// 										"type": "object"
// 									},
// 									{
// 										"properties": {
// 											"bbox": {
// 												"items": {
// 													"type": "number"
// 												},
// 												"minItems": 4,
// 												"type": "array"
// 											},
// 											"coordinates": {
// 												"items": {
// 													"items": {
// 														"type": "number"
// 													},
// 													"minItems": 2,
// 													"type": "array"
// 												},
// 												"type": "array"
// 											},
// 											"type": {
// 												"enum": [
// 													"MultiPoint"
// 												],
// 												"type": "string"
// 											}
// 										},
// 										"required": [
// 											"type",
// 											"coordinates"
// 										],
// 										"title": "GeoJSON MultiPoint",
// 										"type": "object"
// 									},
// 									{
// 										"properties": {
// 											"bbox": {
// 												"items": {
// 													"type": "number"
// 												},
// 												"minItems": 4,
// 												"type": "array"
// 											},
// 											"coordinates": {
// 												"items": {
// 													"items": {
// 														"items": {
// 															"type": "number"
// 														},
// 														"minItems": 2,
// 														"type": "array"
// 													},
// 													"minItems": 2,
// 													"type": "array"
// 												},
// 												"type": "array"
// 											},
// 											"type": {
// 												"enum": [
// 													"MultiLineString"
// 												],
// 												"type": "string"
// 											}
// 										},
// 										"required": [
// 											"type",
// 											"coordinates"
// 										],
// 										"title": "GeoJSON MultiLineString",
// 										"type": "object"
// 									},
// 									{
// 										"properties": {
// 											"bbox": {
// 												"items": {
// 													"type": "number"
// 												},
// 												"minItems": 4,
// 												"type": "array"
// 											},
// 											"coordinates": {
// 												"items": {
// 													"items": {
// 														"items": {
// 															"items": {
// 																"type": "number"
// 															},
// 															"minItems": 2,
// 															"type": "array"
// 														},
// 														"minItems": 4,
// 														"type": "array"
// 													},
// 													"type": "array"
// 												},
// 												"type": "array"
// 											},
// 											"type": {
// 												"enum": [
// 													"MultiPolygon"
// 												],
// 												"type": "string"
// 											}
// 										},
// 										"required": [
// 											"type",
// 											"coordinates"
// 										],
// 										"title": "GeoJSON MultiPolygon",
// 										"type": "object"
// 									}
// 								]
// 							},
// 							"type": "array"
// 						},
// 						"type": {
// 							"emum": [
// 								"GeometryCollection"
// 							],
// 							"type": "string"
// 						}
// 					},
// 					"required": [
// 						"type",
// 						"geometries"
// 					],
// 					"title": "GeoJSON GeometryCollection",
// 					"type": "object"
// 				}
// 			]
// 		},
// 		"properties": {
// 			"type": "object"
// 		},
// 		"type": {
// 			"emum": [
// 				"Feature"
// 			],
// 			"type": "string"
// 		}
// 	},
// 	"required": [
// 		"type",
// 		"properties",
// 		"geometry"
// 	],
// 	"title": "GeoJSON Feature",
// 	"type": "object"
// }
