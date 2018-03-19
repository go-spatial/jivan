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

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/link.yaml
//  for link schema
// Returns a new WFS3 Link object.  href & rel are required, others may be empty strings
func NewLink(url, rel, contenttype, hreflang, title string) *link {
	l := link{
		Href: href{
			Href:        url,
			Rel:         rel,
			ContentType: contenttype,
			Hreflang:    hreflang,
			Title:       title,
		},
	}
	return &l
}

type href struct {
	Href        string `json:"href"`
	Rel         string `json:"rel"`
	ContentType string `json:"type"`
	Hreflang    string `json:"hreflang,omitempty"`
	Title       string `json:"title,omitempty"`
}

type link struct {
	Href href `json:"href"`
}

func (l *link) ContentType(contentType string) {
	l.Href.ContentType = contentType
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/collectionInfo.yaml
//  for collectionInfo schema
type collectionInfo struct {
	// TODO
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/content.yaml
//  for collectionsInfo schema.
type collectionsInfo struct {
	Links       []*link
	Collections []*collectionInfo
}

func (csi *collectionsInfo) ContentType(contentType string) {
	for _, l := range csi.Links {
		l.ContentType(contentType)
	}
}

// --- @See https://raw.githubusercontent.com/opengeospatial/WFS_FES/master/core/openapi/schemas/req-classes.yaml
//  for ConformanceClass schema
type conformanceClass struct {
	ConformsTo []string `json:"conformsTo"`
}
