package main

import (
	"context"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kisunji/ebiten-poc/common"
	"github.com/kisunji/ebiten-poc/game"
)

func main() {
	c := game.NewClient()
	err := c.Dial("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	go c.Listen(context.Background())
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(common.ScreenWidth*2, common.ScreenHeight*2)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&game.Game{
		Op:     &ebiten.DrawImageOptions{},
		Speed:  5,
		Client: c,
		Chars:  make([]*common.Char, common.MaxChars),
	}); err != nil {
		log.Fatal(err)
	}
}
