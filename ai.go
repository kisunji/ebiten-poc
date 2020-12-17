package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"math"
	"math/rand"
	"time"
)

type AI struct {
	Runner

	id int
	killSig chan struct{}
	moveCmd *movement
	running bool
}

const (
	Up direction = iota
	Down
	Left
	Right
	UpRight
	UpLeft
	DownRight
	DownLeft
)

type direction int

type movement struct {
	d     direction
	units int
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
		if ai.moveCmd == nil {
			t := rand.Intn(5000)
			time.Sleep(time.Duration(t) * time.Millisecond)
			ai.moveCmd = &movement{
				d: direction(rand.Intn(8)),
				units: rand.Intn(200),
			}
		}
	}
}

func (a *AI) Move() {
	if a.moveCmd == nil {
		a.vx = 0
		a.vy = 0
		return
	}
	switch a.moveCmd.d {
	case Up:
		a.fy = -1
		a.vy = -1
	case Down:
		a.fy = 1
		a.vy = 1
	case Left:
		a.fx = -1
		a.vx = -1
	case Right:
		a.fx = 1
		a.vx = 1
	case UpLeft:
		a.fx = -1
		a.fy = -1
		a.vx = -1
		a.vy = -1
	case UpRight:
		a.fx = 1
		a.fy = -1
		a.vx = 1
		a.vy = -1
	case DownLeft:
		a.fx = -1
		a.fy = 1
		a.vx = -1
		a.vy = 1
	case DownRight:
		a.fx = 1
		a.fy = 1
		a.vx = 1
		a.vy = 1
	default:
		panic("unknown direction")
	}

	normalized := math.Sqrt(math.Pow(a.vx, 2) + math.Pow(a.vy, 2))
	a.px += (a.vx * a.speed) / normalized
	if a.px >= screenWidth-padding {
		a.px = screenWidth - padding - 1
	}
	if a.px <= padding {
		a.px = padding + 1
	}
	a.py += (a.vy * a.speed) / normalized
	if a.py >= screenHeight-padding-10 {
		a.py = screenHeight - padding - 11
	}
	if a.py <= padding {
		a.py = padding + 1
	}
	a.moveCmd.units--
	if a.moveCmd.units <= 0 {
		a.moveCmd = nil
	}
}

func (a *AI) Draw(screen *ebiten.Image, clock int) {
	op := &ebiten.DrawImageOptions{}
	if a.fx < 0 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(frameWidth, 0)
	}
	op.GeoM.Translate(
		a.px-frameWidth/2,
		a.py-frameHeight/2,
	)

	var sprite *ebiten.Image
	if a.vx != 0 || a.vy != 0 {
		sprite = runnerWalkingFrame(clock)
	} else {
		sprite = runnerWaitingFrame(clock)
	}

	screen.DrawImage(sprite, op)
}
