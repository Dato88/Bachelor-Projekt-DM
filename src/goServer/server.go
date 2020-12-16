package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"

	"github.com/bmizerany/pat"
	_ "github.com/mattn/go-sqlite3"
)

var addr = ":8080"
var staticDir = "./static"

var redetDB = false

const dbFileName = "serverNachrichten.db"

var mainDB *sql.DB

//Zum testen => http://localhost:80/hello/andrej
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Server ist Online, "+req.URL.Query().Get(":name")+"!\n")
}

func main() {
	m := pat.New()
	m.Get("/hello/:name", http.HandlerFunc(HelloServer))

	http.Handle("/", m)

	//http.Handle("/", http.FileServer(http.Dir(staticDir)))

	log.Print("Running on Port: ", addr)

	err := http.ListenAndServe(addr, nil)

	if err != nil {
		log.Fatal("ListenAndServe Fehler: ", err)
	}
}
