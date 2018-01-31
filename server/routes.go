package server

// Provides mappings for URL routes to handler functions.

import (
	"net/http"
)

func setUpRoutes() {
	http.HandleFunc("/api", getOpenapiSpec)
	http.HandleFunc("/api/layers", getLayers)
	http.HandleFunc("/api/layer", getLayerFeatures)
}
