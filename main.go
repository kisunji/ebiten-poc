package main

import (
	_ "image/png"
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
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(game.ScreenWidth*2, game.ScreenHeight*2)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&game.Game{Speed: 5}); err != nil {
		log.Fatal(err)
	}
}
