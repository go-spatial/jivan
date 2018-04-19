package wfs3

var tmpl_base = `<!doctype html>
<html lang="en">
	<head>
	<meta charset="utf-8">
	<title>{{ .config.Metadata.Identification.Title }}</title>
	{{ range .links }}
	<link rel="{{ .Rel }}" type="application/json" href="{{ .Href }}"/>
	{{ end }}
</head>
<body>
	{{ .body }}
</body>
</html>`

var tmpl_root = `
<h2>Links</h2>
	<ul>
	{{ range .data.Links }}
	<li><a href="{{ .Href }}?f=text/html">{{ .Href }}?f=text/html</a></li>
	{{ end }}
	</ul>`

var tmpl_conformance = `
<h2>Conformance</h2>
        <ul>
        {{ range .data.ConformsTo }}
        <li><a href="{{ . }}">{{ . }}</a></li>
        {{ end }}
        </ul>`
