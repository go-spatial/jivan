package html

import (
	"bytes"
	"html/template"

	"github.com/go-spatial/go-wfs/config"
	"github.com/go-spatial/go-wfs/wfs3"
)

var tmpl_conformance = `<!doctype html>
<html lang="en">
	<head>
	<meta charset="utf-8">
	<title>{{ .Config.Metadata.Identification.Title }}</title>
</head>
<body>
	<h1>{{ .Config.Metadata.Identification.Title }}</h1>
	<h2>Conformance</h2>
	<ul>
	{{ range .Data.ConformsTo }}
	<li><a href="{{ . }}">{{ . }}</a></li>
	{{ end }}
	</ul>
</body>
</html>`

func RenderConformanceHTML(c config.Config, r *wfs3.ConformanceClasses) ([]byte, error) {
	var tpl bytes.Buffer

	t := template.New("conformance")
	t, _ = t.Parse(tmpl_conformance)

	data := HTMLTemplateDataConformance{c, r}

	if err := t.Execute(&tpl, data); err != nil {
		return tpl.Bytes(), err
	}

	// FIXME: should be a better way
	return tpl.Bytes(), nil
}
