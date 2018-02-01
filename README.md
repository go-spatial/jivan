server/
  routes.go: maps urls to functions (from handlers.go)
  handlers.go: actual work done here
  server.go: simple interface to start the server.

provider/
  provider.go: Interface definition for data providers

gpkg/
  gpkg.go: GeoPackage data provider (Making heavy use of tegola GeoPackage provider utilities)

Defaults to run on localhost:9000.  Visit http://localhost:9000/api for OpenAPI definition of
service.  Take a look at server/routes.go for a concise list of supported URLs.
