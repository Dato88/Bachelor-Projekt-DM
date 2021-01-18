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

var addr = ":80"

//Zugriff auf den HTML Ordner
var staticDir = "./HTML"

var resetDB = true

const dbFile = "serverNachrichten.db"

var mainDB *sql.DB

//HelloServer URL zum testen => localhost:80/fachbereich/studiengang/semester/group1/hello/andrej
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Server ist Online, "+req.URL.Query().Get(":name")+"!\n")
}

//CreateAcc erstellt einen neuen Account mit den nötigen Parameter
func CreateAcc(w http.ResponseWriter, req *http.Request) {
	newAcc(req.URL.Query().Get(":fdNummer"),
		req.URL.Query().Get(":firstName"),
		req.URL.Query().Get(":lastName"),
		req.URL.Query().Get(":age"),
		req.URL.Query().Get(":degreeCourse"),
		req.URL.Query().Get(":semester"))
	io.WriteString(w, req.URL.Query().Get(":firstName")+" hat einen Account erstellt!"+"\n")
}

//FindAcc alle erstellten Accounts Anzeigen lassen
func FindAcc(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Alle Accounts Anzeigen!"+"\n")
	allAcc(w)
}

//FindGroup Gruppe 1 wird Angezeigt
func FindGroup(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Gruppenchat Anzeigen!"+"\n")
	selGroup1(w)
}

//AddMSG um eine Nachricht in der DB zu Speichern
func AddMSG(w http.ResponseWriter, req *http.Request) {
	insertMSG(req.URL.Query().Get(":message"))
	io.WriteString(w, "Nachricht eingesetzt: "+req.URL.Query().Get(":message")+"\n")
}

//AddGroupMSG eine Nachricht für die Gruppe in der DB zu Speichern
func AddGroupMSG(w http.ResponseWriter, req *http.Request) {
	insertMSG(req.URL.Query().Get(":message"))
	io.WriteString(w, "Nachricht eingesetzt: "+req.URL.Query().Get(":message")+"\n")
}

//ListAllMSG alle gespeicherten Nachrichten Anzeigen
func ListAllMSG(w http.ResponseWriter, req *http.Request) {
	allMSG(w)
}

func main() {
	m := pat.New()
	m.Get("/fachbereich/studiengang/semester/group1/hello/:name", http.HandlerFunc(HelloServer))
	m.Get("/fachbereich/studiengang/semester/group1/add/:message", http.HandlerFunc(AddMSG))
	m.Get("/fachbereich/studiengang/semester/group1/list/all", http.HandlerFunc(ListAllMSG))
	//localhost:80/create/acc/fdai5761/Andrej/Miller/32/DM/5
	m.Get("/create/acc/:fdNummer/:firstName/:lastName/:age/:degreeCourse/:semester", http.HandlerFunc(CreateAcc))

	//localhost:80/acc/search
	m.Get("/acc/search", http.HandlerFunc(FindAcc))
	m.Get("/fachbereich/studiengang/semester/group1/add/:gesendetVon/:gesendetNach/:message/:gesendeteUhrzeit", http.HandlerFunc(AddGroupMSG))
	//localhost:80/fachbereich/studiengang/semester/group1/search
	m.Get("/fachbereich/studiengang/semester/group1/search", http.HandlerFunc(FindGroup))

	DbInit()

	http.Handle("/fachbereich/studiengang/semester/group1/", m)
	http.Handle("/create/", m)
	http.Handle("/acc/", m)

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
		CREATE TABLE user (
			fdNummer VARCHAR(256) PRIMARY KEY,
			vorname VARCHAR(256),
			nachname VARCHAR(256),
			alter INTEGER NULL,
			studiengang VARCHAR(256) NULL,
			semester INTEGER NULL
			);

		CREATE TABLE nachrichten (
			nachrichtID INTEGER PRIMARY KEY AUTOINCREMENT,
			gesendetVon VARCHAR(256),
			gesendetNach INTEGER,
			message VARCHAR(256) NULL,
			gesendeteUhrzeit VARCHAR(256)
			);

		CREATE TABLE chatgroup (
			groupID INTEGER PRIMARY KEY,
			groupName VARCHAR(256),
			anzUser VARCHAR(256) NULL,
			anzNachrichten INTEGER NULL,
			empfangenUhrzeit VARCHAR(256)
			);
		
		CREATE TABLE senden (
			absender VARCHAR(256),
			chatID INTEGER
			);
		
		CREATE TABLE empfangen (
			empfID INTEGER,
			chatID INTEGER
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

func newAcc(fdNummer string, firstName string, lastName string, age string, degreeCourse string, semester string) {
	stmt, err := mainDB.Prepare("INSERT INTO user(fdNummer, vorname, nachname, alter, studiengang, semester) values (?, ?, ?, ?, ?, ?)")
	checkErr(err)

	result, errExec := stmt.Exec(fdNummer, firstName, lastName, age, degreeCourse, semester)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

//allAcc Datenbank für die Suche Auswählen
func allAcc(w http.ResponseWriter) {
	stmt, err := mainDB.Prepare("SELECT * FROM user")
	checkErr(err)

	rows, errQuery := stmt.Query()
	checkErr(errQuery)

	pRallAcc(w, rows)
}

// insertMSG Die direkte Funktion um eine Nachricht in der DB zu speichern
func insertMSG(message string) {
	stmt, err := mainDB.Prepare("INSERT INTO nachrichten(message) values (?)")
	checkErr(err)

	result, errExec := stmt.Exec(message)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

//groupMSG In einer Gruppe eine Nachricht in der DB speichern
func groupMSG(message string) {
	stmt, chGr := mainDB.Prepare("INSERT INTO chatgroup(groupID) values (?, ?)")
	checkErr(chGr)

	stmt, nd := mainDB.Prepare("INSERT INTO nachrichten(nachrichtenID, gesendetVon, gesendetNach, message, gesendeteUhrzeit) values (?, ?, ?, ?, ?)")
	checkErr(nd)

	result, errExec := stmt.Exec(message)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

//allMSG alle gespeicherten Messages widergeben
func allMSG(w http.ResponseWriter) {
	stmt, err := mainDB.Prepare(" SELECT * FROM nachrichten")
	checkErr(err)

	rows, errQuery := stmt.Query()
	checkErr(errQuery)

	processRows(w, rows)
}

//processRows
func processRows(w http.ResponseWriter, rows *sql.Rows) {
	var ID int64
	var message string

	for rows.Next() {
		err := rows.Scan(&ID, &message)
		checkErr(err)

		fmt.Fprintf(w, "ID: %d, message: %s\n", ID, string(message))
	}
}

//pRallAcc (ProcessRaw) alle Acoounts Suchen
func pRallAcc(w http.ResponseWriter, rows *sql.Rows) {
	var FD string
	var vorname string
	var nachname string
	var age string
	var studiengang string
	var semester string

	for rows.Next() {
		err := rows.Scan(&FD, &vorname, &nachname, &age, &studiengang, &semester)
		checkErr(err)

		fmt.Fprintf(w, "fd-Nummer: %s, \nVorname: %s, \nNachname: %s, \nAlter: %s, \nStudiengang: %s, \nSemester: %s\n",
			string(FD), string(vorname), string(nachname), string(age), string(studiengang), string(semester))
	}
}

//groupRows
func groupRows(w http.ResponseWriter, rows *sql.Rows) {
	var vorname string
	var message string

	for rows.Next() {
		err := rows.Scan(&vorname, &message)
		checkErr(err)

		fmt.Fprintf(w, "Nachricht von: %s, Nachricht: %s\n", string(vorname), string(message))
	}
}

//checkErr
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
