package gpkg

import (
	"database/sql"
	"fmt"
	"regexp"

	"log"

	"github.com/terranodo/tegola/provider/gpkg"
)

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

func (g *GPKG) CollectionFeatureIds(collectionName string) ([]int, error) {
	// Treat all non-alphanumeric characters except underscore as unsafe
	regexStr := `![a-zA-Z0-9_]`
	unsafeChars, err := regexp.Compile(regexStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid regexStr: '%v'", regexStr))
	}

	safeCollName := string(unsafeChars.ReplaceAll([]byte(collectionName), []byte("")))
	idFieldname := "fid"
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
