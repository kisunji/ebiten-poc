package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"
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
)

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
