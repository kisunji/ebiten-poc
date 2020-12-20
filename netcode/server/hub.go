package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

const (
	maxClients = 8
)

var (
	ErrNoMoreClientSlots = errors.New("no more client slots")
)

type Hub struct {
	addr string

	clientSlots []bool

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// Create new chat hub.
func NewServer(pattern string) *Hub {
	return &Hub{
		addr:        ":8080",
		clientSlots: make([]bool, maxClients),
		broadcast:   make(chan Message),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
	}
}

func (h *Hub) Listen() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		h.serveWs(w, r)
	})
	err := http.ListenAndServe(h.addr, nil)
	if err != nil {
		log.Fatalf("error listening and serving: %v", err)
	}
}

func (h *Hub) ListenTLS(sslCert string, sslKey string) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		h.serveWs(w, r)
	})
	err := http.ListenAndServeTLS(h.addr, sslCert, sslKey, nil)
	if err != nil {
		log.Fatalf("error listening and serving: %v", err)
	}
}

func (h *Hub) ChRegister() chan *Client { return h.register }

func (h *Hub) ChUnregister() chan *Client { return h.unregister }

func (h *Hub) ChBroadcast() chan Message { return h.broadcast }

func (h *Hub) GetMaxClients() int { return len(h.clientSlots) }

func (h *Hub) GetClients() map[*Client]bool { return h.clients }

func (h *Hub) HasClient(c *Client) bool {
	_, ok := h.clients[c]
	return ok
}

func (h *Hub) RegisterClient(c *Client, data interface{}) {
	c.data = data
	h.clients[c] = true
}

func (h *Hub) RemoveClient(c *Client) bool {
	if _, ok := h.clients[c]; ok {
		h.clientSlots[c.clientSlot] = false
		close(c.send)
		delete(h.clients, c)
		return true
	}
	return false
}

// serveWs handles websocket requests from the peer.
func (h *Hub) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	clientSlot, err := h.getNextFreeClientSlot()
	if err != nil {
		http.Error(w, "client slots full", 99)
		return
	}
	client := &Client{
		hub:        h,
		conn:       conn,
		clientSlot: clientSlot,
		send:       make(chan []byte, 256),
	}
	h.clientSlots[clientSlot] = true
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()

	fmt.Println("client connected")
}

func (h *Hub) getNextFreeClientSlot() (int, error) {
	maxClients := h.GetMaxClients()
	for i := 0; i < maxClients; i++ {
		if !h.clientSlots[i] {
			return i, nil
		}
	}
	return 0, ErrNoMoreClientSlots
}
