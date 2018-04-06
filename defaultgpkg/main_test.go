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

package defaultgpkg

import (
	"log"
	"os"
	"path"
	"runtime"
	"testing"
)

var baseTestPath string

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	baseTestPath = path.Join(path.Dir(thisFilePath), "../test_data/main.defaultGpkg/")
}

func TestDefaultGpkg(t *testing.T) {
	// For use when a data source isn't provided.
	// Scans for files w/ '.gpkg' extension and chooses the first alphabetically.
	// * First scan the working directory
	// * Second, if none found in the working directory and it exists, scan ./data/
	// * Third, if none found in either of the previous and it exists, scan ./test_data/
	// Returns a full path to the file or an empty string if none found.

	cases := []struct {
		// The name of the sub-directory, under ./test_data/main.defaultGpkg/
		in string

		// The full path to the selected file, or an empty string if no file is found
		want string
	}{
		{
			in:   "1",
			want: "",
		},
	}

	for _, c := range cases {
		casePath := path.Join(baseTestPath, c.in)

		if chdirErr := os.Chdir(casePath); chdirErr != nil {
			log.Fatalf("Could not change working directory to '%s'.", casePath)
		}

		got := DefaultGpkg()

		if got != c.want {
			t.Errorf("Under '%s', defaultGpkg() == '%s', wanted '%s'.", casePath, got, c.want)
		}
	}
}
