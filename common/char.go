package common

import (
	"math"
	"math/rand"
	"time"

	"github.com/kisunji/ebiten-poc/pb"
)

func NewChar() *Char {
	return &Char{
		Fx:    1,
		Fy:    1,
		Px:    float64(ScreenPadding + rand.Intn(ScreenWidth-ScreenPadding*3)),
		Py:    float64(ScreenPadding + rand.Intn(ScreenHeight-ScreenPadding*3)),
		Speed: 1,
	}
}

func NewCharAt(px, py float64) *Char {
	return &Char{
		Fx:    1,
		Fy:    1,
		Px:    px,
		Py:    py,
		Speed: 1,
	}
}

type Char struct {
	Fx, Fy int     // facing
	Vx, Vy int     // velocity
	Px, Py float64 // position
	Speed  int

	// used by server only
	lastUpdatedTimer time.Time
}

type Chars []*Char

func (cc Chars) UpdateFromData(input *pb.UpdateEntity) {
	c := cc[input.Index]
	if c == nil {
		c = &Char{}
		cc[input.Index] = c
	}
	c.Px = input.Px
	c.Py = input.Py
	c.Fx = int(input.Fx)
	c.Fy = int(input.Fy)
	c.Vx = int(input.Vx)
	c.Vy = int(input.Vy)
	c.Speed = int(input.Speed)
}

func (c *Char) ProcessInput(input *pb.Input) {
	if input.RightPressed {
		c.Fx = 1
	}
	if input.LeftPressed {
		c.Fx = -1
	}
	if input.UpPressed {
		c.Fy = -1
	}
	if input.DownPressed {
		c.Fy = 1
	}
	c.Vx = c.Fx * c.Speed
	c.Vy = c.Fy * c.Speed
	if input.RightPressed == input.LeftPressed {
		c.Vx = 0
	}
	if input.UpPressed == input.DownPressed {
		c.Vy = 0
	}
	c.lastUpdatedTimer = time.Now()
}

func (c *Char) Move() {
	normalized := math.Sqrt(math.Pow(float64(c.Vx), 2) + math.Pow(float64(c.Vy), 2))
	if normalized == 0 {
		return
	}
	c.Px += float64(c.Vx) / normalized
	if c.Px >= ScreenWidth-ScreenPadding {
		c.Px = ScreenWidth - ScreenPadding - 1
	}
	if c.Px <= ScreenPadding {
		c.Px = ScreenPadding + 1
	}
	c.Py += float64(c.Vy) / normalized
	if c.Py >= ScreenHeight-ScreenPadding-10 {
		c.Py = ScreenHeight - ScreenPadding - 11
	}
	if c.Py <= ScreenPadding {
		c.Py = ScreenPadding + 1
	}
}
