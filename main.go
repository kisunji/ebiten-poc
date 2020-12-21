package main

import (
	"log"

	"ebiten-poc/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	c := game.NewClient()
	err := c.Dial("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	go c.Listen()
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(game.ScreenWidth*2, game.ScreenHeight*2)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&game.Game{
		Op:     &ebiten.DrawImageOptions{},
		Speed:  5,
		Client: c,
		Chars:  make([]*game.Char, game.MaxChars),
	}); err != nil {
		log.Fatal(err)
	}
}
