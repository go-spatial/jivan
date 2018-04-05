# go-wfs

go-wfs is a [Go](https://golang.org) server implementation of [OGC WFS 3.0](https://github.com/opengeospatial/WFS_FES).

**REQUIRES GO >= 1.7**

server/
  routes.go: maps urls to functions (from handlers.go)
  handlers.go: actual work done here
  server.go: simple interface to start the server.

wfs3/
  collection_meta_data.go: generates content for metadata requests
  conformance.go: generates content for conformance requests
  FeatureCollectionJSONSchema: provides a string variable populated with the schema for a geojson FeatureCollection
  features.go: generates content for feature data requests
  FeatureSchema.go: provides a string variable populated with the schema for a geojson Feature
  openapi3.go: encapsulates generation of json OpenAPI3 document for WFS service.
  root.go: generates content for a root path ("/") request
  validation.go: helper functions for validating encoded responses
  wfs3_types.go: go structs to mirror the types & their schemas specified in the wfs3 spec.

data_provider/
  provider.go: wraps `github.com/go-spatial/tegola/provider.Tiler` to provide convenience methods & additional behavior

main.go: Executable entry-point.

Defaults to run on localhost:9000.  Visit http://localhost:9000/api for OpenAPI definition of
service.  Take a look at `server/routes.go` for a concise list of supported URLs.

## Build Instructions

```bash
# create directory for local env
mkdir /path/to/golang-env
export GOPATH=/path/to/golang-env
# install tegola
go get github.com/go-spatial/tegola
cd $GOPATH/src/github.com/go-spatial/tegola
git checkout v0.7.0
# install other supporting packages
go get github.com/jban332/kin-openapi/openapi3
go get github.com/julienschmidt/httprouter
go get github.com/rs/cors

# install go-wfs
go get github.com/go-spatial/go-wfs
```

## Running

```bash
# start server on http://localhost:9000/
go run main.go  # or go build main.go
```


## Requests Overview

Features are identified by a _collection name_ and _feature primary key_ pair.

- API landing: http://localhost:9000/
- API definition: http://localhost:9000/api
- Conformance: http://localhost:9000/conformance
- Collections: http://localhost:9000/collections
- Feature collection metadata: http://localhost:9000/collections/{name}
- Features from a single feature collection: http://localhost:9000/collections/{name}/items
- Single feature from a feature collection: http://localhost:9000/collections/{name}/items/{feature}

## Bugs and Issues

All bugs, enhancements and issues are managed on [GitHub](https://github.com/go-spatial/go-wfs).
