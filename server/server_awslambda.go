// +build awslambda
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

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/akrylysov/algnhsa"
	"github.com/go-spatial/jivan/config"
	"github.com/go-spatial/jivan/data_provider"
)

var Provider data_provider.Provider

func StartServer(p data_provider.Provider) {
	Provider = p
	h := setUpRoutes()
	algnhsa.ListenAndServe(h, nil)
}

// Provides the preferred <scheme>://<host>:<port>/<base> portion of urls for use in responses.
//	Normally this mirrors the request made, but may be overriden in config & via cl args.
func serveSchemeHostPortBase(r *http.Request) string {
	// Preferred host:port
	php := config.Configuration.Server.URLHostPort
	if php == "" {
		php = r.Header["Host"][0]
	}
	php = strings.TrimRight(php, "/")

	// Preferred scheme
	var ps string
	fproto := r.Header["X-Forwarded-Proto"]
	if len(fproto) > 0 && fproto[0] != "" {
		ps = fproto[0]
	} else {
		ps = config.Configuration.Server.URLScheme
	}

	// Preferred base path
	pbp := strings.TrimRight(config.Configuration.Server.URLBasePath, "/")
	var stage string
	if ctx, ok := algnhsa.ProxyRequestFromContext(r.Context()); ok {
		stage = ctx.RequestContext.Stage
	}
	// If you've mapped a custom domain for the API Gateway including the stage,
	//	you don't want to include the stage name in the path.
	stage = ""

	if stage != "" {
		pbp = fmt.Sprintf("/%v%v", stage, pbp)
	}

	// Preferred scheme / host / port / base
	pshpb := fmt.Sprintf("%v://%v%v", ps, php, pbp)

	return pshpb
}
