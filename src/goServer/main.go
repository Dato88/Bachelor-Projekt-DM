package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/bmizerany/pat"
	_ "github.com/mattn/go-sqlite3"
)

//Message = einzelne Nachrichten
type Message struct {
	Content string
}

//Chat = Struct mit dem UserName und Messages Array
type Chat struct {
	UserName string
	Messages []Message
}

// Port 80 und Verzeichnis mit den HTML Dateien
var addr = ":80"
var staticDir = "./static"

// Rücksetzen der Datenbank ?
var resetDB = true

const dbFileName = "nachrichten.db"

var mainDB *sql.DB

//HelloServer = REST API HANDLER
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, "+req.URL.Query().Get(":name")+"!\n")
}

//AddMessage =
func AddMessage(w http.ResponseWriter, req *http.Request) {
	insertMessage(req.URL.Query().Get(":message"))
	io.WriteString(w, "Nachricht hinzugefügt: "+req.URL.Query().Get(":message")+"!\n")
}

//ListAllMessages =
func ListAllMessages(w http.ResponseWriter, req *http.Request) {
	allMessages(w)
}

//ListMessage =
func ListMessage(w http.ResponseWriter, req *http.Request) {
	var msgid = req.URL.Query().Get(":msgid")
	idMessage(w, msgid)
}

func main() {
	m := pat.New()
	//HTTP Handler für die Dynamishen Seiten
	m.Get("/api/v1/hello/:name", http.HandlerFunc(HelloServer))
	m.Get("/api/v1/message/add/:message", http.HandlerFunc(AddMessage))
	m.Get("/api/v1/message/list/id=:msgid", http.HandlerFunc(ListMessage))
	m.Get("/api/v1/message/list/all", http.HandlerFunc(ListAllMessages))

	// Datenbank initialiieren
	Initialize()

	// starte die Dienste (REST und statischer Webserver)
	http.Handle("/api/v1/", m)
	http.Handle("/", http.FileServer(http.Dir(staticDir)))

	log.Print(" Running on ", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

//Initialize = Datenbank initialisieren
func Initialize() {
	sqlStmt := `
		CREATE TABLE nachrichten (id INTEGER PRIMARY KEY AUTOINCREMENT,
		message VARCHAR(256) NULL);	
	`
	Create(sqlStmt)
}

//Create =
func Create(sqlStmt string) {
	if resetDB {
		fmt.Println("Datenbank wird gelöscht")
		os.Remove(dbFileName)
	}

	db, err := sql.Open("sqlite3", dbFileName)
	checkErr(err)
	mainDB = db

	if resetDB {
		fmt.Println("Tabelle erzeugt")
		_, err = db.Exec(sqlStmt)
		checkErr(err)
	}
}

//insertMessage = Nachrichten in die Datenbank eintragen
func insertMessage(message string) {
	stmt, err := mainDB.Prepare("INSERT INTO nachrichten(message) values (?)")
	checkErr(err)

	result, errExec := stmt.Exec(message)
	checkErr(errExec)

	newID, _ := result.LastInsertId()
	fmt.Println(newID)
}

//allMessages =
func allMessages(w http.ResponseWriter) {
	//alle Nachrichten holen und in der Variable err Speichern
	stmt, err := mainDB.Prepare(" SELECT * FROM nachrichten")
	//Checken ob ein Fehler aufgetreten ist
	checkErr(err)

	//mit stmt.Query() setzt man den Query ab
	rows, errQuery := stmt.Query() //Zeilen enthalten id und message
	checkErr(errQuery)

	//rows enthält eine ganze Zeile von Nachrichten die in processRows verabeitet werden
	processRows(w, rows)
}

//idMessage =
func idMessage(w http.ResponseWriter, msgid string) {

	stmt, err := mainDB.Prepare(" SELECT * FROM nachrichten WHERE id = ?")
	checkErr(err)

	rows, errQuery := stmt.Query(msgid) //Zeilen enthalten id und message
	checkErr(errQuery)

	processRows(w, rows)
}

//processRows wird von allMessages Aufgerufen
func processRows(w http.ResponseWriter, rows *sql.Rows) {
	var ID int64
	var message string

	//ich bereite meinen chat hier vor mit einem Vorgegebenen UserNamen und mache eine Leere Liste
	//{} sagt dass ich eine Neue Leere Liste will
	chat := Chat{
		UserName: "Jan-Torsten",
		Messages: []Message{},
	}

	for rows.Next() { //Durchlaufen mit Cursor (Datenbankzeiger)
		//& ist ein Adressoperator
		//mit dem Scan scanne ich die ID und die message ab und übertrage diese in die oberen Variablen ID und message
		err := rows.Scan(&ID, &message)
		checkErr(err)

		//F steht für File
		//Fprintf bekommt einen Writer und einen Formatstring
		//fmt.Fprintf(w, "ID: %d, message: %s\n", ID, string(message))

		//mit der Funktion append hänge ich chat die passende Message an
		//ich initialisiere mit Message{} den Content, also übertrage ich den Inhalt der Reihe in ein neues Messageobjekt
		chat.Messages = append(chat.Messages, Message{Content: message})
	}

	//parsedTemplate: damit hole ich mir aus dem Ordner templates die HTML
	parsedTemplate, _ := template.ParseFiles("templates/index.html")
	//hiermit übergebe ich der template die Datenstruktur chat
	err := parsedTemplate.Execute(w, chat)

	//übliche Fehlersuche
	if err != nil {
		log.Println("Error executing template :", err)
		return
	}
}

// Hilfsfunktion

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
