package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/kisunji/ebiten-poc/game"
	"github.com/kisunji/ebiten-poc/server"
)

var port = flag.String("port", ":8080", "http service address")

func main() {
	flag.Parse()

	go server.Run()

	hub := server.NewHub()
	go hub.Run()

	for i := game.MaxClients; i < game.MaxChars; i++ {
		id := i
		go server.RunAI(server.NewAI(int32(id)), hub)
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	})
	log.Printf("listening on port %s\n", *port)
	err := http.ListenAndServe(*port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
