package main

import (

       "io"

        "net/http"

        "github.com/bmizerany/pat"

        "log"

)

// hello world, the web server

func HelloServer(w http.ResponseWriter, req *http.Request) {

     io.WriteString(w, "Guten Tag, "+req.URL.Query().Get(":name")+"!\n")

}

// Nachricht an Server senden

func NachrichtSenden(w http.ResponseWriter, req *http.Request) {

     io.WriteString(w, "Nachricht ist: " + req.URL.Query().Get(":message")+"\n")

     io.WriteString(w, "Uhrzeit: "+req.URL.Query().Get(":uhrzeit")+"!\n")

     io.WriteString(w, "Absender: "+req.URL.Query().Get(":absender")+"!\n")

}

func NeueNachricht(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Test" + req.URL.Query().Get(":message"))

	i1,_ := mainDB.Prepare("INSERT INTO nachrichten (message) values (?)")
	result, _ := i2.Exec(req.URL.Query().Get(":message"))
	io.WriteString(w, result)
}


func main() {

     m := pat.New() // öffentlich schreibt man Gross

     m.Get("/api/v1/hello/:name", http.HandlerFunc(HelloServer))

     // zweite route

     // eine nachricht an den chat serversenden

     // 

     m.Get("/api/v2/nachricht/:message/:uhrzeit/:absender", http.HandlerFunc(NachrichtSenden))

     m.Get("/api/v2/nachricht/:nachricht_id/:fd_Nummer/:gesendet_uhrzeit", http.HandlerFunc(NeueNachricht))

     m.Get("/api/v2/chatgroup/:anz_user/:chat_id/:anz_nachrichten", http.HandlerFunc(NeueNachricht))

     m.Get("/api/v2/user/:fd_nummer/:status/:name/:alter/:fb/:studiengang/:semester", http.HandlerFunc(NeueNachricht))

     // Register this pat with the default serve mux so that other packages

     // may also be exported. (i.e. /debug/pprof/*)

     sqlStmt := `CREATE TABLE nachrichten (id INTEGER PRIMARY KEY AUTOINCREMENT, message VARCHAR(256) null)`
     http.Handle("/api/", m)

     db, err := sql.Open("sqlite3", "nachrichten.db")
     mainDB = db

     if err != nil {
	     log.Fatal("cannot open database", err)
     }

     

     err := http.ListenAndServe(":80", nil)

     if err != nil {

         log.Fatal("ListenAndServe: ", err)

    }

}
