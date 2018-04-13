package html

import (
	"bytes"
	"html/template"

	"github.com/go-spatial/go-wfs/config"
	"github.com/go-spatial/go-wfs/wfs3"
)

var tmpl = `<!doctype html>
<html lang="en">
	<head>
	<meta charset="utf-8">
	<title>{{ .Config.Metadata.Identification.Title }}</title>
	{{ range .Data.Links }}
	<link rel="{{ .Rel }}" type="application/json" href="{{ .Href }}"/>
	{{ end }}
</head>
<body>
	<h1>{{ .Config.Metadata.Identification.Title }}</h1>
	<h2>Links</h2>
	<ul>
	{{ range .Data.Links }}
	<li><a href="{{ .Href }}?f=text/html">{{ .Href }}?f=text/html</a></li>
	{{ end }}
	</ul>
</body>
</html>`

func RenderRootHTML(c config.Config, r *wfs3.RootContent) ([]byte, error) {
	var tpl bytes.Buffer

	t := template.New("root")
	t, _ = t.Parse(tmpl)

	data := HTMLTemplateDataRoot{c, r}

	if err := t.Execute(&tpl, data); err != nil {
		return tpl.Bytes(), err
	}

	// FIXME: should be a better way
	return tpl.Bytes(), nil
}
