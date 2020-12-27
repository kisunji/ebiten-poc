package game

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/kisunji/ebiten-poc/common"
)

type Scene int

const (
	SceneStartMenu Scene = iota
	SceneLobby
	SceneMainGame
	SceneNotConnected
)

type Game struct {
	Client        *Client
	Scene         Scene
	SceneHandlers map[Scene]SceneHandler

	inited bool
}

type input struct {
	UpPressed     bool
	DownPressed   bool
	LeftPressed   bool
	RightPressed  bool
	ActionPressed bool
}

func (g *Game) init() {
	d := &Debouncer{
		input:    make(chan []byte),
		output:   g.Client.Send,
		duration: 33 * time.Millisecond,
		running:  false,
	}
	go d.debounce()
	g.SceneHandlers = map[Scene]SceneHandler{
		SceneStartMenu:    &StartMenu{client: g.Client},
		SceneNotConnected: &NotConnected{},
		SceneLobby:        NewLobby(g.Client),
		SceneMainGame:     NewMainGame(g.Client, d),
	}
	g.inited = true
}

func (g *Game) Update() error {
	if !g.inited {
		g.init()
	}
	switch g.Scene {
	case SceneStartMenu:
		handler := g.SceneHandlers[g.Scene]
		handler.Update()
		g.Scene = handler.Next()
	case SceneLobby:
		handler := g.SceneHandlers[g.Scene]
		handler.Update()
		g.Scene = handler.Next()
	case SceneNotConnected:
		handler := g.SceneHandlers[g.Scene]
		handler.Update()
		g.Scene = handler.Next()
	case SceneMainGame:
		handler := g.SceneHandlers[g.Scene]
		handler.Update()
		g.Scene = handler.Next()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// screen.Fill(color.RGBA{
	// 	R: 82,
	// 	G: 81,
	// 	B: 82,
	// 	A: 255,
	// })

	switch g.Scene {
	case SceneStartMenu:
		g.SceneHandlers[g.Scene].Draw(screen)
	case SceneNotConnected:
		g.SceneHandlers[g.Scene].Draw(screen)
	case SceneLobby:
		g.SceneHandlers[g.Scene].Draw(screen)
	case SceneMainGame:
		g.SceneHandlers[g.Scene].Draw(screen)
	}

	msg := fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f\nPing: %dms\n",
		ebiten.CurrentTPS(),
		ebiten.CurrentFPS(),
		g.Client.Latency,
	)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return common.ScreenWidth, common.ScreenHeight
}
