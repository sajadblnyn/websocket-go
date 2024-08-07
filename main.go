package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	SetupApi()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func SetupApi() {
	ctx := context.Background()
	manager := NewManager(ctx)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.ServeWs)
	http.HandleFunc("/login", manager.LoginHandler)

}
