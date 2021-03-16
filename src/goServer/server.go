package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bmizerany/pat"
	_ "github.com/mattn/go-sqlite3"
)

var addr = ":80"
var staticDir = "./templates"
var resetDB = false

const dbFile = "serverNachrichten.db"

var mainDB *sql.DB

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
	// Gruppenname      string
	// GesendeteUhrzeit string
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

//HelloServer URL zum testen => localhost:80/fachbereich/studiengang/semester/group1/hello/andrej
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Server ist Online, "+req.URL.Query().Get(":name")+"!\n")
}

//CreateAcc erstellt einen neuen Account mit den nötigen Parameter
func CreateAcc(w http.ResponseWriter, req *http.Request) {
	//fmt.Println("CreateAcc wurde aufgerufen")
	newAcc(req.URL.Query().Get(":fdNummer"),
		req.URL.Query().Get(":firstName"),
		req.URL.Query().Get(":lastName"),
		req.URL.Query().Get(":age"),
		req.URL.Query().Get(":studiengang"),
		req.URL.Query().Get(":semester"))
	//io.WriteString(w, req.URL.Query().Get(":firstName")+" hat einen Account erstellt!"+"\n")
}

//FindAcc alle erstellten Accounts Anzeigen lassen mit allen gespeicherten Informationen
func FindAcc(w http.ResponseWriter, req *http.Request) {
	//io.WriteString(w, "Alle Accounts Anzeigen!"+"\n")
	allAcc(w)
}

//AddGroupMSG eine Nachricht für die Gruppe in der DB zu Speichern
func AddGroupMSG(w http.ResponseWriter, req *http.Request) {
	//fmt.Println("AddGroupMSG wurde aufgerufen")
	insertMSG(req.URL.Query().Get(":fdNummer"),
		req.URL.Query().Get(":groupID"),
		req.URL.Query().Get(":message"))
	//io.WriteString(w, "Nachricht eingesetzt: "+req.URL.Query().Get(":message")+"\n")
}

//AddGroup Gruppe hinzufügen
func AddGroup(w http.ResponseWriter, req *http.Request) {
	//fmt.Println("Gruppe erstellt!")
	insertGroup(req.URL.Query().Get(":GroupName"))
	//io.WriteString(w, "Neue Gruppe erstellt: "+req.URL.Query().Get(":GroupName")+"\n")
}

//ListMSG alle gespeicherten Nachrichten einer Gruppe Anzeigen
func ListMSG(w http.ResponseWriter, req *http.Request) {
	var grID = req.URL.Query().Get(":gruMSG")
	groupMSG(w, grID)
}

func main() {
	m := pat.New()
	m.Get("/fachbereich/studiengang/semester/:groupID/hello/:name", http.HandlerFunc(HelloServer))

	//Ein Account und eine Gruppe muss erstellt werden um die Chatfunktionen testen zu können!!
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

	//http://bachelor-community.informatik.hs-fulda.de/search/1
	// m.Get("/search/:gruMSG", http.HandlerFunc(ListMSG))

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
		fmt.Println("\nTabelle wurde erstellt!")
		_, err = db.Exec(sqlStatement)
		checkErr(err)
	}
}

func newAcc(fdNummer string, firstName string, lastName string, age string, studiengang string, semester string) {
	stmt, err := mainDB.Prepare("INSERT INTO benutzer(fdNummer, Vorname, Nachname, Age, Studiengang, Semester) values (?, ?, ?, ?, ?, ?)")
	checkErr(err)

	result, errExec := stmt.Exec(fdNummer, firstName, lastName, age, studiengang, semester)
	checkErr(errExec)

	fmt.Println("Neuer Nutzer mit der Fd-Nummer: " + fdNummer + " wurde erstellt!!")
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

func insertGroup(Gruppenname string) {
	//fmt.Println("insertGroup wurde aufgerufen")
	stmt, err := mainDB.Prepare("INSERT INTO chatgroup(GroupName) values (?)")
	checkErr(err)

	result, errExec := stmt.Exec(Gruppenname)
	checkErr(errExec)

	fmt.Println("Gruppe " + Gruppenname + " wurde erstellt!!")
	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

// insertMSG Die direkte Funktion um eine Nachricht in der DB zu speichern
func insertMSG(fdNummer string, GroupID string, message string) {

	//fmt.Println("insertMSG wurde aufgerufen")
	stmt, err := mainDB.Prepare("INSERT INTO nachrichten(fdNummer, GroupID, message) values (?, ?, ?)")
	checkErr(err)

	result, errExec := stmt.Exec(fdNummer, GroupID, message)
	checkErr(errExec)

	fmt.Println("Message in " + GroupID + " wurde gespeichert!!")
	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

//groupMSG alle gespeicherten Messages einer Gruppe widergeben
func groupMSG(w http.ResponseWriter, grID string) {
	//Das Select Statement muss mindestens die selbe Anzahl an Argumenten haben wie sie bei groupRows gesucht werden!!!
	stmt, err := mainDB.Prepare("SELECT DISTINCT u.Vorname, c.GroupName, n.message, n.gesendeteUhrzeit FROM nachrichten n, chatgroup c, benutzer u WHERE n.GroupID = c.GroupID AND n.GroupID = ? AND u.fdNummer = n.fdNummer ORDER BY n.gesendeteUhrzeit")
	checkErr(err)

	rows, errQuery := stmt.Query(grID)
	checkErr(errQuery)
	groupRows(w, rows)
}

//processRows Suche der Message Parameter
func processRows(w http.ResponseWriter, rows *sql.Rows) {
	var vorname string
	var message string

	for rows.Next() {
		err := rows.Scan(&vorname, &message)
		checkErr(err)

		fmt.Fprintf(w, "Nachricht von: %s, Nachricht: %s\n", string(vorname), string(message))
	}

	fmt.Println("Nachrichten wurden widergegeben!!")
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
	var groupName string
	var message string
	var gesUhrzeit string

	chat := Chat{
		UserName: "Gruppenname",
		Messages: []Message{},
	}

	for rows.Next() {

		// err := rows.Scan(&vorname, &message)
		err := rows.Scan(&vorname, &groupName, &message, &gesUhrzeit)
		checkErr(err)

		chat.Messages = append(chat.Messages, Message{Vorname: vorname, Content: message})

		//das müsste vielleicht Auskommentiert werden
		// fmt.Fprintf(w, "Nachricht von: %s, Nachricht: %s\n", string(vorname), string(message))
		// fmt.Fprintf(w, "Name: %s\n, Gruppe: %s\n, Nachricht: %s\n, gesUhrzeit: %s\n",
		// 	string(vorname), string(groupName), string(message), string(gesUhrzeit))
	}

	// parsedTemplate, _ := template.ParseFiles("templates/chat1.html")
	parsedTemplate, _ := template.ParseFiles("http://bachelor-community.informatik.hs-fulda.de/templates/chat1.html")
	err := parsedTemplate.Execute(w, chat)
	if err != nil {
		log.Println("Fehler beim Ausführen der Template: ", err)
		return
	}

	fmt.Println("Nachrichten der Gruppe " + groupName + " wurden widergegeben!!")
}

//checkErr
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
