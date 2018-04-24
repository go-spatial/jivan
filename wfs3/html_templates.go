package wfs3

var tmpl_base = `<!doctype html>
<html lang="en">
	<head>
	<meta charset="utf-8">
	<title>{{ .config.Metadata.Identification.Title }}</title>
	{{ range .links }}
	<link rel="{{ .Rel }}" type="application/json" href="{{ .Href }}"/>
	{{ end }}
	<link rel="stylesheet" href="https://openlayers.org/en/v4.6.5/css/ol.css" type="text/css">
	<script src="https://openlayers.org/en/v4.6.5/build/ol.js"></script>
	<style>
		.map {
			height: 400px;
			width: 100%;
			margin-bottom: 10px;
		}
		.arrow_box {
			border-radius: 5px;
			padding: 10px;
		}
		.arrow_box {
			position: relative;
			background: #fff;
			border: 1px solid #003c88;
		}
		.arrow_box:after, .arrow_box:before {
			top: 100%;
			left: 50%;
			border: solid transparent;
			content: " ";
			height: 0;
			width: 0;
			position: absolute;
			pointer-events: none;
		}
		.arrow_box:after {
			border-color: rgba(255, 255, 255, 0);
			border-top-color: #fff;
			border-width: 10px;
			margin-left: -10px;
		}
		.arrow_box:before {
			border-color: rgba(153, 153, 153, 0);
			border-top-color: #003c88;
			border-width: 11px;
			margin-left: -11px;
		}
	</style>
	<script>
		var image = new ol.style.Circle({
			radius: 5,
			fill: new ol.style.Fill({
				color: 'rgb(255, 0, 0)'
			}),
			stroke: new ol.style.Stroke({color: 'red', width: 1})
		});
		var styles = {
			'Point': new ol.style.Style({
				image: image
			}),
			'LineString': new ol.style.Style({
				stroke: new ol.style.Stroke({
					color: 'green',
					width: 1
				})
			}),
			'MultiLineString': new ol.style.Style({
				stroke: new ol.style.Stroke({
					color: 'green',
					width: 1
				})
			}),
			'MultiPoint': new ol.style.Style({
				image: image
			}),
			'MultiPolygon': new ol.style.Style({
				stroke: new ol.style.Stroke({
					color: 'yellow',
					width: 1
				}),
				fill: new ol.style.Fill({
					color: 'rgba(255, 255, 0, 0.1)'
				})
			}),
			'Polygon': new ol.style.Style({
				stroke: new ol.style.Stroke({
					color: 'blue',
					lineDash: [4],
					width: 3
				}),
				fill: new ol.style.Fill({
					color: 'rgba(0, 0, 255, 0.1)'
				})
			}),
			'GeometryCollection': new ol.style.Style({
				stroke: new ol.style.Stroke({
					color: 'magenta',
					width: 2
				}),
				fill: new ol.style.Fill({
					color: 'magenta'
				}),
				image: new ol.style.Circle({
					radius: 10,
					fill: null,
					stroke: new ol.style.Stroke({
						color: 'magenta'
					})
				})
			}),
			'Circle': new ol.style.Style({
				stroke: new ol.style.Stroke({
					color: 'red',
					width: 2
				}),
				fill: new ol.style.Fill({
					color: 'rgba(255,0,0,0.2)'
				})
			})
		};
		var styleFunction = function(feature) {
			return styles[feature.getGeometry().getType()];
		};
	</script>
	</head>
	<body>
		<header>
			<h1><a href="{{ .config.Server.URLBasePath }}?f=text/html">{{ .config.Metadata.Identification.Title }}</a></h1>
			<span itemprop="description">{{ .config.Metadata.Identification.Description }}</span>
		</header>
		{{ .body }}
		<hr/>
		<footer>Powered by <a title="go-wfs" href="https://github.com/go-spatial/go-wfs">go-wfs</a><img width="50" height="50" src="https://raw.githubusercontent.com/go-spatial/branding/master/go-spatial.png"/></footer>
	</body>
</html>`

var tmpl_root = `
<h2><a href="conformance?f=text/html">Conformance</a></h2>
<h2><a href="collections?f=text/html">Collections</a></h2>
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
	<div><a href="./{{ .data.Name }}/items?f=text/html">Browse Features</a></div>
	<h2>Links</h2>
	<ul>
		{{ range .data.Links }}
		<li><a href="{{ .Href }}?f=text/html">{{ .Href }}?f=text/html</a></li>
		{{ end }}
	</ul>`

var tmpl_collection_features = `
<link rel="stylesheet" href="https://openlayers.org/en/v4.6.5/css/ol.css" type="text/css">
<script src="https://openlayers.org/en/v4.6.5/build/ol.js"></script>
<h2>Features</h2>
	<h2>Links</h2>
	<table>
		<tr>
			<td>
	<ul>
	{{ range .data.Features }}
		<li><a href="items/{{ .ID }}?f=text/html">{{ .ID }}</a></li>
	{{ end }}
	</ul>
			</td>
			<td>
				<div id="map" class="map"></div>
				<div id="popup-container" class="arrow_box"></div>
			</td>
		</tr>
	</table>
	<script>

      var image = new ol.style.Circle({
        radius: 5,
        fill: new ol.style.Fill({
          color: 'rgb(255, 0, 0)'
        }),
        stroke: new ol.style.Stroke({color: 'red', width: 1})
      });

      var styles = {
        'Point': new ol.style.Style({
		image: image
        }),
        'LineString': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'green',
            width: 1
          })
        }),
        'MultiLineString': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'green',
            width: 1
          })
        }),
        'MultiPoint': new ol.style.Style({
		image: image
        }),
        'MultiPolygon': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'yellow',
            width: 1
          }),
          fill: new ol.style.Fill({
            color: 'rgba(255, 255, 0, 0.1)'
          })
        }),
        'Polygon': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'blue',
            lineDash: [4],
            width: 3
          }),
          fill: new ol.style.Fill({
            color: 'rgba(0, 0, 255, 0.1)'
          })
        }),
        'GeometryCollection': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'magenta',
            width: 2
          }),
          fill: new ol.style.Fill({
            color: 'magenta'
          }),
          image: new ol.style.Circle({
            radius: 10,
            fill: null,
            stroke: new ol.style.Stroke({
              color: 'magenta'
            })
          })
        }),
        'Circle': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'red',
            width: 2
          }),
          fill: new ol.style.Fill({
            color: 'rgba(255,0,0,0.2)'
          })
        })
      };

      var styleFunction = function(feature) {
        return styles[feature.getGeometry().getType()];
      };
		var geojsonObject = {{ .data }};

		var vectorSource = new ol.source.Vector({
			features: (new ol.format.GeoJSON()).readFeatures(geojsonObject, {
				dataProjection: "EPSG:4326",
				featureProjection: "EPSG:3857"
			})
		});
		var vectorLayer = new ol.layer.Vector({
			source: vectorSource,
			style: styleFunction,
		});

		var map = new ol.Map({
			layers: [
				new ol.layer.Tile({
					source: new ol.source.OSM()
				}),
				vectorLayer
			],
			target: 'map',
			controls: ol.control.defaults({
				attributionOptions: {
					collapsible: false
				}
			}),
			view: new ol.View({
			//    projection: 'EPSG:4326',
				//center: [0, 0],
				zoom: -10
			})
		});
		map.getView().fit(vectorLayer.getSource().getExtent(), map.getSize());

		var overlay = new ol.Overlay({
			element: document.getElementById('popup-container'),
			positioning: 'bottom-center',
			offset: [0, -10]
		});
		map.addOverlay(overlay);

		map.on('click', function(e) {
			overlay.setPosition();
			var features = map.getFeaturesAtPixel(e.pixel);
			if (features) {
				var identifier = features[0].getId();
				var coords = features[0].getGeometry().getCoordinates();
				var hdms = ol.coordinate.toStringHDMS(ol.proj.toLonLat(coords));
				var popup = '<a href="items/' + identifier + '?f=text/html">' + identifier + '</a>';
				overlay.getElement().innerHTML = popup;
				overlay.setPosition(coords);
			}
		});
	</script>`

var tmpl_collection_feature = `
<link rel="stylesheet" href="https://openlayers.org/en/v4.6.5/css/ol.css" type="text/css">
<script src="https://openlayers.org/en/v4.6.5/build/ol.js"></script>
<h2>Features</h2>
	<h2>Links</h2>
	<table>
		<tr>
			<td>
	<h3>Properties</h3>
	<ul>
	{{ range $key, $value := .data.Properties }}
		<li>{{ $key }}: {{ $value }}</li>
	{{ end }}
	</ul>
			</td>
			<td>
				<div id="map" class="map"></div>
				<div id="popup-container" class="arrow_box"></div>
			</td>
		</tr>
	</table>
	<script>

      var image = new ol.style.Circle({
        radius: 5,
        fill: new ol.style.Fill({
          color: 'rgb(255, 0, 0)'
        }),
        stroke: new ol.style.Stroke({color: 'red', width: 1})
      });

      var styles = {
        'Point': new ol.style.Style({
		image: image
        }),
        'LineString': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'green',
            width: 1
          })
        }),
        'MultiLineString': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'green',
            width: 1
          })
        }),
        'MultiPoint': new ol.style.Style({
		image: image
        }),
        'MultiPolygon': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'yellow',
            width: 1
          }),
          fill: new ol.style.Fill({
            color: 'rgba(255, 255, 0, 0.1)'
          })
        }),
        'Polygon': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'blue',
            lineDash: [4],
            width: 3
          }),
          fill: new ol.style.Fill({
            color: 'rgba(0, 0, 255, 0.1)'
          })
        }),
        'GeometryCollection': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'magenta',
            width: 2
          }),
          fill: new ol.style.Fill({
            color: 'magenta'
          }),
          image: new ol.style.Circle({
            radius: 10,
            fill: null,
            stroke: new ol.style.Stroke({
              color: 'magenta'
            })
          })
        }),
        'Circle': new ol.style.Style({
          stroke: new ol.style.Stroke({
            color: 'red',
            width: 2
          }),
          fill: new ol.style.Fill({
            color: 'rgba(255,0,0,0.2)'
          })
        })
      };

      var styleFunction = function(feature) {
        return styles[feature.getGeometry().getType()];
      };
		var geojsonObject = {{ .data }};

		var vectorSource = new ol.source.Vector({
			features: (new ol.format.GeoJSON()).readFeatures(geojsonObject, {
				dataProjection: "EPSG:4326",
				featureProjection: "EPSG:3857"
			})
		});
		var vectorLayer = new ol.layer.Vector({
			source: vectorSource,
			style: styleFunction,
		});

		var map = new ol.Map({
			layers: [
				new ol.layer.Tile({
					source: new ol.source.OSM()
				}),
				vectorLayer
			],
			target: 'map',
			controls: ol.control.defaults({
				attributionOptions: {
					collapsible: false
				}
			}),
			view: new ol.View({
			//    projection: 'EPSG:4326',
				//center: [0, 0],
				zoom: -10
			})
		});
		map.getView().fit(vectorLayer.getSource().getExtent(), map.getSize());
		map.getView().setZoom(15);

		var overlay = new ol.Overlay({
			element: document.getElementById('popup-container'),
			positioning: 'bottom-center',
			offset: [0, -10]
		});
		map.addOverlay(overlay);
	</script>`
