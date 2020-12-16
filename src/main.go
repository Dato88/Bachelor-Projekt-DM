package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	//muss vorher Installiert werden
	"github.com/bmizerany/pat"
	_ "github.com/mattn/go-sqlite3"
)

// Port 80 und Verzeichnis mit den HTML Dateien
var addr = ":80"
var staticDir = "./static"

// Rücksetzen der Datenbank ?
var resetDB = true

const dbFileName = "nachrichten.db"

//die Main Variable bezieht sich auf einen Datentyp der in SQL definiert worden ist
//* Referenziert sql und greift dann auf DB zu
//dadurch kann der Wert verändert werden (Pointer)
var mainDB *sql.DB

//HelloServer erstellen
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, "+req.URL.Query().Get(":name")+"!\n")
}

//AddMessage um eine Nachricht zu Speichern
func AddMessage(w http.ResponseWriter, req *http.Request) {
	//aus dem Request wird die Entprechende Nachricht geholt
	insertMessage(req.URL.Query().Get(":message"))
	//Nachricht die eingefügt wurde, wird ausgegeben
	io.WriteString(w, "Nachricht hinzugefügt: "+req.URL.Query().Get(":message")+"!\n")
}

//ListAllMessages um alle Nachrichten widerzugeben
func ListAllMessages(w http.ResponseWriter, req *http.Request) {
	allMessages(w)
}

//ListMessage umd eine bestimmte Nachricht der id widerzugeben
func ListMessage(w http.ResponseWriter, req *http.Request) {
	var msgid = req.URL.Query().Get(":msgid")
	idMessage(w, msgid)
}

//main Funktion
func main() {
	m := pat.New()
	m.Get("/api/v1/hello/:name", http.HandlerFunc(HelloServer))
	m.Get("/api/v1/message/add/:message", http.HandlerFunc(AddMessage))
	m.Get("/api/v1/message/list/id=:msgid", http.HandlerFunc(ListMessage))
	m.Get("/api/v1/message/list/all", http.HandlerFunc(ListAllMessages))

	// Datenbank initialiieren
	Initialize()

	// starte die Dienste (REST und statischer Webserver)
	http.Handle("/api/v1/", m)

	//den Pfad / mit dem Handler (der Funktion) versehen die das verarbeitet
	//wir nehmen einen FileServer geben dazu das entsprechende Verzeichnis (http.Dir)
	//staticDir wurde weiter oben Definiert. Bei ./static handelt es sich um statische html geschriebene Seiten
	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	log.Print(" Running on ", addr)       //log Ausgabe der Adresse
	err := http.ListenAndServe(addr, nil) //addr gibt die entsprechende Adresse an, nil kann eine Erweiterung sein (Websocketerweiterung)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

//Initialize initialisiert die Datenbank
//hier id mit Autoincrement und messages werden gespeichert
func Initialize() {
	sqlStmt := `
		CREATE TABLE nachrichten (id INTEGER PRIMARY KEY AUTOINCREMENT,
		message VARCHAR(256) NULL);
	`
	Create(sqlStmt)
}

//Create erstellt die Datenbank
func Create(sqlStmt string) {
	//wenn resetDB true dann wird die Datenbank wieder gelöscht
	if resetDB {
		fmt.Println("Datenbank wird gelöscht")
		os.Remove(dbFileName)
	}

	//hier wird die Datenbank "sqlite3" geöffnet
	db, err := sql.Open("sqlite3", dbFileName)
	checkErr(err)
	mainDB = db

	if resetDB {
		fmt.Println("Tabelle erzeugt")
		_, err = db.Exec(sqlStmt)
		checkErr(err)
	}
}

//insertMessage: um Nachrichten in die Datenbank einzutragen
func insertMessage(message string) {
	stmt, err := mainDB.Prepare("INSERT INTO nachrichten(message) values (?)")
	checkErr(err)

	result, errExec := stmt.Exec(message)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

//allMessages um alle gespeicherten Nachrichten samt id widerzugeben
func allMessages(w http.ResponseWriter) {
	stmt, err := mainDB.Prepare(" SELECT * FROM nachrichten")
	checkErr(err)

	rows, errQuery := stmt.Query() //Zeilen enthalten id und message
	checkErr(errQuery)

	processRows(w, rows)
}

//idMessage um eine bestimmte id Nachricht widerzugeben
func idMessage(w http.ResponseWriter, msgid string) {

	stmt, err := mainDB.Prepare(" SELECT * FROM nachrichten WHERE id = ?")
	checkErr(err)

	rows, errQuery := stmt.Query(msgid) //Zeilen enthalten id und message
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

// Hilfsfunktion

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
