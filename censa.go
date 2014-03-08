package main

import (
	"fmt"
	"sql"
	_ "go-sqlite"
	"os"
	"log"
	"encoding/json"
	"strconv"
)

const History = "History_copy" // filename.  should be taken as an argument in the future
const limit = 10 // query LIMIT in sql ... should be a command line argument
// ...no need of type specification

type visit struct{
	// from "visits" table
	id int // "visits.id"
	from int // id in "visits.from_visit" pointing to "visits.id"
	time int // from epoch UTC ("visits.visit_time")
	transition int
	// from "urls" table
	url string
	title string
}

type node struct{
	id int 
	reflexive bool
	_type string `json:"type"`// in json, it should be "type"
	desc string
	index int // forgot how it worked...
	weight int
	x int
	y int
	px int
	py int
}

func main() {
	// (I)   read data from  sql "History"
	moveToDir() // for current use
	db, err := sql.Open("sqlite3", History)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	sql := chooseSqlStmt(1) // should be fully implemented

	visits := [limit]visit

	raws, err := db.Query(sql)
	if err != nil {
		log.Fatal(err)
	}
	
	for i:=0; raws.Next() {
		visits[
	}
	// (II)  interpret the data as graph (json)
	// (III) output json file to stdout.
	
}
func output(data []byte) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(data)
}
func moveToDir() { // should take argument as a filename
	err := os.Chdir("/Users/DRW/Desktop/")
	if err != nil{
		log.Fatal(err)
	}
	dir, err:= os.Getwd()
	if err != nil{
		log.Fatal(err)
	}
	fmt.Println("the current dir:"+dir)

	if _, err := os.Stat(History); os.IsNotExist(err){
		fmt.Printf("no such file: %s\n", History)
	}
}

func chooseSqlStmt(flag int) string{
	// switch sql statement according to the argument
	switch flag {
	case 1:
		return "SELECT visits.id, visits.from_visit, visits.visit_time, vists.transition, urls.url, urls.title FROM visits LEFT JOIN urls ON visits.url = urls.id ORDER BY visit_time DESC LIMIT "+strconv.Itoa(limit)
	default:
		return "" // invokes error
	}
}
