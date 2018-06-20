// +build !awslambda
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

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-spatial/jivan/config"
	"github.com/go-spatial/jivan/data_provider"
)

var Provider data_provider.Provider

func StartServer(p data_provider.Provider) {
	sconf := config.Configuration.Server
	bindAddress := fmt.Sprintf("%v:%v", sconf.BindHost, sconf.BindPort)

	fmt.Printf("Bound to: %v\n", bindAddress)
	if sconf.URLHostPort != "" {
		fmt.Printf("Expecting traffic at %v\n", sconf.URLHostPort)
	}

	Provider = p
	handler := setUpRoutes()
	err := http.ListenAndServe(bindAddress, handler)
	if err != nil {
		panic(fmt.Sprintf("Problem starting web server: %v", err))
	}
}

// Provides the preferred <scheme>://<host>:<port>/<base> portion of urls for use in responses.
//	Normally this mirrors the request made, but may be overriden in config & via cl args.
func serveSchemeHostPortBase(r *http.Request) string {
	// Preferred host:port
	php := config.Configuration.Server.URLHostPort
	if php == "" {
		php = r.Host
	}
	php = strings.TrimRight(php, "/")

	// Preferred scheme
	ps := config.Configuration.Server.URLScheme

	// Preferred base path
	pbp := strings.TrimRight(config.Configuration.Server.URLBasePath, "/")

	// Preferred scheme / host / port / base
	pshpb := fmt.Sprintf("%v://%v%v", ps, php, pbp)

	return pshpb
}
