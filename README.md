server/
  routes.go: maps urls to functions (from handlers.go)
  handlers.go: actual work done here
  server.go: simple interface to start the server.
  openapi.go: encapsulates generation of json OpenAPI document for WFS service.

main.go: Executable entry-point.

Defaults to run on localhost:9000.  Visit http://localhost:9000/api for OpenAPI definition of
service.  Take a look at server/routes.go for a concise list of supported URLs.

Build Instructions
------------------

These are temporary while waiting on some tegola PRs.  Things will be simpler before long.

Essentially go-wfs needs the tegola branch `gpkg_autoconfig` plus the `geom/`
package from the tegola branch `issue-288_geojson_encoding`.

Here is how you can make that happen:

1. clone github.com/terranodo/tegola
1. cd tegola, check out branch `gpkg_autoconfig`
1. cp -r geom geom-bak
1. check out branch `gpkg_autoconfig`
1. mv geom-bak geom
1. cd ..
1. clone github.com/go-spatial/go-wfs
1. cd go-wfs
1. mkdir -p src/github.com/teranodo/
1. ln -s /<path>/<to>/tegola src/github.com/terranodo/tegola

Then run / build like you normally would for go:

1. Make sure your GOPATH & GOROOT are set to /<path>/<to>/go-wfs & /<path>/<to>/<golang-installation>
1. `go run main.go` or `go build main.go`
