///////////////////////////////////////////////////////////////////////////////
//
// The MIT License (MIT)
// Copyright (c) 2018 Jivan Amara
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

// go-wfs project root.go

package server

import "fmt"

func root() rootContent {
	apiUrl := fmt.Sprintf("http://%v/api", serveAddress)
	conformanceUrl := fmt.Sprintf("http://%v/conformance", serveAddress)
	collectionsUrl := fmt.Sprintf("http://%v/collections", serveAddress)

	r := rootContent{
		Links: []*link{
			&link{
				Href: apiUrl,
				Rel:  "service",
			},
			&link{
				Href: conformanceUrl,
				Rel:  "conformance",
			},
			&link{
				Href: collectionsUrl,
				Rel:  "data",
			},
		},
	}

	return r
}
