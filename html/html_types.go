package html

import (
	"github.com/go-spatial/go-wfs/config"
	"github.com/go-spatial/go-wfs/wfs3"
)

type HTMLTemplateDataRoot struct {
	Config config.Config
	Data   *wfs3.RootContent
}
