package gpkg

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"regexp"

	"github.com/terranodo/tegola/geom/encoding/geojson"
	"github.com/terranodo/tegola/geom/encoding/wkb"
	"github.com/terranodo/tegola/provider/gpkg"
)

var idFieldname string = "fid"
var geomFieldname string = "geom"

func OpenGPKG(filepath string) (g GPKG) {
	db, err := gpkg.GetGpkgConnection(filepath)
	if err != nil {
		panic(fmt.Sprintf("Unable to open gpkg at '%v'", filepath))
	}

	g = GPKG{
		Filepath: filepath,
		DB:       db,
	}
	g.populateFeatureTableNames()

	return g
}

func CloseGPKG(filepath string) {
	gpkg.ReleaseGpkgConnection(filepath)
}

type GPKG struct {
	Filepath      string
	DB            *sql.DB
	featureTables []string
}

func (g *GPKG) FeatureTables() []string {
	return g.featureTables
}

func safeCollectionName(rawCollectionName string) string {
	// Treat all alphanumeric characters plus underscore as safe
	regexStr := `![a-zA-Z0-9_]`
	unsafeChars, err := regexp.Compile(regexStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid regexStr: '%v'", regexStr))
	}

	safeCollName := string(unsafeChars.ReplaceAll([]byte(rawCollectionName), []byte("")))
	return safeCollName
}

func (g *GPKG) CollectionFeatureIds(collectionName string) ([]int, error) {
	safeCollName := safeCollectionName(collectionName)
	qtext := fmt.Sprintf("SELECT %v FROM %v;", idFieldname, safeCollName)

	qparams := []interface{}{idFieldname, string(collectionName)}
	rows, err := g.DB.Query(qtext, qparams...)
	if err != nil {
		log.Printf("Problem getting feature ids from '%v': %v", collectionName, err)
		return nil, err
	}
	defer rows.Close()

	featureIds := make([]int, 10)
	for rows.Next() {
		var id int
		rows.Scan(&id)
		featureIds = append(featureIds, id)
	}

	return featureIds, nil
}

func (g *GPKG) populateFeatureTableNames() {
	qtext := "SELECT * FROM gpkg_contents WHERE data_type = 'features';"
	qparams := []interface{}{}
	rows, err := g.DB.Query(qtext, qparams...)
	if err != nil {
		panic(fmt.Sprintf("Problem getting gpkg contents for %v", g.Filepath))
	}
	defer rows.Close()

	for rows.Next() {
		var tablename string
		var data_type string
		var identifier string
		var description string
		var last_change string
		var min_x float64
		var min_y float64
		var max_x float64
		var max_y float64
		var srs_id int
		rows.Scan(
			&tablename, &data_type, &identifier, &description, &last_change,
			&min_x, &min_y, &max_x, &max_y, &srs_id)
		g.featureTables = append(g.featureTables, tablename)
	}
}

func (g *GPKG) GetFeature(collectionName string, id int) ([]byte, error) {
	safeCollName := safeCollectionName(collectionName)
	qtext := fmt.Sprintf("SELECT * FROM %v WHERE %v = ?", safeCollName, idFieldname)
	qparams := []interface{}{id}
	rows, err := g.DB.Query(qtext, qparams...)
	if err != nil {
		log.Printf("Problem getting feature with id '%v' from collection '%v': %v", id, collectionName, err)
		return nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		log.Printf("Unable to identify columns for collection '%v': %v", collectionName, err)
		return nil, err
	}

	rowCount := 0
	//	var geomHeader *gpkg.GeoPackageBinaryHeader
	var wkbGeom []byte
	//	var tags map[string]interface{}
	for rows.Next() {
		// id, geomHeader, tags not currently needed
		_, _, wkbGeom, _, err = gpkg.ReadFeatureRow(cols, rows, idFieldname, geomFieldname)
		if err != nil {
			log.Printf("Problem reading feature row: %v", err)
		}
		rowCount++
	}

	switch {
	case rowCount == 0:
		return nil, fmt.Errorf("Feature not found with id '%v' in collection '%v'", id, collectionName)
	case rowCount > 1:
		log.Printf("Warning: Multiple features with id '%v' in collection '%v'\n", id, collectionName)
	}

	// Convert wkb to GeoJSON (TODO: Currently converting to WKT, need to make a GeoJSON encoder)
	wkbReader := bytes.NewReader(wkbGeom)
	geom, err := wkb.Decode(wkbReader)
	if err != nil {
		// TODO: Handle more appropriately
		panic(err.Error())
	}
	// TODO: Cast columns appropriately & package nicely
	encoding, err := geojson.Encode(geom)
	if err != nil {
		// TODO: Handle more appropriately
		panic(err.Error())
	}
	return encoding, nil
}
