package game

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	frameWidth  = 32
	frameHeight = 32
)

var (
	runnerImage *ebiten.Image
)

type Runner struct {
	fx, fy      int     // facing
	vx, vy      int     // velocity
	px, py      float64 // position
	speed       int
	clockOffset int
	sprite      func(clock int) *ebiten.Image
}

func (r *Runner) Move() {
	isXPressed := false
	isYPressed := false
	if ebiten.IsKeyPressed(ebiten.KeyD) || rightTouched() {
		r.fx = 1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || leftTouched() {
		r.fx = -1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) || upTouched() {
		r.fy = -1
		isYPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || downTouched() {
		r.fy = 1
		isYPressed = true
	}
	r.vx = r.fx * r.speed
	r.vy = r.fy * r.speed
	if !isXPressed {
		r.vx = 0
	}
	if !isYPressed {
		r.vy = 0
	}
	if r.vx != 0 || r.vy != 0 {
		r.sprite = runnerWalkingFrame
	} else {
		r.sprite = runnerWaitingFrame
	}
	if !isXPressed && !isYPressed {
		return
	}

	normalized := math.Sqrt(math.Pow(float64(r.vx), 2) + math.Pow(float64(r.vy), 2))
	r.px += float64(r.vx) / normalized
	if r.px >= ScreenWidth-ScreenPadding {
		r.px = ScreenWidth - ScreenPadding - 1
	}
	if r.px <= ScreenPadding {
		r.px = ScreenPadding + 1
	}
	r.py += float64(r.vy) / normalized
	if r.py >= ScreenHeight-ScreenPadding-10 {
		r.py = ScreenHeight - ScreenPadding - 11
	}
	if r.py <= ScreenPadding {
		r.py = ScreenPadding + 1
	}
}

func (r *Runner) Draw(screen *ebiten.Image, clock int) {
	op := &ebiten.DrawImageOptions{}
	if r.fx < 0 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(frameWidth, 0)
	}
	op.GeoM.Translate(
		r.px-frameWidth/2,
		r.py-frameHeight/2,
	)

	screen.DrawImage(r.sprite(clock+r.clockOffset), op)
}

func runnerWalkingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 8
		frameWidth  = 32
		frameHeight = 32
		frameOX     = 0
		frameOY     = 32
	)

	i := clock % frameNum
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func runnerWaitingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 5
		frameWidth  = 32
		frameHeight = 32
		frameOX     = 0
		frameOY     = 0
	)
	i := clock / 2 % frameNum
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}
