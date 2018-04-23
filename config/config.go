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

package config

import (
	"github.com/BurntSushi/toml"
)

const (
	JSONContentType = "application/json"
	HTMLContentType = "text/html"
)

// These are the MIME types that the handlers support.
var SupportedContentTypes []string = []string{JSONContentType, HTMLContentType}

var Configuration Config

func init() {
	Configuration = Config{
		Server: Server{
			DefaultMimeType: JSONContentType,
			Encoding:        "utf8",
			URLScheme:       "http",
			URLBasePath:     "/",
			Language:        "en-US",
			PrettyPrint:     false,
			DefaultLimit:    10,
			MaxLimit:        1000,
		},
		Logging: Logging{
			Level:   "NONE",
			Logfile: "",
		},
		Metadata: Metadata{
			Identification: Identification{
				Title:             "go-wfs",
				Description:       "go-wfs is a Go server implementation of OGC WFS 3.0",
				Keywords:          []string{"geospatial", "features", "collections", "access"},
				KeywordsType:      "theme",
				Fees:              "None",
				AccessConstraints: "None",
			},
			ServiceProvider: ServiceProvider{
				Name: "Organization Name",
				URL:  "https://github.com/go-spatial/go-wfs",
			},
			Contact: Contact{
				Name:            "Lastname, Firstname",
				Position:        "Position Title",
				Address:         "Mailing Address",
				City:            "City",
				StateOrProvince: "Adminstrative Area",
				PostalCode:      "Zip or Postal Code",
				Country:         "Country",
				Phone:           "+xx-xxx-xxx-xxxx",
				Fax:             "+xx-xxx-xxx-xxxx",
				Email:           "you@example.org",
				URL:             "http://example.org",
				Hours:           "Hours of Service",
				Instructions:    "During hours of service.  Off on weekends.",
				Role:            "pointOfContact",
			},
		},
		Providers: Providers{},
	}
}

// Config provides an object model for configuration.
type Server struct {
	BindHost        string `toml:"bind_host"`
	BindPort        int    `toml:"bind_port"`
	URLScheme       string `toml:"url_scheme"`
	URLHostPort     string `toml:"url_hostport"`
	URLBasePath     string `toml:"url_basepath"`
	DefaultMimeType string
	Encoding        string
	Language        string
	PrettyPrint     bool
	DefaultLimit    uint
	MaxLimit        uint
}

type Logging struct {
	Level   string
	Logfile string
}

type Metadata struct {
	Identification  Identification
	ServiceProvider ServiceProvider
	Contact         Contact
}

type Identification struct {
	Title             string
	Description       string
	Keywords          []string
	KeywordsType      string
	Fees              string
	AccessConstraints string
}

type ServiceProvider struct {
	Name string
	URL  string
}

type Contact struct {
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

type Providers struct {
	Data string
}

type Config struct {
	Server    Server
	Logging   Logging
	Metadata  Metadata
	Providers Providers
}

// LoadFromFile read YAML into configuration
func LoadConfigFromFile(tomlFile string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(tomlFile, &config); err != nil {
		return config, err
	}
	return config, nil
}
