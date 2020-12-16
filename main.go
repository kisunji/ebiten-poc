package main

import (
	"bytes"
	"image"
	_ "image/png"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

const (
	screenWidth  = 320
	screenHeight = 240
	padding      = 10

	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameNum    = 8
)

var (
	runnerImage *ebiten.Image
)

type Point struct {
	x, y float64
}

type Runner struct {
	facing Point
	vx, vy float64
	x, y   float64
	speed  float64
}

func (r *Runner) Move() {
	isXPressed := false
	isYPressed := false
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		r.facing.x = 1
		r.vx = 1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		r.facing.x = -1
		r.vx = -1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		r.facing.y = -1
		r.vy = -1
		isYPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		r.facing.y = 1
		r.vy = 1
		isYPressed = true
	}
	if !isXPressed {
		r.vx = 0
	}
	if !isYPressed {
		r.vy = 0
	}
	if !isXPressed && !isYPressed {
		return
	}
	normalized := math.Sqrt(math.Pow(r.vx, 2) + math.Pow(r.vy, 2))
	r.x += (r.vx * r.speed) / normalized
	if r.x >= screenWidth-padding {
		r.x = screenWidth - padding - 1
	}
	if r.x <= padding {
		r.x = padding + 1
	}
	r.y += (r.vy * r.speed) / normalized
	if r.y >= screenHeight-padding-10 {
		r.y = screenHeight - padding - 11
	}
	if r.y <= padding {
		r.y = padding + 1
	}
}

func (r *Runner) Draw(screen *ebiten.Image, clock int) {
	op := &ebiten.DrawImageOptions{}
	if r.facing.x < 0 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(frameWidth, 0)
	}
	op.GeoM.Translate(
		r.x-frameWidth/2,
		r.y-frameHeight/2,
	)

	i := clock % frameNum
	sx, sy := frameOX+i*frameWidth, frameOY
	screen.DrawImage(
		runnerImage.SubImage(
			image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image),
		op,
	)
}

type Game struct {
	count  int
	speed  int
	runner Runner
	inited bool
}

func (g *Game) Update() error {
	if !g.inited {
		g.init()
	}
	g.count++
	g.runner.Move()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.runner.Draw(screen, g.count/g.speed)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) init() {
	defer func() {
		g.inited = true
	}()
	g.runner = Runner{
		speed: 1,
		x:     screenWidth / 2,
		y:     screenHeight / 2,
	}
}

func main() {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{speed: 5}); err != nil {
		log.Fatal(err)
	}
}
