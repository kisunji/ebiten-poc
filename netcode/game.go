package netcode

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Char struct {
	fx, fy      int     // facing
	vx, vy      int     // velocity
	px, py      float64 // position
	speed       int
	clockOffset int

	sprite      func(clock int) *ebiten.Image

	// used by server only
	lastUpdatedTimer time.Time
}

var chars []*Char = make([]*Char, 0, 256)
