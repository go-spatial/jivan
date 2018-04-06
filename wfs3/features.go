package wfs3

import (
	"fmt"
	"hash/fnv"

	"github.com/go-spatial/go-wfs/data_provider"
	"github.com/go-spatial/tegola/geom/encoding/geojson"
)

func Feature(cname string, fid uint64, p *data_provider.Provider, checkOnly bool) (content *geojson.Feature, contentId string, err error) {
	// TODO: This calculation of contentId assumes an unchanging data set.
	// 	When a changing data set is needed this will have to be updated, hopefully after data providers can tell us
	// 	something about updates.
	hasher := fnv.New64()
	hasher.Write([]byte(fmt.Sprintf("%v%v", cname, fid)))
	contentId = fmt.Sprintf("%x", hasher.Sum64())

	if checkOnly {
		return nil, contentId, nil
	}

	pfs, err := p.GetFeatures(
		[]data_provider.FeatureId{
			{Collection: cname, FeaturePk: fid},
		})
	if err != nil {
		return nil, "", err
	}

	if len(pfs) != 1 {
		return nil, "", fmt.Errorf("Invalid collection/fid: %v/%v", cname, fid)
	}

	pf := pfs[0]
	content = &geojson.Feature{
		ID: &pf.ID, Geometry: geojson.Geometry{Geometry: pf.Geometry}, Properties: pf.Tags,
	}

	return content, contentId, nil
}

func FeatureCollection(cName string, startIdx, stopIdx uint, p *data_provider.Provider, checkOnly bool) (content *geojson.FeatureCollection, contentId string, err error) {
	// TODO: This calculation of contentId assumes an unchanging data set.
	// 	When a changing data set is needed this will have to be updated, hopefully after data providers can tell us
	// 	something about updates.
	hasher := fnv.New64()
	hasher.Write([]byte(cName))
	contentId = fmt.Sprintf("%x", hasher.Sum64())

	if checkOnly {
		return nil, contentId, nil
	}

	// all collection features
	cfs, err := p.CollectionFeatures(cName, nil)
	if err != nil {
		return nil, "", err
	}

	uLenCfs := uint(len(cfs))
	originalStopIdx := stopIdx
	if stopIdx > uLenCfs {
		stopIdx = uLenCfs
	}

	if startIdx >= uLenCfs || stopIdx < startIdx {
		return nil, "", fmt.Errorf(
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
	content = &geojson.FeatureCollection{Features: gfs}

	return content, contentId, nil
}
