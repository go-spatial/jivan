package server

// --- Implements req/core/conformance-op
func conformance() *conformanceClass {
	c := conformanceClass{
		ConformsTo: []string{
			"http://www.opengis.net/spec/wfs-1/3.0/req/core",
			"http://www.opengis.net/spec/wfs-1/3.0/req/geojson",
			// TODO: "http://www.opengis.net/spec/wfs-1/3.0/req/html",
		},
	}

	return &c
}
