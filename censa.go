package main

import (
	"fmt"
	"database/sql"
	"os"
	"log"
	"encoding/json"
	"strconv"
)

//import "github.com/kisielk/sqlstruct"
import _ "github.com/mattn/go-sqlite3"

const History = "History_copy" // filename.  should be taken as an argument in the future
const limit = 5 // query LIMIT in sql ... should be a command line argument
// ...no need of type specification

type Visit struct {
	// from "visits" table
	id int // "visits.id"
	from int // id in "visits.from_visit" pointing to "visits.id"
	time int // from epoch UTC ("visits.visit_time")
	transition int
	// from "urls" table
	url string
	title string
}

type Node struct {
	id int  
	reflexive bool
	_type string `json:"type"`// in json, it should be "type"
	desc string
	index int // forgot how it worked... not important here
	weight int
	// they're irrelevant here (calculated by app.js)
	x int
	y int
	px int
	py int
}

type Link struct {
	source, target int
	left, right bool
//	style string  --> do it later!  seek for the minimal implementation
}

type Graph struct {
	nodes [5]Node
	links [5]Link
}

func main() {
	// (I)   read data from  sql "History"
	moveToDir() // for current use
	db, err := sql.Open("sqlite3", History)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	sqlStmt := chooseSqlStmt(1) // should be fully implemented
	fmt.Println("Query:" + sqlStmt)
	Visits := [limit]Visit{}

	raws, err := db.Query(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
//	fmt.Println(raws) // no prob so far
	Nodes := [limit]Node{}
	Links := [limit]Link{}

	for i:=0; raws.Next(); i++{
//		Visits[i] = Visit{}
		v := &Visits[i]
		raws.Scan(&v.id, &v.from, &v.time, &v.transition, &v.url, &v.title)
//		fmt.Println(v) // no prob so far
		if v.from == 0 {
			v.from = v.id - 1 // connect to the node right before
			//link.style = "new" // --> nonminimal
		}
		// make node
		node := &Nodes[i] // must be POINTER
		node.id = v.id	// id is the same among node and visit
//		node.misc = misc{v.time, v.url}
		node.desc = v.title
		// abced specific fields
		node._type = "A" // necessary
		node.weight = 1 // necessary
		// make link
		link := &Links[i]
		link.source = v.from
		link.target = v.id // always points to itself
		link.left = false
		link.right = true
//		fmt.Println("node.id:")
//		fmt.Println(node.id)
	}
//	fmt.Println(Visits) no prob for Visits
//	fmt.Println(Nodes)
//	fmt.Println(Links)
	// (II)  interpret the data as graph (json)
	graph := Graph{Nodes, Links}
//	fmt.Println(graph)
	jsonData, err := json.Marshal(graph)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(jsonData)
	// (III) output json file to stdout.
	os.Stdout.Write(jsonData)
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
		return "SELECT visits.id, visits.from_visit, visits.visit_time, visits.transition, urls.url, urls.title FROM visits LEFT JOIN urls ON visits.url = urls.id ORDER BY visit_time DESC LIMIT "+strconv.Itoa(limit)
	default:
		return "" // invokes error
	}
}
