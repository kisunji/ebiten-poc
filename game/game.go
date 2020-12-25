package game

import (
	"bytes"
	"fmt"
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
	"github.com/kisunji/ebiten-poc/common"
)

func init() {
	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)
}

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
	g.SceneHandlers = map[Scene]SceneHandler{
		SceneStartMenu:    &StartMenu{client: g.Client},
		SceneNotConnected: &NotConnected{},
		SceneLobby:        NewLobby(g.Client),
		SceneMainGame: &MainGame{
			Op:     &ebiten.DrawImageOptions{},
			Speed:  5,
			Client: g.Client,
			Chars:  make([]*common.Char, common.MaxChars),
		},
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
		if g.Scene != SceneStartMenu {

		}
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
