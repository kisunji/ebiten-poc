package server

import (
	"log"
	"time"

	"github.com/kisunji/ebiten-poc/common"
)

const (
	updateFrequency = 1 * time.Second / 60
)

type Updater struct {
	world *World
}

func NewUpdater() *Updater {
	return &Updater{
		world: &World{
			Running:     false,
			Chars:       make(common.Chars, common.MaxChars),
			tick:        0,
			PlayerSlots: make([]bool, common.MaxClients),
			AIs:         make([]*AI,0),
		},
	}
}

type State int

type World struct {
	Running     bool
	Chars       common.Chars
	tick        int64
	PlayerSlots []bool
	HostSlot    int32
	AIs         []*AI
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
	for _, char := range w.Chars {
		if char == nil {
			continue
		}
		char.Move()
	}
}
