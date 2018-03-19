///////////////////////////////////////////////////////////////////////////////
//
// The MIT License (MIT)
// Copyright (c) 2018 Jivan Amara
// Copyright (c) 2018 Tom Kralidis
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
// OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE
// USE OR OTHER DEALINGS IN THE SOFTWARE.
//
///////////////////////////////////////////////////////////////////////////////

// go-wfs project main.go
package main

import (
	"path"
	"runtime"

	"github.com/go-spatial/go-wfs/data_provider"
	"github.com/go-spatial/go-wfs/server"
	"github.com/go-spatial/tegola/provider/gpkg"
)

func main() {
	// Grab the path to the testing gpkg (we know where it is relative to this file)
	// This will be a last-resort after the following are implemented:
	//	* command-line data source & config file options
	//	* search working directory for gpkg
	_, execPath, _, _ := runtime.Caller(0)
	execDir := path.Dir(execPath)
	testGpkgPath := path.Join(execDir, "test-data", "athens-osm-20170921.gpkg")
	gpkgPath := testGpkgPath

	gpkgConfig, err := gpkg.AutoConfig(gpkgPath)
	if err != nil {
		panic(err)
	}
	gpkgProvider, err := gpkg.NewTileProvider(gpkgConfig)
	if err != nil {
		panic(err)
	}

	bindAddress := "127.0.0.1:9000"
	serveAddress := bindAddress
	p := data_provider.Provider{Tiler: gpkgProvider}

	server.StartServer(bindAddress, serveAddress, p)
}
