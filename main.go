package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/wat", WatHandler)
	http.ListenAndServe(":8787", nil)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OHAI"))
}

func WatHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("WAT"))
}
