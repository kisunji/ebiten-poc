package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		r.fx = 1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		r.fx = -1
		isXPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		r.fy = -1
		isYPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
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
	if r.px >= screenWidth-padding {
		r.px = screenWidth - padding - 1
	}
	if r.px <= padding {
		r.px = padding + 1
	}
	r.py += float64(r.vy) / normalized
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
	msg := fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f\n", ebiten.CurrentTPS(), ebiten.CurrentFPS())
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) init() {
	defer func() {
		g.inited = true
	}()
	g.runner = Runner{
		speed:  1,
		px:     screenWidth / 2,
		py:     screenHeight / 2,
		sprite: runnerWaitingFrame,
	}
	numAI := 50
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < numAI; i++ {
		g.ais = append(g.ais, newAI(1))
	}
	for _, ai := range g.ais {
		go Run(ai)
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
