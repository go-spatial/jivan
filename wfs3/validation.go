package wfs3

import (
	"fmt"
	"io"
	"net/http"

	"github.com/jban332/kin-openapi/openapi3"
	"github.com/jban332/kin-openapi/openapi3filter"
	"github.com/xeipuuv/gojsonschema"
)

// Validate a json response provided by a Reader using kin-openapi/openapi3 against the openapi3
//	scaffolding set up in wfs3/openapi3.go
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

// Validate a Reader providing the response body against a string json schema
func ValidateJSONResponseAgainstJSONSchema(jsonResponse []byte, jsonSchema string) error {
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)
	respLoader := gojsonschema.NewStringLoader(string(jsonResponse))

	result, err := gojsonschema.Validate(schemaLoader, respLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		err = fmt.Errorf("document is invalid")
		return err
	}

	return nil
}
