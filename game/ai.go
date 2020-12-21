package game

import (
	"log"
	"math"
	"math/rand"
	"time"
)

var AIId int

func newAI(speed int) *AI {
	AIId++
	ai := &AI{
		Runner: Runner{
			speed:       speed,
			px:          float64(ScreenPadding + rand.Intn(ScreenWidth-ScreenPadding*3)),
			py:          float64(ScreenPadding + rand.Intn(ScreenHeight-ScreenPadding*3)),
			clockOffset: rand.Intn(10),
			sprite:      runnerWaitingFrame,
		},
		id:      AIId,
		moveCmd: nil,
		running: false,
	}
	return ai
}

type AI struct {
	Runner

	id      int
	moveCmd *movement
	running bool
}

type movement struct {
	fx, fy int
	repeat int
}

func Run(ai *AI) {
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
		if ai.moveCmd == nil {
			ai.moveCmd = computeMovement(ai.px, ai.py)
		}
	}
}

func computeMovement(px, py float64) *movement {
	biasx := px/float64(ScreenWidth) - .5
	biasy := py/float64(ScreenHeight) - .5
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

	if fx == 0 && fy == 0 {
		return nil
	}

	return &movement{
		fx:     fx,
		fy:     fy,
		repeat: rand.Intn(200),
	}
}

func (a *AI) Move() {
	if a.moveCmd == nil {
		a.vx = 0
		a.vy = 0
		a.sprite = runnerWaitingFrame
		return
	}
	defer func() {
		a.moveCmd.repeat--
		if a.moveCmd.repeat <= 0 {
			a.moveCmd = nil
		}
	}()

	a.sprite = runnerWalkingFrame

	a.fx = a.moveCmd.fx
	a.vx = a.fx * a.speed

	a.fy = a.moveCmd.fy
	a.vy = a.fy * a.speed

	normalized := math.Sqrt(math.Pow(float64(a.vx), 2) + math.Pow(float64(a.vy), 2))
	a.px += float64(a.vx) / normalized
	if a.px >= ScreenWidth-ScreenPadding {
		a.px = ScreenWidth - ScreenPadding - 1
	}
	if a.px <= ScreenPadding {
		a.px = ScreenPadding + 1
	}
	a.py += float64(a.vy) / normalized
	if a.py >= ScreenHeight-ScreenPadding-10 {
		a.py = ScreenHeight - ScreenPadding - 11
	}
	if a.py <= ScreenPadding {
		a.py = ScreenPadding + 1
	}
}
