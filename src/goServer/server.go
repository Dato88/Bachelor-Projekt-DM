package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bmizerany/pat"
	_ "github.com/mattn/go-sqlite3"
)

var addr = ":8080"
var staticDir = "./static"

var resetDB = true

const dbFile = "serverNachrichten.db"

var mainDB *sql.DB

//HelloServer URL zum testen => localhost:8080/group1/hello/andrej
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Server ist Online, "+req.URL.Query().Get(":name")+"!\n")
}

//AddMSG um eine Nachricht zu Speichern
func AddMSG(w http.ResponseWriter, req *http.Request) {
	insertMSG(req.URL.Query().Get(":message"))
	io.WriteString(w, "Nachricht eingesetzt: "+req.URL.Query().Get(":message")+"!\n")
}

func ListAllMSG(w http.ResponseWriter, req *http.Request) {
	allMSG(w)
}

func main() {
	m := pat.New()
	m.Get("/group1/hello/:name", http.HandlerFunc(HelloServer))
	m.Get("/group1/add/:message", http.HandlerFunc(AddMSG))
	m.Get("/group1/list/all", http.HandlerFunc(ListAllMSG))

	DbInit()

	http.Handle("/group1/", m)

	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	log.Print("Running on Port: ", addr)

	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatal("ListenAndServe Fehler: ", err)
	}
}

//DbInit initialisiert die Datenbank
func DbInit() {
	sqlStatement := `
		CREATE TABLE nachrichten (
			nachrichtID INTEGER PRIMARY KEY AUTOINCREMENT,
			fdNummer VARCHAR(256), 
			message VARCHAR(256),
			gesendeteUhrzeit VARCHAR(256)
			);

		CREATE TABLE user (
			fdNummer VARCHAR(256) PRIMARY KEY,
			status INTEGER,
			name VARCHAR(256),
			age INTEGER,
			studiengang VARCHAR(256),
			semester INTEGER
			);

		CREATE TABLE chatgroup (
			anzNachrichten INTEGER,
			charID INTEGER PRIMARY KEY,
			anzUser VARCHAR(256),
			empfangenUhrzeit VARCHAR(256)
			);
		
		CREATE TABLE senden (
			fdNummer VARCHAR(256),
			nachrichtenID INTEGER PRIMARY KEY
			);
		
		CREATE TABLE empfangen (
			nachrichtenID INTEGER PRIMARY KEY,
			charID INTEGER
			);
		
		CREATE TABLE zeigt_an (
			charID INTEGER,
			fdNummer VARCHAR(256) PRIMARY KEY
			);
	`

	Create(sqlStatement)
}

//Create erstellt die Datenbank
func Create(sqlStatement string) {
	if resetDB {
		fmt.Println("Alte Datenbank wird gelöscht!")
		os.Remove(dbFile)
	}

	db, err := sql.Open("sqlite3", dbFile)
	checkErr(err)

	mainDB = db

	if resetDB {
		fmt.Println("Tabelle wurde erstellt!")
		_, err = db.Exec(sqlStatement)
		checkErr(err)
	}
}

// insertMSG um eine Nachricht in der DB zu speichern
func insertMSG(message string) {
	stmt, err := mainDB.Prepare("INSERT INTO nachrichten(message) values (?)")
	checkErr(err)

	result, errExec := stmt.Exec(message)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

func allMSG(w http.ResponseWriter) {
	stmt, err := mainDB.Prepare(" SELECT * FROM nachrichten")
	checkErr(err)

	rows, errQuery := stmt.Query()
	checkErr(errQuery)

	processRows(w, rows)
}

func processRows(w http.ResponseWriter, rows *sql.Rows) {
	var ID int64
	var message string

	for rows.Next() { //Durchlaufen mit Cursor (Datenbankzeiger)
		//& ist ein Adressoperator
		err := rows.Scan(&ID, &message)
		checkErr(err)

		//F steht für File
		//Fprintf bekommt einen Writer und einen Formatstring
		fmt.Fprintf(w, "ID: %d, message: %s\n", ID, string(message))
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
