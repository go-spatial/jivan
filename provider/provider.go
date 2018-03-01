///////////////////////////////////////////////////////////////////////////////
//
// The MIT License (MIT)
// Copyright (c) 2018 Jivan Amara
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

package provider

import (
	"context"
	"fmt"
	"sort"
	"time"

	prv "github.com/go-spatial/tegola/provider"
)

type EmptyTile struct {
	extent *[2][2]float64
	srid   uint64
}

func (_ EmptyTile) ZXY() (uint64, uint64, uint64) {
	return 0, 0, 0
}

func (et EmptyTile) Extent() (extent [2][2]float64, srid uint64) {
	if et.extent == nil {
		max := 20037508.34
		et.srid = 3857
		et.extent = &[2][2]float64{{-max, -max}, {max, max}}
	}
	return *et.extent, et.srid
}

func (et EmptyTile) BufferedExtent() (extent [2][2]float64, srid uint64) {
	if et.extent == nil {
		max := 20037508.34
		et.srid = 3857
		et.extent = &[2][2]float64{{-max, -max}, {max, max}}
	}
	return *et.extent, et.srid
}

type ErrDuplicateCollectionName struct {
	name string
}

func (e ErrDuplicateCollectionName) Error() string {
	return fmt.Sprintf("collection name '%v' already in use", e.name)
}

type tempCollection struct {
	lastAccess time.Time
	featureIds []FeatureId
}

type Provider struct {
	Tiler           prv.Tiler
	tempCollections map[string]*tempCollection
}

type FeatureId struct {
	Collection string
	FeaturePk  uint64
}

// Filter out features based on params passed
func (p *Provider) FilterFeatures(extent *[2][2]float64, collections []string, properties map[string]string) ([]FeatureId, error) {
	if len(collections) < 1 {
		var err error
		collections, err = p.CollectionNames()
		if err != nil {
			return nil, err
		}
	}
	// To maintain a consistent order for paging & testing
	sort.Strings(collections)

	fids := make([]FeatureId, 0, 100)
	for _, col := range collections {
		fs, err := p.CollectionFeatures(col, extent)
		if err != nil {
			return nil, err
		}

	NEXT_FEATURE:
		for _, f := range fs {
			for k, v := range properties {
				if v != f.Tags[k] {
					continue NEXT_FEATURE
				}
			}
			fids = append(fids, FeatureId{Collection: col, FeaturePk: f.ID})
		}

	}

	return fids, nil
}

// Create a new collection given collection/pk pairs to populate it
func (p *Provider) MakeCollection(name string, featureIds []FeatureId) (string, error) {
	collectionIds, err := p.CollectionNames()
	if err != nil {
		return "", err
	}
	for _, cid := range collectionIds {
		if name == cid {
			e := ErrDuplicateCollectionName{name: name}
			return "", e
		}
	}

	if p.tempCollections == nil {
		p.tempCollections = make(map[string]*tempCollection)
	}

	p.tempCollections[name] = &tempCollection{lastAccess: time.Now(), featureIds: featureIds}
	return name, nil
}

// Get all features for a particular collection
func (p *Provider) CollectionFeatures(collectionName string, extent *[2][2]float64) ([]*prv.Feature, error) {
	// return a temp collection with this name if there is one
	for tcn, _ := range p.tempCollections {
		if collectionName == tcn {
			p.tempCollections[collectionName].lastAccess = time.Now()
			return p.GetFeatures(p.tempCollections[collectionName].featureIds)
		}
	}

	// otherwise hit the Tiler provider to get features for this collectionName
	pFs := make([]*prv.Feature, 0, 100)

	getFeatures := func(f *prv.Feature) error {
		pFs = append(pFs, f)
		return nil
	}

	t := EmptyTile{extent: extent, srid: 4326}
	err := p.Tiler.TileFeatures(context.TODO(), collectionName, t, getFeatures)
	if err != nil {
		return nil, err
	}

	return pFs, nil
}

// Get features given collection/pk pairs
func (p *Provider) GetFeatures(featureIds []FeatureId) ([]*prv.Feature, error) {
	// Feature pks grouped by collection
	cf := make(map[string][]uint64)
	fcount := 0
	for _, fid := range featureIds {
		if _, ok := cf[fid.Collection]; !ok {
			cf[fid.Collection] = make([]uint64, 0, 100)
		}
		cf[fid.Collection] = append(cf[fid.Collection], fid.FeaturePk)
		fcount += 1
	}

	// Desired features
	fs := make([]*prv.Feature, 0, fcount)
	for col, fpks := range cf {
		colFs, err := p.CollectionFeatures(col, nil)
		if err != nil {
			return nil, err
		}

		for _, colF := range colFs {
			for _, fpk := range fpks {
				if colF.ID == fpk {
					fs = append(fs, colF)
					break
				}
			}
		}
	}

	return fs, nil
}

// Fetch a list of all collection names from provider
func (p *Provider) CollectionNames() ([]string, error) {
	featureTableInfo, err := p.Tiler.Layers()
	if err != nil {
		return nil, err
	}

	ftNames := make([]string, len(featureTableInfo))
	for i, fti := range featureTableInfo {
		ftNames[i] = fti.Name()
	}
	sort.Strings(ftNames)

	return ftNames, err
}
