///////////////////////////////////////////////////////////////////////////////
//
// The MIT License (MIT)
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
	"github.com/BurntSushi/toml"
)

// Config provides an object model for configuration.
type Config struct {
	Server struct {
		URL         string
		MimeType    string
		Encoding    string
		Language    string
		PrettyPrint bool
		Limit       int
	}
	Logging struct {
		Level   string
		Logfile string
	}
	Metadata struct {
		Identification struct {
			Title             string
			Abstract          string
			Keywords          []string
			KeywordsType      string
			Fees              string
			AccessConstraints string
		}
		Provider struct {
			Name string
			URL  string
		}
		Contact struct {
			Name            string
			Position        string
			Address         string
			City            string
			StateOrProvince string
			PostalCode      string
			Country         string
			Phone           string
			Fax             string
			Email           string
			URL             string
			Hours           string
			Instructions    string
			Role            string
		}
	}
	Collections struct {
		Data string
	}
}

// LoadFromFile read YAML into configuration
func LoadConfigFromFile(tomlFile string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(tomlFile, &config); err != nil {
		return config, err
	}
	return config, nil
}
