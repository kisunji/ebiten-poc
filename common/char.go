package common

import (
	"math"
	"math/rand"
	"time"

	"github.com/kisunji/ebiten-poc/pb"
)

type Char struct {
	Fx, Fy      int     // facing
	Vx, Vy      int     // velocity
	Px, Py      float64 // position
	Speed       int
	Offset      int // animation offset
	AttackFrame int
	IsDead      bool

	// used by server only
	lastUpdatedTimer time.Time
}

func NewChar() *Char {
	return &Char{
		Fx:     rand.Intn(2) - 1,
		Fy:     rand.Intn(2) - 1,
		Px:     float64(ScreenPadding + rand.Intn(ScreenWidth-ScreenPadding*3)),
		Py:     float64(ScreenPadding + rand.Intn(ScreenHeight-ScreenPadding*3)),
		Speed:  1,
		Offset: rand.Intn(10),
	}
}

type Chars []*Char

func (cc Chars) UpdateFromData(input *pb.UpdateEntity) {
	c := cc[input.Index]
	if c == nil {
		c = NewChar()
		cc[input.Index] = c
	}
	c.Px = input.Px
	c.Py = input.Py
	c.Fx = int(input.Fx)
	c.Fy = int(input.Fy)
	c.Vx = int(input.Vx)
	c.Vy = int(input.Vy)
	c.Speed = int(input.Speed)
	c.AttackFrame = int(input.AttackFrame)
	c.IsDead = input.IsDead
}

func (c *Char) ProcessInput(input *pb.Input) {
	c.Vx = 0
	c.Vy = 0
	if input.ActionPressed {
		c.Attack()
		return
	}
	if input.RightPressed {
		c.Vx = 1
	}
	if input.LeftPressed {
		c.Vx = -1
	}
	if input.UpPressed {
		c.Vy = -1
	}
	if input.DownPressed {
		c.Vy = 1
	}
	if input.RightPressed == input.LeftPressed {
		c.Vx = 0
	}
	if input.UpPressed == input.DownPressed {
		c.Vy = 0
	}
	c.lastUpdatedTimer = time.Now()
}

// Attack sets decrements attackFrame
func (c *Char) Attack() {
	if c.AttackFrame == 0 {
		c.AttackFrame = 20
		return
	}
	c.AttackFrame--
}

func (c *Char) Attacking() bool {
	return c.AttackFrame > 0
}

func (c *Char) Move() {
	if c.IsDead {
		return
	}
	if c.Vx == 0 && c.Vy == 0 {
		return
	}
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
	if c.Vx > 0 {
		c.Fx = 1
	} else if c.Vx < 0 {
		c.Fx = -1
	} else {
		c.Fx = 0
	}
	if c.Vy > 0 {
		c.Fy = 1
	} else if c.Vy < 0 {
		c.Fy = -1
	} else {
		c.Fy = 0
	}
}

func (c *Char) ImpactSite(radius float64) (x, y float64) {
	normalized := math.Sqrt(math.Pow(float64(c.Fx), 2) + math.Pow(float64(c.Fy), 2))
	if normalized == 0 {
		return
	}
	x = c.Px + float64(c.Fx)*radius/normalized
	y = c.Py + float64(c.Fy)*radius/normalized
	return x, y
}
