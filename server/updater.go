package server

import (
	"log"
	"math"
	"time"

	"github.com/kisunji/ebiten-poc/common"
	"github.com/kisunji/ebiten-poc/pb"
	"google.golang.org/protobuf/proto"
)

const (
	updateFrequency = 1 * time.Second / 60
)

func NewWorld(broadcast chan []byte) *World {
	return &World{
		Running:     false,
		Chars:       make(common.Chars, common.MaxChars),
		tick:        0,
		PlayerSlots: make([]bool, common.MaxClients),
		AIs:         make([]*AI, 0),
		broadcast:   broadcast,
	}
}

type World struct {
	Running     bool
	Chars       common.Chars
	tick        int64
	PlayerSlots []bool
	HostSlot    int32
	AIs         []*AI
	broadcast   chan []byte
}

func (w *World) Setup(aiChan chan AIData) {
	for i := 0; i < common.MaxChars; i++ {
		if i < common.MaxClients && w.PlayerSlots[i] {
			w.Chars[i] = common.NewChar()
		} else {
			char := common.NewChar()
			w.Chars[i] = char
			ai := &AI{
				Char: char,
				id:   int32(i),
			}
			w.AIs = append(w.AIs, ai)
			go w.RunAI(ai, aiChan)
		}
	}
}

// Run should be called in a goroutine
func (w *World) Run() {
	w.Running = true
	previous := time.Now()
	var lag time.Duration

	for w.Running {
		current := time.Now()
		elapsed := current.Sub(previous)
		previous = current
		lag += elapsed

		for lag >= updateFrequency {
			w.update()
			lag -= updateFrequency
		}
	}
	log.Println("stopping world")
}

func (w *World) update() {
	for i, char := range w.Chars {
		if char == nil || char.IsDead {
			continue
		}
		if char.Attacking() {
			char.Attack()
			// reached end of animation
			if !char.Attacking() {
				const radius = 12.0
				x0, y0 := char.ImpactSite(radius)
				for j, isPlayer := range w.PlayerSlots {
					if !isPlayer || i == j {
						continue
					}
					target := w.Chars[j]
					if target.IsDead {
						continue
					}
					x1 := target.Px
					y1 := target.Py
					dx := x1 - x0
					dy := y1 - y0
					distance := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))
					if distance <= radius {
						target.IsDead = true
						ue := &pb.ServerMessage{
							Content: &pb.ServerMessage_UpdateEntity{
								UpdateEntity: &pb.UpdateEntity{
									Index:  int32(j),
									Fx:     int32(target.Fx),
									Fy:     int32(target.Fy),
									Px:     target.Px,
									Py:     target.Py,
									Speed:  int32(target.Speed),
									IsDead: true,
								},
							},
						}
						data, err := proto.Marshal(ue)
						if err != nil {
							log.Fatalln("client connect: marshaling error: ", err)
						}
						w.broadcast <- data
					}
				}
			}
		}
		char.Move()
	}
}
