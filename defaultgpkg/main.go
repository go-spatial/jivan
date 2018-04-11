package defaultgpkg

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
)

// For use when a data source isn't provided.
// Scans for files w/ '.gpkg' extension and chooses the first alphabetically.
// * First scan the working directory
// * Second, if none found in the working directory and it exists, scan ./data/
// * Third, if none found in either of the previous and it exists, scan ./test_data/
// Returns a full path to the file or an empty string if none found.
func DefaultGpkg() string {
	// Directories to check in decreasing order of priority
	dirs := []string{"", "data", "test_data"}
	gpkgPath := ""
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Unable to get working directory: %v", err)
		return ""
	}

	for _, dir := range dirs {
		searchGlob := path.Join(dir, "*.gpkg")
		gpkgFiles, err := filepath.Glob(searchGlob)
		if err != nil {
			panic("Invalid glob pattern hardcoded")
		}
		if len(gpkgFiles) == 0 {
			continue
		}
		sort.Strings(gpkgFiles)
		gpkgFilename := gpkgFiles[0]
		gpkgPath = path.Clean(path.Join(wd, gpkgFilename))
		break
	}

	return gpkgPath
}
