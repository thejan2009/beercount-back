package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var dbMap = initDatabase()

func main() {
	e := http.ListenAndServe(":8080", newRouter())
	checkErr(e, "Server problem")
}

// =======
// routing
// =======

type route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

func newRouter() *mux.Router {
	rs := mux.NewRouter().StrictSlash(true)
	for _, r := range appRoutes {
		rs.Methods(r.Method).Path(r.Pattern).HandlerFunc(r.Handler).Name(r.Name)
	}

	return rs
}

var appRoutes = []route{
	route{"index", "GET", "/", indexHandler},

	route{"beerList", "GET", "/beer", beerList},
	route{"createBeer", "POST", "/beer", beerCreateHandler},

	route{"batchList", "GET", "/batch/{user}", batchListHandler},
	route{"createBatch", "POST", "/batch", batchCreateHandler},
}

// ========
// handlers
// ========

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world!")
}

func beerList(w http.ResponseWriter, r *http.Request) {
	var beers []beer
	_, e := dbMap.Select(&beers, "SELECT * FROM beers ORDER BY beer_id")
	checkErr(e, "Beer list select failed")

	e = json.NewEncoder(w).Encode(beers)
	checkErr(e, "Problem encoding beer list to json")
}

func batchListHandler(w http.ResponseWriter, r *http.Request) {
	var batches []batch
	_, e := dbMap.Select(&batches, "SELECT * FROM batches WHERE user = ?", parseVars(r, "user"))
	checkErr(e, "Batch list select failed")

	e = json.NewEncoder(w).Encode(batches)
	checkErr(e, "Problem encoding batch list to json")
}

func batchCreateHandler(w http.ResponseWriter, r *http.Request) {
	body, e := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	checkErr(e, "Problem reading json")

	var b batch
	e = json.Unmarshal(body, &b)
	checkErr(e, "Problem unmarshalling json.")

	b = createBatch(b)
	e = json.NewEncoder(w).Encode(b)
	checkErr(e, "Problem encoding batch to json")
}

func beerCreateHandler(w http.ResponseWriter, r *http.Request) {
	body, e := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	checkErr(e, "Problem reading json")

	var b beer
	e = json.Unmarshal(body, &b)
	checkErr(e, "Problem unmarshalling json.")

	b = createBeer(b)
	e = json.NewEncoder(w).Encode(b)
	checkErr(e, "Problem encoding beer to json")
}

// =====
// utils
// =====

func checkErr(e error, msg string) {
	if e != nil {
		log.Fatalln(msg, e)
	}
}

func parseVars(r *http.Request, item string) string {
	v := mux.Vars(r)
	return v[item]
}

// ========
// database
// ========

func initDatabase() *gorp.DbMap {
	db, e := sql.Open("sqlite3", "beerCount.db")
	checkErr(e, "sql.Open")

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbMap.AddTableWithName(beer{}, "beers").SetKeys(true, "ID")
	dbMap.AddTableWithName(batch{}, "batches").SetKeys(true, "ID")

	e = dbMap.CreateTablesIfNotExists()
	checkErr(e, "Table creation failed")

	return dbMap
}

func createBatch(b batch) batch {
	e := dbMap.Insert(&b)
	checkErr(e, "batch insert")
	return b
}

func createBeer(b beer) beer {
	e := dbMap.Insert(&b)
	checkErr(e, "beer insert")
	return b
}

// =======
// structs
// =======

type beer struct {
	ID   int64  `db:"beer_id" json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type batch struct {
	ID     int64  `db:"batch_id" json:"id"`
	BeerID int64  `db:"beer_id" json:"beerId"`
	User   string `db:"user" json:"user"`
	Date   int64  `json:"date"`
	C3     int    `json:"count03"`
	C5     int    `json:"count05"`
}
