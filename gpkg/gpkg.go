package gpkg

import (
	"database/sql"
	"fmt"
)

func OpenGPKG(filepath string) (g GPKG) {
	db, err := getGpkgConnection(filepath)
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

type GPKG struct {
	Filepath      string
	DB            *sql.DB
	featureTables []string
}

func (g *GPKG) FeatureTables() []string {
	return g.featureTables
}

func (g *GPKG) populateFeatureTableNames() {
	qtext := "SELECT * FROM gpkg_contents WHERE data_type = 'features';"
	qparams := []interface{}{}
	rows, err := g.DB.Query(qtext, qparams...)
	if err != nil {
		panic(fmt.Sprintf("Problem getting gpkg contents for %v", g.Filepath))
	}

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
