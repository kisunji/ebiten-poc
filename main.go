package main

import (
	"bytes"
	"image"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

const (
	screenWidth  = 640
	screenHeight = 480
	padding      = 10

	frameWidth  = 32
	frameHeight = 32
)

var (
	runnerImage *ebiten.Image
)

type Runner struct {
	fx, fy float64 // facing
	vx, vy float64 // velocity
	px, py float64 // position
	speed  float64
}

func (r *Runner) Move() {
	isXPressed := false
	isYPressed := false
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		r.fx = 1
		r.vx = 1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		r.fx = -1
		r.vx = -1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		r.fy = -1
		r.vy = -1
		isYPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		r.fy = 1
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
	r.px += (r.vx * r.speed) / normalized
	if r.px >= screenWidth-padding {
		r.px = screenWidth - padding - 1
	}
	if r.px <= padding {
		r.px = padding + 1
	}
	r.py += (r.vy * r.speed) / normalized
	if r.py >= screenHeight-padding-10 {
		r.py = screenHeight - padding - 11
	}
	if r.py <= padding {
		r.py = padding + 1
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

	var sprite *ebiten.Image
	if r.vx != 0 || r.vy != 0 {
		sprite = runnerWalkingFrame(clock)
	} else {
		sprite = runnerWaitingFrame(clock)
	}

	screen.DrawImage(sprite, op)
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

type Game struct {
	count  int
	speed  int
	runner Runner
	ais    []*AI
	inited bool
}

func (g *Game) Update() error {
	if !g.inited {
		g.init()
	}
	g.count++
	g.runner.Move()
	for _, ai := range g.ais {
		ai.Move()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.runner.Draw(screen, g.count/g.speed)
	for _, ai := range g.ais {
		ai.Draw(screen, g.count/g.speed)
	}
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
		px:    screenWidth / 2,
		py:    screenHeight / 2,
	}
	g.ais = append(g.ais, newAI())
	g.ais = append(g.ais, newAI())
	g.ais = append(g.ais, newAI())
	g.ais = append(g.ais, newAI())
	g.ais = append(g.ais, newAI())
	for _, ai := range g.ais {
		go Run(ai)
	}
}

var AIId int

func newAI() *AI {
	AIId++
	ai := &AI{
		Runner: Runner{
			speed: 1,
			px:    padding + float64(rand.Intn(screenWidth-padding*2)),
			py:    padding + float64(rand.Intn(screenHeight-padding*2)),
		},
		id:      AIId,
		killSig: nil,
		moveCmd: nil,
		running: false,
	}
	return ai
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
