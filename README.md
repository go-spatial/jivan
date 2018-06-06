# go-wfs [![Build Status](https://travis-ci.org/go-spatial/go-wfs.png)](https://travis-ci.org/go-spatial/go-wfs)

go-wfs is a [Go](https://golang.org) server implementation of [OGC WFS 3.0](https://github.com/opengeospatial/WFS_FES).

**REQUIRES GO >= 1.8**

This project provides a straightforward and simple way to publish your geospatial data on the web.
go-wfs currently supports [GeoPackage](http://www.geopackage.org/spec/) and
[PostGIS](https://postgis.net/) backends.  Providers implement a straightforward interface
so others can be added fairly easily.

## Running
* The simplest way to start is to simply put a GeoPackage file in the same directory as the
executable and run it.

* You can provide the connection details for the data backend and the provider
will scan your data collection and publish any tables with geographical data each as a separate
collection.  Support to customize the data published, including collections based on SQL are
a short coding effort away.

* You can also provide a config file.  Configuration support is in a fairly early state.
Take a look at `go-wfs-config.toml` for an example, and keep in mind::
  * Currently the [logging] section is not used
  * From the [metadata] section, only the following are currently used:
    * title
    * description
  * In the [server] section, only the following are currently used:
    * bind_host
    * bind_port
    * url_scheme
    * url_hostport
    * url_basepath
    * default_mimetype
    * paging_maxlimit

GeoPackage Example:
`go-wfs -d /path/to/my.gpkg`

PostGIS Example:
`go-wfs -d 'host=my.dbhost.org port=5432 dbname=mydbname user=myuser password=mypassword'`

Then visit http://127.0.0.1:9000 to view your data as a wfs3 service.

**go-wfs** provides a number of handy flags to customize where it binds and the links it generates
in results to make it simple for sysadmins to, for example, deploy behind a proxy.
Run `go-wfs --help` for details.

See go-wfs-config.toml
**TODO**: TOML configuration files

Currently supported TOML:
* config.Configuration.Server.DefaultMimeType
* config.Configuration.Server.MaxLimit

## Bugs and Issues
All bug reports, enhancement requests, and other issues are managed on
[GitHub](https://github.com/go-spatial/go-wfs).


# Developer Notes
for further details see the README in each folder

* **config/**
  Responsible for dealing with configuration files & providing the parameters from a config
  to the rest of the system.

* **server/**
  Responsible for the actual server and all html traffic specific tasks such as determining
  content encoding and collecting url query parameters.

* **wfs3/**
  Responsible for wfs3-specific details such as go versions of wfs3 types and collecting
  appropriate data for each wfs3 endpoint.  The types here implement the supported encodings.

* **data_provider/**
  Responsible for access to data backends.  Essentially a wrapper with some functionality added
  for [tegola data providers](https://github.com/go-spatial/tegola/tree/filterer_implementation/provider)

* **main.go**
  Executable entry-point.

Visit http://localhost:9000/api for OpenAPI definition of the service.
Take a look at `server/routes.go` for a concise list of supported URLs.

## Build Instructions

```bash
# create directory for local env
mkdir /path/to/golang-env
export GOPATH=/path/to/golang-env
# install go-wfs
go get github.com/go-spatial/go-wfs
# install 'dep' dependency manager
go get github.com/golang/dep/...
# install dependencies in vendor/
dep ensure
go build -i -o go-wfs github.com/go-spatial/go-wfs
```
To build for AWS Lambda deployments, add `-tags awslambda` when building

## Running

```bash
# start server on http://localhost:9000/
go run main.go  # or go build main.go
```

## Dependencies
`dep ensure` will install dependencies at the current HEAD when you run it (equivalent to `go get ...`)

Run `dep ensure -update` periodically to stay current with these dependencies. (equivalent to subsequent `go get ...`)

Please don't add `Gopkg.lock` to the repo.

## Requests Overview

Features are identified by a _collection name_ and _feature id_ pair.

- API landing: http://localhost:9000/
- API definition: http://localhost:9000/api
- Conformance: http://localhost:9000/conformance
- Collections: http://localhost:9000/collections
- Feature collection metadata: http://localhost:9000/collections/{name}
- Features from a single feature collection: http://localhost:9000/collections/{name}/items
- Single feature from a feature collection: http://localhost:9000/collections/{name}/items/{featureid}
