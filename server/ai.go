package server

import (
	"log"
	"math"
	"math/rand"
	"time"

	"ebiten-poc/game"
)

type AIData struct {
	Id           int32
	UpPressed    bool
	DownPressed  bool
	LeftPressed  bool
	RightPressed bool
}

func NewAI(id int32) *AI {
	char := game.NewChar()
	chars[id] = char
	ai := &AI{
		Char:    char,
		id:      id,
		running: false,
	}
	return ai
}

type AI struct {
	*game.Char

	id      int32
	running bool
}

func RunAI(ai *AI, hub *Hub) {
	log.Printf("running ai %d", ai.id)
	ai.running = true
	rand.Seed(int64(time.Now().Nanosecond()))
	for {
		if !ai.running {
			log.Printf("killed ai %d", ai.id)
			break
		}
		t := rand.Intn(10000)
		time.Sleep(time.Duration(t) * time.Millisecond)
		hub.AIChan <- computeMovement(ai)
	}
}

func computeMovement(ai *AI) AIData {
	biasx := ai.Px/float64(game.ScreenWidth) - .5
	biasy := ai.Py/float64(game.ScreenHeight) - .5
	fx := 0
	if rawx := math.Round(rand.NormFloat64() - biasx); rawx < 0 {
		fx = -1
	} else if rawx > 0 {
		fx = 1
	}
	fy := 0
	if rawy := math.Round(rand.NormFloat64() - biasy); rawy < 0 {
		fy = -1
	} else if rawy > 0 {
		fy = 1
	}

	return AIData{
		Id:           ai.id,
		UpPressed:    fy == -1,
		DownPressed:  fy == 1,
		LeftPressed:  fx == -1,
		RightPressed: fx == 1,
	}
}
