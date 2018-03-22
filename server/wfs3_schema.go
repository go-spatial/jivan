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

package server

import "github.com/jban332/kin-openapi/openapi3"

// --- @See http://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/root.yaml
//	for rootContentSchema Definition
// What the endpoint at "/" returns
type rootContent struct {
	Links []*link `json:"links"`
}

func (rc rootContent) ContentType(contentType string) {
	for _, l := range rc.Links {
		l.ContentType(contentType)
	}
}

var rootContentSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"links"},
	Properties: map[string]*openapi3.SchemaRef{
		"links": &openapi3.SchemaRef{
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

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/bbox.yaml
//	for bbox schema
// maxItems is needed for setting the bbox array MaxItems in the below Schema literal.
var maxItems int64 = 4

type bbox struct {
	Crs  string    `json:"crs"`
	Bbox []float64 `json:"bbox"`
}

var bboxSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"bbox"},
	Properties: map[string]*openapi3.SchemaRef{
		"crs": &openapi3.SchemaRef{
			// TODO: This is supposed to have an enum & default based on: http://www.opengis.net/def/crs/OGC/1.3/CRS84
			Value: openapi3.NewStringSchema(),
		},
		"bbox": &openapi3.SchemaRef{
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
type link struct {
	Href     string `json:"href"`
	Rel      string `json:"rel"`
	Type     string `json:"type"`
	Hreflang string `json:"hreflang"`
	Title    string `json:"title"`
}

var linkSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"href"},
	Properties: map[string]*openapi3.SchemaRef{
		"href": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"rel": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"type": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"hreflang": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"title": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
	},
}

func (l *link) ContentType(contentType string) {
	l.Type = contentType
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/collectionInfo.yaml
//  for collectionInfo schema
type collectionInfo struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Links       []*link  `json:"links"`
	Extent      *bbox    `json:"extent"`
	Crs         []string `json:"crs"`
}

func (ci *collectionInfo) ContentType(contentType string) {
	for _, l := range ci.Links {
		l.ContentType(contentType)
	}
}

var collectionInfoSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"name", "links"},
	Properties: map[string]*openapi3.SchemaRef{
		"name": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"title": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"description": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "string",
			},
		},
		"links": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &linkSchema,
				},
			},
		},
		"extent": &openapi3.SchemaRef{
			Value: &bboxSchema,
		},
		"crs": &openapi3.SchemaRef{
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
type collectionsInfo struct {
	Links       []*link           `json:"links"`
	Collections []*collectionInfo `json:"collections"`
}

func (csi *collectionsInfo) ContentType(contentType string) {
	for _, l := range csi.Links {
		l.ContentType(contentType)
	}
	for _, c := range csi.Collections {
		c.ContentType(contentType)
	}
}

var collectionsInfoSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"links", "collections"},
	Properties: map[string]*openapi3.SchemaRef{
		"links": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &linkSchema,
				},
			},
		},
		"collections": &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: "array",
				Items: &openapi3.SchemaRef{
					Value: &collectionInfoSchema,
				},
			},
		},
	},
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/req-classes.yaml
//  for ConformanceClasses schema
type conformanceClasses struct {
	ConformsTo []string `json:"conformsTo"`
}

var conformanceClassesSchema openapi3.Schema = openapi3.Schema{
	Type:     "object",
	Required: []string{"conformsTo"},
	Properties: map[string]*openapi3.SchemaRef{
		"conformsTo": &openapi3.SchemaRef{
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
