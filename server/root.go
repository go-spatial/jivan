package server

import "fmt"

// What the endpoint at "/" returns
type rootContent struct {
	Links []*link `json:"links"`
}

func (rc rootContent) ContentType(contentType string) {
	for _, l := range rc.Links {
		l.ContentType(contentType)
	}
}

func root() rootContent {
	fmt.Printf("serveAddress in root.go/root(): %v\n", serveAddress)
	apiUrl := fmt.Sprintf("http://%v/api", serveAddress)
	conformanceUrl := fmt.Sprintf("http://%v/conformance", serveAddress)
	collectionsUrl := fmt.Sprintf("http://%v/collections", serveAddress)

	r := rootContent{
		Links: []*link{
			NewLink(apiUrl, "service", "", "", ""),
			NewLink(conformanceUrl, "conformance", "", "", ""),
			NewLink(collectionsUrl, "data", "", "", ""),
		},
	}

	return r
}
