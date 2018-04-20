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
		<header>
			<h1>{{ .config.Metadata.Identification.Title }}</h1>
			<span itemprop="description">{{ .config.Metadata.Identification.Description }}</span>
		</header>
		{{ .body }}
		<footer>Powered by <a title="go-wfs" href="https://github.com/go-spatial/go-wfs">go-wfs</a></footer>
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

var tmpl_collections = `
<h2>Collections</h2>
	<ul>
	{{ range .data.Collections }}
		<li><a href="./collections/{{ .Name }}?f=text/html">{{ .Name }}</a></li>
	{{ end }}
	</ul>`

var tmpl_collection = `
<h2>{{ .data.Name }} </h2>
	<span>{{ .data.Description }}</span>
	<div><a href="./{{ .data.Name }}/items">Browse Features</a></div>
	<h2>Links</h2>
	<ul>
		{{ range .data.Links }}
		<li><a href="{{ .Href }}?f=text/html">{{ .Href }}?f=text/html</a></li>
		{{ end }}
	</ul>`

