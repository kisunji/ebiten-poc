package game

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
)

const (
	ScreenWidth   = 640
	ScreenHeight  = 480
	ScreenPadding = 10
)

func init() {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)
}

type Game struct {
	count  int
	Speed  int
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
	g.runner.Draw(screen, g.count/g.Speed)
	for _, ai := range g.ais {
		ai.Draw(screen, g.count/g.Speed)
	}
	msg := fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f\n", ebiten.CurrentTPS(), ebiten.CurrentFPS())
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) init() {
	defer func() {
		g.inited = true
	}()
	g.runner = Runner{
		speed:  1,
		px:     ScreenWidth / 2,
		py:     ScreenHeight / 2,
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
