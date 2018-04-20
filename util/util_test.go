///////////////////////////////////////////////////////////////////////////////
//
// The MIT License (MIT)
// Copyright (c) 2018 James Lucktaylor
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
package util

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

var baseTestPath string

func init() {
	_, thisFilePath, _, _ := runtime.Caller(0)
	baseTestPath = path.Join(path.Dir(thisFilePath), "test_data/DefaultGpkg")
}

func TestDefaultGpkg(t *testing.T) {
	cases := []struct {
		// The name of the sub-directory, under ./test_data/DefaultGpkg/
		testSubPath string

		// The full path to the selected file, or an empty string if no file is found
		dummyFileName string
	}{
		{
			testSubPath:   "1",
			dummyFileName: filepath.Join(baseTestPath, "1/dummyA.gpkg"),
		},
		{
			testSubPath:   "2",
			dummyFileName: filepath.Join(baseTestPath, "2/data/dataDummy.gpkg"),
		},
		{
			testSubPath:   "3",
			dummyFileName: filepath.Join(baseTestPath, "3/test_data/testDataDummy.gpkg"),
		},
		{
			testSubPath:   "4",
			dummyFileName: "",
		},
	}

	for i, c := range cases {
		casePath := path.Join(baseTestPath, c.testSubPath)

		if chdirErr := os.Chdir(casePath); chdirErr != nil {
			t.Errorf("[%d] Could not change working directory to '%s'.", i, casePath)
		}

		got := DefaultGpkg()

		if got != c.dummyFileName {
			t.Errorf("[%d] Under '%s', DefaultGpkg() == '%s', wanted '%s'.", i, casePath, got, c.dummyFileName)
		}
	}
}

func TestRenderTemplate(t *testing.T) {
	var tmpl = `<foo>{{ .value }}</foo>`
	data := map[string]interface{}{"value": "bar"}
	expected := []byte("<foo>bar</foo>")

	value, _ := RenderTemplate(tmpl, data)

	if bytes.Equal(value, expected) == false {
		t.Errorf("got %s, wanted %s", value, expected)
	}
}
