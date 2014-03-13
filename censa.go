package main

import (
	"fmt"
	"database/sql"
	"os"
	"log"
//	"encoding/json"
	"encoding/xml"
	"strconv"
//	"html"
	"html/template"
	"bytes"
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
	Url string
	Title string
}

type Node struct {
	Id int  `json:"id"`
	Reflexive bool `json:"reflexive"`
	Type string `json:"type"`// in json, it should be "type"
	Desc string `json:"desc"`
	index int `json:"index"`// forgot how it worked... not important here
	Weight int `json:"weight"`
	// they're irrelevant here (calculated by app.js)
	X int `json:"x"`
	Y int `json:"y"`
	Px int `json:"px"`
	Py int `json:"py"`
}

type Link struct {
	Source int `json:"source"`
	Target int `json:"target"`
	Left bool `json:"left"`
	Right bool `json:"right"`
//	style string  --> do it later!  seek for the minimal implementation
}

type Graph struct {
	Nodes [5]Node `json:"nodes"`
	Links [5]Link `json:"links"`
}

const ContentType = "html"
type Content struct {
	ContentType string  `xml:"type,attr"`
	CDATA string `xml:"![CDATA["` // create tags <![CDATA[> and replace them later
}
type Ref struct {
	Refname string `xml:"name,attr"`
}
type Dpreds struct {
	Abnode_ref []Ref `xml:"abnode-ref,omitempty"`
}
type Dsuccs struct {
	Abnode_ref []Ref `xml:"abnode-ref"`
}
type Abnode struct {
	XMLName   xml.Name `xml:"abnode"`
	Name string `xml:"name,attr"`
	Dpreds Dpreds `xml:"dpreds,omitempty"` // omitempty not working?
	Dsuccs Dsuccs `xml:"dsuccs,omitempty"` // omitempty not working?
	Content Content `xml:"content"`
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
	//	fmt.Println("Query:" + sqlStmt)
	Visits := make(map[int]Visit)
	
	raws, err := db.Query(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
	//	fmt.Println(raws) // no prob so far
	// Nodes := [limit]Node{}
	// Links := [limit]Link{}
	
	//	Id2title := make(map[int]string)
	Abnodes := make(map[int]Abnode)

	const cdataTemp = `<a href='{{.Url}}'>{{.Title}}</a>`
	t := template.Must(template.New("cdataTemplate").Parse(cdataTemp))
	var buffer bytes.Buffer
	for raws.Next() {
		//		Visits[i] = Visit{}
		v := Visit{}
		raws.Scan(&v.id, &v.from, &v.time, &v.transition, &v.Url, &v.Title)
		Visits[v.id] = v
		//		Id2title[v.id] = v.title
		//		fmt.Println(v) // no prob so far
		// if v.from == 0 {
		// 	v.from = v.id - 1 // connect to the node right before
		// 	//link.style = "new" // --> nonminimal
		// }
		// make node
		// 		node := &Nodes[i] // must be POINTER
		// 		node.Id = v.id	// id is the same among node and visit
		// //		node.misc = misc{v.time, v.url}
		// 		node.Desc = v.title
		// 		// abced specific fields
		// 		node.Type = "A" // necessary
		// 		node.Weight = 1 // necessary
		// 		// make link
		// 		link := &Links[i]
		// 		link.Source = v.from
		// 		link.Target = v.id // always points to itself
		// 		link.Left = false
		// 		link.Right = true
		//		fmt.Println("node.id:")
		//		fmt.Println(node.id)
		buffer.Reset()
		t.Execute(&buffer, v)
		// create A-node
//		Abnodes[v.id] = Abnode{Name: strconv.Itoa(v.id), Content: Content{ContentType: ContentType, CDATA: buffer.String()}}
		Abnodes[v.id] = Abnode{Name: strconv.Itoa(v.id), Content: Content{ContentType: ContentType, CDATA: buffer.String()}}
//		fmt.Println("buffer content:",buffer.String())
	}
	
	const maskVal = 100 // id of B-node is masked by *100
	// create B-node)ake <abstructure-ref> 
	for visitId , visit := range Visits {
		sourceId := visit.from  // int
		if _, exists := Visits[sourceId]; exists {
			sref := Ref{strconv.Itoa(sourceId)}
			srefs:= []Ref{sref}
			tref := Ref{strconv.Itoa(visitId)}
			trefs:= []Ref{tref}
			masked := sourceId * maskVal
			Abnodes[masked] = Abnode{Name: strconv.Itoa(masked),Dpreds: Dpreds{srefs}, Dsuccs: Dsuccs{trefs}}
		}
	}
	type Abstructure struct {
		XMLName xml.Name `xml:"abstructure"`
		Abnodes []Abnode
	}
	
	Ablists := make([]Abnode, 0, len(Abnodes))//map2list(Abnodes)
	for _, abnode := range Abnodes {
		Ablists = append(Ablists, abnode)
	}
	xmlOutput, err := xml.MarshalIndent(Abstructure{Abnodes: Ablists}, "  ", "    ")
	if err != nil {
		log.Fatal(err)
	}
	// format CDATA tag in byte
	xmlOutput = bytes.Replace(xmlOutput, []byte("<![CDATA[>"), []byte("<![CDATA["), -1)
	xmlOutput = bytes.Replace(xmlOutput, []byte("</![CDATA[>"), []byte("]]>"), -1)
	xmlOutput = bytes.Replace(xmlOutput, []byte("&lt;"), []byte("<"), -1)
	xmlOutput = bytes.Replace(xmlOutput, []byte("&gt;"), []byte(">"), -1)
	xmlOutput = bytes.Replace(xmlOutput, []byte("&#39;"), []byte("'"), -1)
	os.Stdout.Write(xmlOutput)
	
//	fmt.Println(Visits) no prob for Visits

	// (II)  interpret the data as graph (json)
//	n := Link{13, 12, false, true}
//	fmt.Println("Link struct:",n)
//	nj, err := json.Marshal(n)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// os.Stdout.Write(nj)
	// fmt.Println("stdout of Links")
	// linksJ, _ := json.Marshal(Links)
	// os.Stdout.Write(linksJ)
	// fmt.Println("")

	// fmt.Println("Nodes:")
	// fmt.Println(Nodes)
	// nodesJ, _ := json.Marshal(Nodes)
	// os.Stdout.Write(nodesJ)
	// fmt.Println("\n")
// 	graph := Graph{Nodes, Links}
// //	fmt.Println(graph)
// 	jsonData, err := json.Marshal(graph)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// //	fmt.Println(jsonData) gives empty return
// 	// (III) output json file to stdout.
// 	os.Stdout.Write(jsonData)
}
// func output(data []byte) {
// 	encoder := json.NewEncoder(os.Stdout)
// 	encoder.Encode(data)
// }
func moveToDir() { // should take argument as a filename
	err := os.Chdir("/Users/DRW/Desktop/")
	if err != nil{
		log.Fatal(err)
	}
	// dir, err:= os.Getwd()
	// if err != nil{
	// 	log.Fatal(err)
	// }
//	fmt.Println("the current dir:"+dir)

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

//func map2list (map[
