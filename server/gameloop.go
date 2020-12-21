package server

import (
	"time"

	"github.com/kisunji/ebiten-poc/game"
)

const (
	updateFrequency = 1 * time.Second / 60
)

var chars = make([]*game.Char, game.MaxChars)

func Run() {
	previous := time.Now()
	var lag time.Duration

	for {
		current := time.Now()
		elapsed := current.Sub(previous)
		previous = current
		lag += elapsed

		for lag >= updateFrequency {
			update()
			lag -= updateFrequency
		}
	}
}

func update() {
	for _, char := range chars {
		if char == nil {
			continue
		}
		char.Move()
	}
}
