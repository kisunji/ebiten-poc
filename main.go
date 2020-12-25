package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kisunji/ebiten-poc/common"
	"github.com/kisunji/ebiten-poc/game"
)

func main() {
	c := game.NewClient()

	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(common.ScreenWidth*2, common.ScreenHeight*2)
	ebiten.SetWindowTitle(common.GameTitle)

	g := &game.Game{
		Client: c,
		Scene:  game.SceneStartMenu,
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
