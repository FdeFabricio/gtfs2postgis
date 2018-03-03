package query

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/fdefabricio/gtfs2postgis/config"
	"github.com/fdefabricio/gtfs2postgis/reader"
	"github.com/lib/pq"
	"github.com/nleof/goyesql"
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

func (r *Repository) PopulateTable(tableName, filePath string) error {
	rows, err := reader.CSV(filePath)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	err = r.dropTable(tx, tableName)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = r.createTable(tx, tableName)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = r.loadTable(tx, tableName, rows)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	return err
}

func (r *Repository) runQuery(tx *sql.Tx, query string, args ...interface{}) error {
	_, err := tx.Exec(query, args...)
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

func (r *Repository) createTable(tx *sql.Tx, tableName string) error {
	return r.runQuery(tx, queries[goyesql.Tag(fmt.Sprintf("create-%s-table", tableName))])
}

func (r *Repository) dropTable(tx *sql.Tx, tableName string) error {
	return r.runQuery(tx, fmt.Sprintf(queries["drop-table"], tableName))
}

func (r *Repository) loadTable(tx *sql.Tx, tableName string, rows [][]string) error {
	if len(rows) < 1 {
		return errors.New(fmt.Sprintf("load %s table: no records found in the file", tableName))
	}

	header := rows[0]
	header[0] = strings.TrimPrefix(header[0], "\uFEFF")

	return r.runCopyIn(tx, tableName, header, rows[1:])
}

func convertColumnType(column, arg string) (interface{}, error) {
	if len(arg) == 0 {
		return nil, nil
	}
	arg = strings.TrimSpace(arg)
	switch column {
	case "stop_lat", "stop_lon":
		return strconv.ParseFloat(arg, 8)
	case "bikes_allowed", "location_type", "wheelchair_accessible", "wheelchair_boarding":
		return strconv.Atoi(arg)
	default:
		return arg, nil
	}
}
