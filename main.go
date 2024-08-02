package main

import (
	"log"
	"net/http"
)

func main() {
	SetupApi()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func SetupApi() {
	manager := NewManager()
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.ServeWs)
}
