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
var staticDir = "./templates"

var resetDB = true

type Account struct {
	FdNummer    string
	Vorname     string
	Nachname    string
	Age         int8
	Studiengang string
	Semester    int8
}

//Message hier wird der Inhalt und der Vorname der Nachrichten gespeichert
type Message struct {
	Vorname string
	Content string
}

//Chat hier wird der passende Array für die Nachrichten gespeichert
type Chat struct {
	//NachrichtStr []NachrichtStr
	UserName string
	Messages []Message
}

type Chatgroup struct {
	GroupID   int64
	GroupName string
}

const dbFile = "serverNachrichten.db"

var mainDB *sql.DB

//HelloServer URL zum testen => localhost:80/fachbereich/studiengang/semester/group1/hello/andrej
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Server ist Online, "+req.URL.Query().Get(":name")+"!\n")
}

//CreateAcc erstellt einen neuen Account mit den nötigen Parameter
func CreateAcc(w http.ResponseWriter, req *http.Request) {
	fmt.Println("CreateAcc wurde aufgerufen")
	newAcc(req.URL.Query().Get(":fdNummer"),
		req.URL.Query().Get(":firstName"),
		req.URL.Query().Get(":lastName"),
		req.URL.Query().Get(":age"),
		req.URL.Query().Get(":studiengang"),
		req.URL.Query().Get(":semester"))
	fmt.Println("CreateAcc Checkpoint 1")
	io.WriteString(w, req.URL.Query().Get(":firstName")+" hat einen Account erstellt!"+"\n")
}

//FindAcc alle erstellten Accounts Anzeigen lassen mit allen gespeicherten Informationen
func FindAcc(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Alle Accounts Anzeigen!"+"\n")
	allAcc(w)
}

//AddGroupMSG eine Nachricht für die Gruppe in der DB zu Speichern
func AddGroupMSG(w http.ResponseWriter, req *http.Request) {
	fmt.Println("AddGroupMSG wurde aufgerufen")
	insertMSG(req.URL.Query().Get(":fdNummer"),
		req.URL.Query().Get(":groupID"),
		req.URL.Query().Get(":message"))
	//io.WriteString(w, "Nachricht eingesetzt: "+req.URL.Query().Get(":message")+"\n")
}

//AddGroup Gruppe hinzufügen
func AddGroup(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Gruppe erstellt!")
	insertGroup(req.URL.Query().Get(":GroupName"))
	io.WriteString(w, "Neue Gruppe erstellt: "+req.URL.Query().Get(":GroupName")+"\n")
}

//ListAllMSG alle gespeicherten Nachrichten Anzeigen
// func ListAllMSG(req *http.Request) {
// 	groupMSG(req.URL.Query().Get(":group"))
// }

//ListMSG alle gespeicherten Nachrichten einer Gruppe Anzeigen
func ListMSG(w http.ResponseWriter, req *http.Request) {
	groupMSG(w, req)
}

func main() {
	m := pat.New()
	m.Get("/fachbereich/studiengang/semester/:groupID/hello/:name", http.HandlerFunc(HelloServer))
	//m.Get("/fachbereich/studiengang/:semester/:group/add/:message", http.HandlerFunc(AddMSG))
	//m.Get("/fachbereich/studiengang/semester/:groupID/all", http.HandlerFunc(ListAllMSG))

	//http://bachelor-community.informatik.hs-fulda.de/create/acc/fdai5761/Andrej/Miller/32/DM/5
	m.Get("/create/acc/:fdNummer/:firstName/:lastName/:age/:studiengang/:semester", http.HandlerFunc(CreateAcc))

	//http://bachelor-community.informatik.hs-fulda.de/create/group/Gruppenchat1
	m.Get("/create/group/:GroupName", http.HandlerFunc(AddGroup))

	//http://bachelor-community.informatik.hs-fulda.de/acc/search
	m.Get("/acc/search", http.HandlerFunc(FindAcc))

	//URL um eine Nachricht in gewählter Gruppe bzw. zu einer Person Speichern
	//für Leerzeichen %20 oder + verwenden
	//http://bachelor-community.informatik.hs-fulda.de/fachbereich/studiengang/semester/add/fdai5761/1/Eine+ganze%20nachricht%20Neu
	m.Get("/fachbereich/studiengang/semester/add/:fdNummer/:groupID/:message", http.HandlerFunc(AddGroupMSG))

	//http://bachelor-community.informatik.hs-fulda.de/fachbereich/studiengang/semester/search/1
	m.Get("/fachbereich/studiengang/semester/search/:gruMSG", http.HandlerFunc(ListMSG))

	DbInit()

	http.Handle("/fachbereich/studiengang/semester/", m)
	//http.Handle("/fachbereich/studiengang/semester/group/", m)
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
		CREATE TABLE benutzer (
			fdNummer VARCHAR(256) PRIMARY KEY, 
			Vorname VARCHAR(256) NOT NULL, 
			Nachname VARCHAR(256) NULL,
			Age TINYINT NULL,
			Studiengang VARCHAR(256) NULL,
			Semester TINYINT NULL
			);

		CREATE TABLE nachrichten (
			NachrichtID INTEGER PRIMARY KEY AUTOINCREMENT,
			fdNummer VARCHAR(256) NOT NULL,
			GroupID INTEGER NOT NULL,
			message VARCHAR(256) NOT NULL,
			gesendeteUhrzeit DEFAULT CURRENT_TIMESTAMP
			);

		CREATE TABLE chatgroup (
			GroupID INTEGER PRIMARY KEY,
			GroupName VARCHAR(256) NOT NULL
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

func newAcc(fdNummer string, firstName string, lastName string, age string, studiengang string, semester string) {
	stmt, err := mainDB.Prepare("INSERT INTO benutzer(fdNummer, Vorname, Nachname, Age, Studiengang, Semester) values (?, ?, ?, ?, ?, ?)")
	checkErr(err)

	result, errExec := stmt.Exec(fdNummer, firstName, lastName, age, studiengang, semester)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

//allAcc Datenbank für die Suche Auswählen
func allAcc(w http.ResponseWriter) {
	stmt, err := mainDB.Prepare("SELECT * FROM benutzer")
	checkErr(err)

	rows, errQuery := stmt.Query()
	checkErr(errQuery)

	pRallAcc(w, rows)
}

// insertMSG Die direkte Funktion um eine Nachricht in der DB zu speichern
func insertMSG(fdNummer string, GroupID string, message string) {

	fmt.Println("insertMSG wurde aufgerufen")
	stmt, err := mainDB.Prepare("INSERT INTO nachrichten(fdNummer, GroupID, message) values (?, ?, ?)")
	checkErr(err)

	result, errExec := stmt.Exec(fdNummer, GroupID, message)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

func insertGroup(Gruppenname string) {
	fmt.Println("insertGroup wurde aufgerufen")
	stmt, err := mainDB.Prepare("INSERT INTO chatgroup(GroupName) values (?)")
	checkErr(err)

	result, errExec := stmt.Exec(Gruppenname)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

//groupMSG alle gespeicherten Messages widergeben
func groupMSG(w http.ResponseWriter, r *http.Request) {
	//srch := r.URL.Query().Get(":gruMSG")
	stmt, err := mainDB.Prepare("SELECT DISTINCT u.Vorname, c.GroupName, n.message, n.gesendeteUhrzeit FROM nachrichten n, chatgroup c, benutzer u WHERE n.GroupID = c.GroupID AND n.GroupID = ? AND u.fdNummer = n.fdNummer ORDER BY n.gesendeteUhrzeit")
	checkErr(err)

	//rows, errQuery := stmt.Query(srch)
	rows, errQuery := stmt.Query()
	checkErr(errQuery)

	groupRows(w, rows)
}

//processRows Suche der Message Parameter
func processRows(w http.ResponseWriter, rows *sql.Rows) {
	var Vorname string
	var GroupName string
	var message string
	var gesUhrzeit string

	for rows.Next() {
		err := rows.Scan(&Vorname, &GroupName, &message, &gesUhrzeit)
		checkErr(err)

		fmt.Fprintf(w, "Name: %s\n, Gruppe: %s\n, Nachricht: %s\n, gesUhrzeit: %s\n",
			string(Vorname), string(GroupName), string(message), string(gesUhrzeit))
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

	// chat := Chat{
	// 	UserName: "Name",
	// 	Messages: []Message{},
	// }

	for rows.Next() {

		err := rows.Scan(&vorname, &message)
		checkErr(err)

		//chat.Messages = append(chat.Messages, Message{Vorname: vorname, Content: message})

		fmt.Fprintf(w, "Nachricht von: %s, Nachricht: %s\n", string(vorname), string(message))
	}

	// parsedTemplate, _ := template.ParseFiles("templates/chat1.html")
	// err := parsedTemplate.Execute(w, chat)
	// if err != nil {
	// 	log.Println("Fehler beim Ausführen der Template :", err)
	// 	return
	// }

}

//checkErr
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
