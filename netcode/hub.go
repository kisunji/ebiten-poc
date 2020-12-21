package netcode

import (
	"errors"
	"log"
	"math/rand"
	"net/http"

	"ebiten-poc/game"
	"ebiten-poc/pb"
	"google.golang.org/protobuf/proto"
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
	clients map[*Client]struct{}

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// Create new chat hub.
func NewHub(port string) *Hub {
	return &Hub{
		addr:        port,
		clientSlots: make([]bool, maxClients),
		broadcast:   make(chan Message),
		Register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clients[client] = struct{}{}
			h.clientSlots[client.clientSlot] = true
			char := &Char{
				fx:               1,
				fy:               1,
				px:               float64(game.ScreenPadding + rand.Intn(game.ScreenWidth-game.ScreenPadding*3)),
				py:               float64(game.ScreenPadding + rand.Intn(game.ScreenHeight-game.ScreenPadding*3)),
				speed:            1,
				clockOffset:      rand.Intn(10),
			}
			chars = append(chars, char)

			msg := pb.ConnectResponse{
				ClientSlot: client.clientSlot,
			}
			data, err := proto.Marshal(&msg)
			if err != nil {
				log.Fatal("client connect: marshaling error: ", err)
			}
			// Send to connecting player their information
			packetData := make([]byte, 1, len(data)+1)
			packetData[0] = byte(pb.MsgConnectResponse)
			packetData = append(packetData, data...)
			client.SendMessage(packetData)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.clientSlots[client.clientSlot] = false
				close(client.Send)
				delete(h.clients, client)
			}
		case message := <-h.broadcast:
			log.Println("broadcasting")
			for client := range h.clients {
				select {
				case client.Send <- message.data:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) GetMaxClients() int { return len(h.clientSlots) }

// serveWs handles websocket requests from the peer.
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
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
	log.Printf("slot %d connected\n", clientSlot)
	client := &Client{
		Hub:        h,
		Conn:       conn,
		clientSlot: clientSlot,
		Send:       make(chan []byte, 256),
	}
	h.clientSlots[clientSlot] = true
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()

	log.Println("client connected")
}

func (h *Hub) getNextFreeClientSlot() (int32, error) {
	maxClients := h.GetMaxClients()
	for i := 0; i < maxClients; i++ {
		if !h.clientSlots[i] {
			return int32(i), nil
		}
	}
	return 0, ErrNoMoreClientSlots
}
