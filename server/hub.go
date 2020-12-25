package server

import (
	"log"
	"net/http"

	"github.com/kisunji/ebiten-poc/common"
	"github.com/kisunji/ebiten-poc/pb"
	"google.golang.org/protobuf/proto"
)

type Hub struct {
	// game engine
	updater *Updater
	// Registered clients.
	clients map[*Client]int32
	// Inbound messages from the clients.
	clientData chan clientData
	// register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
	// Inputs from AI.
	AIChan    chan AIData
	isRunning bool
}

// Create new chat hub.
func NewHub() *Hub {
	return &Hub{
		updater:    NewUpdater(),
		clients:    make(map[*Client]int32),
		clientData: make(chan clientData),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	h.isRunning = true
	for {
		select {
		case client := <-h.register:
			clientSlot := h.getNextFreeClientSlot()
			if clientSlot < 0 {
				resp := &pb.ServerMessage{
					Content: &pb.ServerMessage_ConnectError{
						ConnectError: &pb.ConnectError{
							Message: "all slots are full",
						},
					},
				}
				data, err := proto.Marshal(resp)
				if err != nil {
					log.Fatalln("client connect: marshaling error: ", err)
				}
				client.Send <- data
				return
			}
			// state check (running already?)
			if h.updater.world.Running {
				resp := &pb.ServerMessage{
					Content: &pb.ServerMessage_ConnectError{
						ConnectError: &pb.ConnectError{
							Message: "game is already running",
						},
					},
				}
				data, err := proto.Marshal(resp)
				if err != nil {
					log.Fatalln("client connect: marshaling error: ", err)
				}
				client.Send <- data
				return
			}
			log.Printf("player %d connected\n", clientSlot)
			client.clientSlot = clientSlot
			if len(h.clients) == 0 {
				h.updater.world.HostSlot = clientSlot
			}
			h.updater.world.PlayerSlots[clientSlot] = true
			h.clients[client] = clientSlot

			resp := &pb.ServerMessage{
				Content: &pb.ServerMessage_ConnectResponse{
					ConnectResponse: &pb.ConnectResponse{
						ClientSlot: client.clientSlot,
						IsHost:     h.updater.world.HostSlot == clientSlot,
					},
				},
			}
			data, err := proto.Marshal(resp)
			if err != nil {
				log.Fatalln("client connect: marshaling error: ", err)
			}
			client.Send <- data

			resp = &pb.ServerMessage{
				Content: &pb.ServerMessage_UpdateLobby{
					UpdateLobby: &pb.UpdateLobby{
						ConnectedSlots: h.updater.world.PlayerSlots,
						HostSlot:       h.updater.world.HostSlot,
					},
				},
			}
			h.sendToAll(resp)
		case client := <-h.unregister:
			h.disconnect(client)
		case clientMsg := <-h.clientData:
			msg := &pb.ClientMessage{}
			err := proto.Unmarshal(clientMsg.data, msg)
			if err != nil {
				log.Println("error unmarshalling from client")
			}
			switch buf := msg.Content.(type) {
			case *pb.ClientMessage_Input:
				char := h.updater.world.Chars[clientMsg.client.clientSlot]
				char.ProcessInput(buf.Input)

				resp := &pb.ServerMessage{
					Content: &pb.ServerMessage_UpdateEntity{
						UpdateEntity: &pb.UpdateEntity{
							Index: clientMsg.client.clientSlot,
							Fx:    int32(char.Fx),
							Fy:    int32(char.Fy),
							Vx:    int32(char.Vx),
							Vy:    int32(char.Vy),
							Px:    char.Px,
							Py:    char.Py,
							Speed: int32(char.Speed),
						},
					},
				}
				h.sendToAll(resp)
			case *pb.ClientMessage_StartGame:
				log.Println("starting!")
				if !h.updater.world.Running {
					h.updater.world.Setup(h.AIChan)
					go h.updater.world.Run()
				}
				msg := &pb.ServerMessage{
					Content: &pb.ServerMessage_GameStart{
						GameStart: &pb.GameStart{},
					},
				}
				h.sendToAll(msg)
			default:
				h.disconnect(clientMsg.client)
			}
		case aiInput := <-h.AIChan:
			log.Println("ai input got")
			char := h.updater.world.Chars[aiInput.Id]
			input := &pb.Input{
				UpPressed:    aiInput.UpPressed,
				DownPressed:  aiInput.DownPressed,
				LeftPressed:  aiInput.LeftPressed,
				RightPressed: aiInput.RightPressed,
			}
			if char == nil {
				char = common.NewChar()
				h.updater.world.Chars[aiInput.Id] = char
			}
			char.ProcessInput(input)
			resp := &pb.ServerMessage{
				Content: &pb.ServerMessage_UpdateEntity{
					UpdateEntity: &pb.UpdateEntity{
						Index: aiInput.Id,
						Fx:    int32(char.Fx),
						Fy:    int32(char.Fy),
						Vx:    int32(char.Vx),
						Vy:    int32(char.Vy),
						Px:    char.Px,
						Py:    char.Py,
						Speed: int32(char.Speed),
					},
				},
			}
			h.sendToAll(resp)
		}
	}
}

func (h *Hub) sendToAll(msg *pb.ServerMessage) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Println("client connect: marshaling error: ", err)
	}
	for c := range h.clients {
		c.Send <- data
	}
}

func (h *Hub) disconnect(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
	}
	if client.clientSlot >= 0 {
		h.updater.world.PlayerSlots[client.clientSlot] = false
		msg := &pb.ServerMessage{
			Content: &pb.ServerMessage_PlayerDisconnected{
				PlayerDisconnected: &pb.PlayerDisconnected{
					Id: client.clientSlot,
				},
			},
		}
		h.sendToAll(msg)
		log.Printf("player %d disconnected\n", client.clientSlot)
	}
	if h.updater.world.HostSlot == client.clientSlot {
		for i, p := range h.updater.world.PlayerSlots {
			if p {
				msg := &pb.ServerMessage{
					Content: &pb.ServerMessage_NewHost{
						NewHost: &pb.NewHost{
							Id: int32(i),
						},
					},
				}
				h.sendToAll(msg)
				log.Printf("%d is new host\n", i)
			}
		}
	}
	if h.updater.world.Running && len(h.clients) == 0 {
		log.Println("no clients found")
		h.updater.world.Running = false
		for _, ai := range h.updater.world.AIs {
			ai.stop = true
		}
		h.updater = NewUpdater()
	}
}

// serveWs handles websocket requests from the peer.
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	if !h.isRunning {
		log.Fatal("hub is not running")
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		Hub:        h,
		Conn:       conn,
		clientSlot: -1,
		Send:       make(chan []byte, 256),
	}
	client.Hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}

func (h *Hub) getNextFreeClientSlot() int32 {
	for i := 0; i < common.MaxClients; i++ {
		if !h.updater.world.PlayerSlots[i] {
			return int32(i)
		}
	}
	return -1
}
