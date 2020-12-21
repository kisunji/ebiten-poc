package main

import (
	"flag"
	"log"
	"net/http"

	"ebiten-poc/netcode"
)

var port = flag.String("port", ":8080", "http service address")

func main() {
	flag.Parse()
	hub := netcode.NewHub(*port)
	go hub.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	})
	log.Printf("listening on port %s\n", *port)
	err := http.ListenAndServe(*port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
