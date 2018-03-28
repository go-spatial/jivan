package wfs3

import (
	"fmt"
	"io"
	"net/http"

	"github.com/jban332/kin-openapi/openapi3"
	"github.com/jban332/kin-openapi/openapi3filter"
)

func ValidateJSONResponse(request *http.Request, path string, status int, header http.Header, respBodyRC io.ReadCloser) error {
	var op *openapi3.Operation
	switch request.Method {
	case "GET":
		if OpenAPI3Schema.Paths[path] == nil {
			return fmt.Errorf("Path not found in schema: '%v'", path)
		}
		op = OpenAPI3Schema.Paths[path].Get
	default:
		return fmt.Errorf("unsupported request.Method: %v", request.Method)
	}

	rvi := openapi3filter.RequestValidationInput{
		Request: request,
		Route: &openapi3filter.Route{
			Swagger:   &OpenAPI3Schema,
			Server:    &openapi3.Server{},
			Path:      path,
			PathItem:  &openapi3.PathItem{},
			Method:    request.Method,
			Operation: op,
		},
	}

	err := openapi3filter.ValidateResponse(nil, &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &rvi,
		Status:                 status,
		Header:                 header,
		Body:                   respBodyRC,
	})
	return err
}
