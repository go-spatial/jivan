# go-wfs

go-wfs is a [Go](https://golang.org) server implementation of [OGC WFS 3.0](https://github.com/opengeospatial/WFS_FES).

**REQUIRES GO >= 1.7**

server/
  routes.go: maps urls to functions (from handlers.go)
  handlers.go: actual work done here
  server.go: simple interface to start the server.
  openapi.go: encapsulates generation of json OpenAPI document for WFS service.

provider/
  provider.go: wraps github.com/go-spatial/tegola/provider.Tiler to provide convenience methods & additional behavior

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
# FIXME: temporary hack to check out gpkg_autoconfig branch
cd $GOPATH/src/github.com/go-spatial/tegola
git checkout gpkg_autoconfig
go build
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

- Collections: http://localhost:9000/api/collectionNames
- Features primary keys: http://localhost:9000/api/featurePks
- Features from a single collection: http://localhost:9000/api/collection
- Single feature from a single collection: http://localhost:9000/api/feature
- Create new, filtered, temporary collection: http://localhost:9000/api/feature_set

## Bugs and Issues

All bugs, enhancements and issues are managed on [GitHub](https://github.com/go-spatial/go-wfs).
