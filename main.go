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
	x, y   float64
	speed  float64
}

func (r *Runner) Move() {
	isXPressed := false
	isYPressed := false
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		r.facing.x = 1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		r.facing.x = -1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		r.facing.y = -1
		isYPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		r.facing.y = 1
		isYPressed = true
	}
	if !isXPressed {
		r.facing.x = 0
	}
	if !isYPressed {
		r.facing.y = 0
	}
	if !isXPressed && !isYPressed {
		return
	}
	normalized := math.Sqrt(math.Pow(r.facing.x, 2) + math.Pow(r.facing.y, 2))
	r.x += (r.facing.x * r.speed) / normalized
	if r.x >= screenWidth-padding {
		r.x = screenWidth - padding - 1
	}
	if r.x <= padding {
		r.x = padding + 1
	}
	r.y += (r.facing.y * r.speed) / normalized
	if r.y >= screenHeight-padding-10 {
		r.y = screenHeight - padding - 11
	}
	if r.y <= padding {
		r.y = padding + 1
	}
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
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		g.runner.x-frameWidth/2,
		g.runner.y-frameHeight/2,
	)
	i := (g.count / g.speed) % frameNum
	sx, sy := frameOX+i*frameWidth, frameOY
	screen.DrawImage(
		runnerImage.SubImage(
			image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image),
		op,
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth / 2, outsideHeight / 2
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
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{speed: 5}); err != nil {
		log.Fatal(err)
	}
}
