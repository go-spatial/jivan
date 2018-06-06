This package provides go structs conforming to api schemas defined by the wfs3 spec.
This package provides go structs defining the schemas of these structs (via kin-openapi/openapi3)
  for easy use in validation & generating an openapi3 document.
This package provides functions used by the handlers package to collect content in the form of
  read-to-be-marshalled go structs.  Currently only marshalling to JSON is supported.

  collection_meta_data.go: generates content for metadata requests
  conformance.go: generates content for conformance requests
  FeatureCollectionJSONSchema: provides a string variable populated with the schema for a geojson FeatureCollection
  features.go: generates content for feature data requests
  FeatureSchema.go: provides a string variable populated with the schema for a geojson Feature
  openapi3.go: encapsulates generation of json OpenAPI3 document for WFS service.
  root.go: generates content for a root path ("/") request
  validation.go: helper functions for validating encoded responses
  wfs3_types.go: go structs to mirror the types & their schemas specified in the wfs3 spec.
