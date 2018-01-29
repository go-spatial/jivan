package server

// Provides mappings for URL routes to handler functions.

import (
	"net/http"
)

func setUpRoutes() {
	http.HandleFunc("/api", getOpenapiSpec)
	http.HandleFunc("/deleteme", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Heyya"))
	})
}
