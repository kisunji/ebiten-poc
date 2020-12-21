package server

import (
	"errors"
	"log"
	"net/http"

	"ebiten-poc/game"
	"ebiten-poc/pb"
	"google.golang.org/protobuf/proto"
)

var (
	ErrNoMoreClientSlots = errors.New("no more client slots")
)

type Hub struct {
	clientSlots []bool

	// Registered clients.
	clients map[*Client]*game.Char

	// Inbound messages from the clients.
	clientData chan Message

	// register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	AIChan chan AIData
}

// Create new chat hub.
func NewHub() *Hub {
	return &Hub{
		clientSlots: make([]bool, game.MaxClients),
		clientData:  make(chan Message),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]*game.Char),
		AIChan:      make(chan AIData),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			char := game.NewChar()
			chars[client.clientSlot] = char
			h.clients[client] = char
			h.clientSlots[client.clientSlot] = true

			msg := pb.ConnectResponse{
				ClientSlot: client.clientSlot,
				Px:         char.Px,
				Py:         char.Py,
			}
			data, err := proto.Marshal(&msg)
			if err != nil {
				log.Println("client connect: marshaling error: ", err)
				h.disconnect(client)
				continue
			}
			// Send to connecting player their information
			client.Send <- pb.AddHeader(data, pb.MsgConnectResponse)
			updateAll := &pb.UpdateAll{}
			for i, char := range chars {
				if char == nil {
					continue
				}
				ue := &pb.UpdateEntity{
					Index: int32(i),
					Fx:    int32(char.Fx),
					Fy:    int32(char.Fy),
					Vx:    int32(char.Vx),
					Vy:    int32(char.Vy),
					Px:    char.Px,
					Py:    char.Py,
					Speed: int32(char.Speed),
				}
				updateAll.Updates = append(updateAll.Updates, ue)
			}
			client.Send <- pb.AddHeader(data, pb.MsgUpdateAll)
		case client := <-h.unregister:
			h.disconnect(client)
		case msg := <-h.clientData:
			log.Printf("received msg from slot %d", msg.client.clientSlot)
			kind := pb.Kind(msg.data[0])
			buf := msg.data[1:]
			switch kind {
			case pb.MsgPlayerInput:
				pi := &pb.Input{}
				err := proto.Unmarshal(buf, pi)
				if err != nil {
					log.Println(err)
				}

				char := chars[msg.client.clientSlot]
				char.ProcessInput(pi)

				ue := &pb.UpdateEntity{
					Index: msg.client.clientSlot,
					Fx:    int32(char.Fx),
					Fy:    int32(char.Fy),
					Vx:    int32(char.Vx),
					Vy:    int32(char.Vy),
					Px:    char.Px,
					Py:    char.Py,
					Speed: int32(char.Speed),
				}
				data, err := proto.Marshal(ue)
				if err != nil {
					log.Println("client connect: marshaling error: ", err)
					continue
				}

				for c := range h.clients {
					c.Send <- pb.AddHeader(data, pb.MsgUpdateEntity)
				}
			default:
				h.disconnect(msg.client)
			}
		case aiInput := <-h.AIChan:
			char := chars[aiInput.Id]
			input := &pb.Input{
				UpPressed:    aiInput.UpPressed,
				DownPressed:  aiInput.DownPressed,
				LeftPressed:  aiInput.LeftPressed,
				RightPressed: aiInput.RightPressed,
			}
			if char == nil {
				char = &game.Char{}
				chars[aiInput.Id] = char
			}
			char.ProcessInput(input)
			ue := &pb.UpdateEntity{
				Index: aiInput.Id,
				Fx:    int32(char.Fx),
				Fy:    int32(char.Fy),
				Vx:    int32(char.Vx),
				Vy:    int32(char.Vy),
				Px:    char.Px,
				Py:    char.Py,
				Speed: int32(char.Speed),
			}
			data, err := proto.Marshal(ue)
			if err != nil {
				log.Println("client connect: marshaling error: ", err)
				continue
			}

			for c := range h.clients {
				c.Send <- pb.AddHeader(data, pb.MsgUpdateEntity)
			}
		}
	}
}

func (h *Hub) disconnect(client *Client) {
	if _, ok := h.clients[client]; ok {
		h.clientSlots[client.clientSlot] = false
		client.Send <- []byte{byte(pb.MsgDisconnectPlayer)}
		close(client.Send)
		delete(h.clients, client)
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
	client.Hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
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
