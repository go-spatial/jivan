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
	"os"
	"path"
	"runtime"
	"testing"
)

var baseTestPath string

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	baseTestPath = path.Join(path.Dir(thisFilePath), "testdata")
}

func TestDefaultGpkg(t *testing.T) {
	cases := []struct {
		// The name of the sub-directory, under ./testdata/
		testSubPath string

		// The full path to the selected file, or an empty string if no file is found
		dummyFileName string
	}{
		{
			testSubPath:   "1",
			dummyFileName: path.Join(baseTestPath, "1/dummyA.gpkg"),
		},
		{
			testSubPath:   "2",
			dummyFileName: path.Join(baseTestPath, "2/data/dataDummy.gpkg"),
		},
		{
			testSubPath:   "3",
			dummyFileName: path.Join(baseTestPath, "3/test_data/testDataDummy.gpkg"),
		},
		{
			testSubPath:   "4",
			dummyFileName: "",
		},
	}

	for _, c := range cases {
		casePath := path.Join(baseTestPath, c.testSubPath)

		if chdirErr := os.Chdir(casePath); chdirErr != nil {
			t.Errorf("Could not change working directory to '%s'.", casePath)
		}

		got := Get()

		if got != c.dummyFileName {
			t.Errorf("Under '%s', defaultGpkg() == '%s', wanted '%s'.", casePath, got, c.dummyFileName)
		}
	}
}
