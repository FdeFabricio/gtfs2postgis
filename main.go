package main

import (
	"fmt"
	"path"
	"runtime"

	"github.com/fdefabricio/gtfs2postgis/config"
	"github.com/fdefabricio/gtfs2postgis/query"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	conf *config.Configuration
	repo *query.Repository
)

func init() {
	conf = new(config.Configuration)
	repo = new(query.Repository)

	err := config.Init(conf)
	if err != nil {
		panic(err)
	}
}

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	dir := path.Dir(filename)

	err := repo.Connect(conf.Database)
	if err != nil {
		panic(err)
	}

	err = repo.PopulateStopsTable(fmt.Sprintf("%s/gtfs_bhtransit/stops.txt", dir))
	if err != nil {
		panic(err)
	}
}
