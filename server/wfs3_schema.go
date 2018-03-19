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
