package main

import (
    	"io"
    	"net/http"
	"github.com/bmizerany/pat"
	"log"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, " + req.URL.Query().Get(":name")+"!\n")
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
