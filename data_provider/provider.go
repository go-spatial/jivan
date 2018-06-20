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

// jivan project provider.go

package data_provider

// Builds upon the tegola Tiler interface to reuse data providers from tegola.
// Instantiate by:
//	p := Provider{Tiler: <my Tiler-based provider>}

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-spatial/geom"
	prv "github.com/go-spatial/tegola/provider"
)

type BadTimeString struct {
	msg string
}

func (bts *BadTimeString) Error() string {
	return bts.msg
}

type EmptyTile struct {
	extent *geom.Extent
	srid   uint64
}

func (_ EmptyTile) ZXY() (uint, uint, uint) {
	return 0, 0, 0
}

func (et EmptyTile) Extent() (extent *geom.Extent, srid uint64) {
	if et.extent == nil {
		max := 20037508.34
		et.srid = 3857
		et.extent = &geom.Extent{-max, -max, max, max}
	}
	return et.extent, et.srid
}

func (et EmptyTile) BufferedExtent() (extent *geom.Extent, srid uint64) {
	if et.extent == nil {
		max := 20037508.34
		et.srid = 3857
		et.extent = &geom.Extent{-max, -max, max, max}
	}
	return et.extent, et.srid
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

func parse_time_string(ts string) (t time.Time, err error) {
	fmtstrings := []string{
		"2006-01-02T15:04:05Z-0700",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, fmts := range fmtstrings {
		t, err = time.Parse(fmts, ts)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, &BadTimeString{msg: fmt.Sprintf("unable to parse time string: '%v'", ts)}
}

// Checks for any intersection of start_time - stop_time period or timestamp value
// If the feature has none of these tags, we'll consider it non-intersecting.
// If only one of start_time or stop_time is provided, the other will be considered
//	infitity or negative infinity respectively.
func feature_time_intersects_time_filter(f *prv.Feature, start_time_str, stop_time_str, timestamp_str string) (bool, error) {
	// --- Collect any time parameters from feature's tags
	// Feature start, feature stop, feature timestamp
	var fstart_str, fstop_str, fts_str string

	for k, v := range f.Properties {
		switch k {
		case "start_time":
			fstart_str = v.(string)
		case "stop_time":
			fstop_str = v.(string)
		case "timestamp":
			fts_str = v.(string)
		}
	}

	// --- Convert all time strings to time.Time instances
	var start_time, stop_time, timestamp, fstart, fstop, fts time.Time
	times := []time.Time{start_time, stop_time, timestamp, fstart, fstop, fts}
	timestrings := []string{start_time_str, stop_time_str, timestamp_str, fstart_str, fstop_str, fts_str}
	if len(times) != len(timestrings) {
		panic("array length mismatch")
	}
	var err error
	for i := 0; i < len(times); i++ {
		if timestrings[i] == "" {
			continue
		}
		times[i], err = parse_time_string(timestrings[i])
		if err != nil {
			return false, err
		}
	}

	// if the feature doesn't have any time data, treat as a match
	if fstart_str == "" && fstop_str == "" && fts_str == "" {
		return true, nil
	}

	// if there's no start_time, but there's a stop_time and feature timestamp before the stop_time
	if start_time_str == "" && stop_time_str != "" && fts_str != "" && fts.Sub(stop_time) <= 0 {
		return true, nil
	}
	// if there's no start time, but there's a stop time and a feature start and/or stop time
	if start_time_str == "" && stop_time_str != "" && (fstart_str != "" || fstop_str != "") {
		if fstart_str != "" && fstart.Sub(stop_time) <= 0 {
			return true, nil
		}
		if fstop_str != "" && fstop.Sub(stop_time) <= 0 {
			return true, nil
		}
	}
	// If there's no stop_time, but there's a start_time and feature timestamp after the start_time
	if start_time_str != "" && stop_time_str == "" && fts_str != "" && fts.Sub(start_time) >= 0 {
		return true, nil
	}
	// If there's no stop time, but there's a start time and a feature start and/or stop time
	if start_time_str != "" && stop_time_str == "" && (fstart_str != "" || fstop_str != "") {
		if fstart_str != "" && fstart.Sub(start_time) >= 0 {
			return true, nil
		}
		if fstop_str != "" && fstop.Sub(start_time) >= 0 {
			return true, nil
		}
	}
	// If there's a start_time and stop_time {
	if start_time_str != "" && stop_time_str != "" {
		if fts_str != "" && fts.Sub(start_time) >= 0 && fts.Sub(stop_time) <= 0 {
			return true, nil
		}
		if fstart_str != "" && fstart.Sub(start_time) >= 0 && fstart.Sub(stop_time) <= 0 {
			return true, nil
		}
		if fstop_str != "" && fstop.Sub(start_time) >= 0 && fstop.Sub(stop_time) <= 0 {
			return true, nil
		}
	}
	// If there's a timestamp
	if timestamp_str != "" {
		if fts_str != "" && fts == timestamp {
			return true, nil
		}
		if fstart_str != "" && timestamp.Sub(fstart) >= 0 {
			if fstop_str == "" || timestamp.Sub(fstop) <= 0 {
				return true, nil
			}
		}
		if fstop_str != "" && timestamp.Sub(fstop) <= 0 {
			if fstart_str == "" || timestamp.Sub(fstart) >= 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

// Filter out features based on params passed
// start_time, stop_time, timestamp parameters are specifically used for timestamp filtering
// 	@see check_time_filter().
func (p *Provider) FilterFeatures(extent *geom.Extent, collections []string, properties map[string]string) ([]FeatureId, error) {
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
		fs, err := p.CollectionFeatures(col, properties, extent)
		if err != nil {
			return nil, err
		}
		for _, f := range fs {
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

// Returns f if items from properties match the properties of f.  Otherwise returns nil.
func property_filter(f *prv.Feature, properties map[string]string) (*prv.Feature, error) {
	starttime := ""
	stoptime := ""
	timestamp := ""
	for k, v := range properties {
		// --- grab any time-related properties for intersection processing instead of equality testing.
		switch k {
		case "start_time":
			starttime = v
			continue
		case "stop_time":
			stoptime = v
			continue
		case "timestamp":
			timestamp = v
			continue
		}

		if v != f.Properties[k] {
			return nil, nil
		}
	}
	in_time_filter, err := feature_time_intersects_time_filter(f, starttime, stoptime, timestamp)
	if err != nil {
		return nil, err
	}
	if in_time_filter {
		return f, nil
	} else {
		return nil, nil
	}
}

// Get all features for a particular collection
func (p *Provider) CollectionFeatures(collectionName string, properties map[string]string, extent *geom.Extent) ([]*prv.Feature, error) {
	// return a temp collection with this name if there is one
	for tcn := range p.tempCollections {
		if collectionName == tcn {
			p.tempCollections[collectionName].lastAccess = time.Now()
			return p.GetFeatures(p.tempCollections[collectionName].featureIds)
		}
	}

	// otherwise hit the Tiler provider to get features for this collectionName
	pFs := make([]*prv.Feature, 0, 100)

	var err error
	getFeatures := func(f *prv.Feature) error {
		if properties != nil {
			f, err = property_filter(f, properties)
			if err != nil {
				return err
			}
			if f == nil {
				return nil
			}
		}
		pFs = append(pFs, f)
		return nil
	}

	t := EmptyTile{extent: extent, srid: 4326}
	err = p.Tiler.TileFeatures(context.TODO(), collectionName, t, getFeatures)
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
		colFs, err := p.CollectionFeatures(col, nil, nil)
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
