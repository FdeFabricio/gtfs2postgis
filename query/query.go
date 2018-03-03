package query

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nleof/goyesql"
	"github.com/fdefabricio/gtfs2postgis/config"
	"github.com/fdefabricio/gtfs2postgis/reader"
	"github.com/lib/pq"
)

type Repository struct{ db *sql.DB }

var queries goyesql.Queries

func init() {
	queries = goyesql.MustParseFile("./query/queries.sql")
}

func (r *Repository) Connect(c config.DatabaseConfiguration) error {
	passwordArg := ""
	if len(c.Password) > 0 {
		passwordArg = "password=" + c.Password
	}

	var err error
	r.db, err = sql.Open(c.Driver, fmt.Sprintf("host=%s port=%d user=%s %s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, passwordArg, c.Database))

	return err
}

func (r *Repository) PopulateStopsTable(filePath string) error {
	rows, err := reader.CSV(filePath)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	err = r.dropStopsTable(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = r.createStopsTable(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = r.loadStopsTable(tx, rows)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	fmt.Println()

	return err
}

func (r *Repository) runQuery(tx *sql.Tx, queryName goyesql.Tag, args ...interface{}) error {
	_, err := tx.Exec(queries[queryName], args...)
	return err
}

func (r *Repository) runCopyIn(tx *sql.Tx, tableName string, header []string, rows [][]string) error {
	stmt, err := tx.Prepare(pq.CopyIn(tableName, header...))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, row := range rows {
		args := make([]interface{}, len(row))
		for i, arg := range row {
			args[i], err = convertColumnType(header[i], arg)
			if err != nil {
				return err
			}
		}

		_, err = stmt.Exec(args...)
		if err != nil {
			return err
		}

	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("%d rows inserted into table \"%s\"", len(rows), tableName))

	return stmt.Close()

}
func (r *Repository) createStopsTable(tx *sql.Tx) error {
	return r.runQuery(tx, "create-stops-table")
}

func (r *Repository) dropStopsTable(tx *sql.Tx) error {
	return r.runQuery(tx, "drop-stops-table")
}

func (r *Repository) loadStopsTable(tx *sql.Tx, rows [][]string) error {
	if len(rows) < 1 {
		return errors.New("load stop table: no records found in the file")
	}

	header := rows[0]
	header[0] = strings.TrimPrefix(header[0], "\uFEFF")

	return r.runCopyIn(tx, "stops", header, rows[1:])
}

func convertColumnType(column, arg string) (interface{}, error) {
	if len(arg) == 0 {
		return nil, nil
	}
	arg = strings.TrimSpace(arg)
	switch column {
	case "stop_lat", "stop_lon":
		return strconv.ParseFloat(arg, 8)
	case "location_type", "wheelchair_boarding":
		return strconv.Atoi(arg)
	default:
		return arg, nil
	}
}
