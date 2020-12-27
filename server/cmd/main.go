package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/kisunji/ebiten-poc/server"
)

var port = flag.String("port", ":8080", "http service address")
var insecure = flag.Bool("insecure", false, "listen over insecure ws")
var cert = flag.String("cert", "", "path to cert file")
var key = flag.String("key", "", "path to key file")

func main() {
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	hub := server.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	})
	log.Printf("listening on port %s\n", *port)
	if *insecure {
		err := http.ListenAndServe(*port, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	} else {
		if *cert == "" || *key == "" {
			log.Fatal("must provide --cert and --key")
		}
		err := http.ListenAndServeTLS(*port, *cert, *key, nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	}
}
