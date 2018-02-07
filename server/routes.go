package server

// Provides mappings for URL routes to handler functions.

import (
	"net/http"
)

func setUpRoutes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/api", 307) })

	http.HandleFunc("/api", getOpenapiSpec)
	http.HandleFunc("/api/collectionIds", getCollectionIds)
	http.HandleFunc("/api/featureIds", getFeatureIds)
	http.HandleFunc("/api/feature", getFeature)
}
