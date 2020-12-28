package server

import (
	"log"
	"math"
	"math/rand"
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
		Score:       make([]int32, common.MaxClients),
		AIs:         make([]*AI, 0),
		broadcast:   broadcast,
		killSig:     make(chan struct{}),
	}
}

type World struct {
	Running     bool
	Chars       common.Chars
	Coins       []*common.Coin
	Score       []int32
	tick        int64
	PlayerSlots []bool
	HostSlot    int32
	AIs         []*AI
	broadcast   chan []byte
	killSig     chan struct{}
}

func (w *World) Setup(aiChan chan AIData) {
	for i := 0; i < common.MaxChars; i++ {
		if i < common.MaxClients && w.PlayerSlots[i] {
			w.Chars[i] = common.NewChar()
		} else {
			char := common.NewChar()
			w.Chars[i] = char
			ai := &AI{
				Char:    char,
				id:      int32(i),
				killSig: make(chan struct{}),
			}
			w.AIs = append(w.AIs, ai)
			go w.RunAI(ai, aiChan)
		}
	}
	go w.makeCoins()
}

func (w *World) makeCoins() {
	timer := time.NewTimer(time.Duration(rand.Intn(5000)) * time.Millisecond)
	defer func() {
		log.Println("stopping makeCoins")
		timer.Stop()
	}()
	for {
		select {
		case <-timer.C:
			coin := common.NewCoin()
			w.Coins = append(w.Coins, coin)
			msg := &pb.ServerMessage{
				Content: &pb.ServerMessage_NewCoin{
					NewCoin: &pb.NewCoin{
						Index:       int32(len(w.Coins) - 1),
						Px:          coin.Px,
						Py:          coin.Py,
						FrameOffset: int32(coin.FrameOffset),
					},
				},
			}
			bytes, err := proto.Marshal(msg)
			if err != nil {
				log.Fatal(err)
			}
			w.broadcast <- bytes
			timer.Reset(time.Duration(rand.Intn(5000)) * time.Millisecond)
		case <-w.killSig:
			return
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
	for _, ai := range w.AIs {
		ai.killSig <- struct{}{}
	}
	w.killSig <- struct{}{}
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
				x0, y0 := char.ImpactSite(common.HitRadius)
				for j, isPlayer := range w.PlayerSlots {
					if !isPlayer || i == j {
						continue
					}
					target := w.Chars[j]
					if target.IsDead {
						continue
					}
					if isHit(x0, y0, target.Px, target.Py, common.HitRadius) {
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
						bytes, err := proto.Marshal(ue)
						if err != nil {
							log.Fatalln("client connect: marshaling error: ", err)
						}
						w.broadcast <- bytes
					}
				}
			}
		}
		char.Move()
	}
	for i, coin := range w.Coins {
		if coin.PickedUp {
			continue
		}
		for j, isPlayer := range w.PlayerSlots {
			if !isPlayer {
				continue
			}
			target := w.Chars[j]
			if target.IsDead {
				continue
			}
			if isHit(target.Px, target.Py, coin.Px, coin.Py, coin.PickupRadius) {
				w.Score[j]++
				w.Coins[i].PickedUp = true
				ue := &pb.ServerMessage{
					Content: &pb.ServerMessage_CoinGot{
						CoinGot: &pb.CoinGot{Index: int32(i)},
					},
				}
				bytes, err := proto.Marshal(ue)
				if err != nil {
					log.Fatalln("client connect: marshaling error: ", err)
				}
				w.broadcast <- bytes
			}
		}
	}
	var alive []int
	for j, isPlayer := range w.PlayerSlots {
		if !isPlayer {
			continue
		}
		target := w.Chars[j]
		if !target.IsDead {
			alive = append(alive, j)
		}
	}
	if len(alive) == 1 {
		ge := &pb.ServerMessage{
			Content: &pb.ServerMessage_GameEnd{
				GameEnd: &pb.GameEnd{
					Survivor: int32(alive[0]),
					Score:    w.Score,
				},
			},
		}
		bytes, err := proto.Marshal(ge)
		if err != nil {
			log.Fatalln("client connect: marshaling error: ", err)
		}
		w.broadcast <- bytes
		w.Running = false
	}
}

func isHit(x0, y0, x1, y1, radius float64) bool {
	dx := x1 - x0
	dy := y1 - y0
	distance := math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2))
	if distance <= radius {
		return true
	}
	return false
}
