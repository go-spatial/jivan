package wfs3

import (
	"fmt"

	"github.com/go-spatial/go-wfs/data_provider"
	"github.com/go-spatial/tegola/geom/encoding/geojson"
)

func Feature(cname string, fid uint64, p *data_provider.Provider) (*geojson.Feature, error) {
	pfs, err := p.GetFeatures(
		[]data_provider.FeatureId{
			data_provider.FeatureId{Collection: cname, FeaturePk: fid},
		})
	if err != nil {
		return nil, err
	}

	if len(pfs) != 1 {
		return nil, fmt.Errorf("Invalid collection/fid: %v/%v", cname, fid)
	}

	pf := pfs[0]
	gf := &geojson.Feature{
		ID: &pf.ID, Geometry: geojson.Geometry{Geometry: pf.Geometry}, Properties: pf.Tags,
	}

	return gf, nil
}

func FeatureCollection(cName string, startIdx, stopIdx uint, p *data_provider.Provider) (
	*geojson.FeatureCollection, error) {
	// all collection features
	cfs, err := p.CollectionFeatures(cName, nil)
	if err != nil {
		return nil, err
	}

	uLenCfs := uint(len(cfs))
	originalStopIdx := stopIdx
	if stopIdx > uLenCfs {
		stopIdx = uLenCfs
	}

	if startIdx >= uLenCfs || stopIdx < startIdx {
		return nil, fmt.Errorf(
			"Invalid start/stop indices [%v, %v] for collection of length %v", startIdx, originalStopIdx, uLenCfs)
	}

	// Convert the provider features to geojson features.
	gfs := make([]geojson.Feature, stopIdx-startIdx)
	for i, pf := range cfs[startIdx:stopIdx] {
		gfs[i] = geojson.Feature{
			ID: &pf.ID, Geometry: geojson.Geometry{Geometry: pf.Geometry}, Properties: pf.Tags,
		}
	}

	// Wrap the features up in a FeatureCollection
	fc := geojson.FeatureCollection{Features: gfs}

	return &fc, nil
}
