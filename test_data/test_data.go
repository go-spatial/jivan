package test_data

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/go-spatial/tegola/provider"
	"github.com/go-spatial/tegola/provider/gpkg"
)

func GetTestGPKGTiler() provider.Tiler {
	_, execPath, _, _ := runtime.Caller(0)
	// This is the directory that this file is located in (at compile-time)
	thisFileDir := path.Dir(execPath)
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
		// TODO: log error
	}

	// places to look for the test gpkg in decreasing priority
	pathOptions := []string{
		path.Join(thisFileDir, "athens-osm-20170921.gpkg"),
		path.Join(wd, "test_data", "athens-osm-20170921.gpkg"),
		// This is for tests running in sub-packages
		path.Join(wd, "..", "test_data", "athens-osm-20170921.gpkg"),
	}
	gpkgPath := ""

	for _, p := range pathOptions {
		if _, err := os.Stat(p); err == nil {
			gpkgPath = p
			break
		}
	}

	if gpkgPath == "" {
		panic(fmt.Sprintf("unable to find test gpkg in %v", pathOptions))
	}

	gpkgConfig, err := gpkg.AutoConfig(gpkgPath)
	if err != nil {
		panic(fmt.Sprintf("gpkg auto-config failure for '%v': %v", gpkgPath, err))
	}
	gpkgProvider, err := gpkg.NewTileProvider(gpkgConfig)
	if err != nil {
		panic(fmt.Sprintf("gpkg tile provider creation error for '%v': %v", gpkgPath, err))
	}
	return gpkgProvider
}
