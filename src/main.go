package main

import (
	"io"
	"log"
	"net/http"

	"github.com/bmizerany/pat"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Hallo mein kleiner Freund, "+req.URL.Query().Get(":name")+"!\n")
}

func main() {
	m := pat.New()

	m.Get("/hello/:name", http.HandlerFunc(HelloServer))

	http.Handle("/", m)
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
