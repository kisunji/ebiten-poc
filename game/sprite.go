package game

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	frameWidth  = 32
	frameHeight = 32
)

var (
	runnerImage *ebiten.Image
)

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
