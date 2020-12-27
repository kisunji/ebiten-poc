package game

import (
	"bytes"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	runnerImage *ebiten.Image
)

func init() {
	img, _, err := image.Decode(bytes.NewReader(sprite_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)
}

func downRestingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 8
		frameWidth  = 16
		frameHeight = 32
		frameOX     = 0
		frameOY     = 0
	)
	i := clock % frameNum
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func downRightRestingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 8
		frameWidth  = 16
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

func rightRestingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 8
		frameWidth  = 16
		frameHeight = 32
		frameOX     = 0
		frameOY     = 64
	)
	i := clock % frameNum
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func upRightRestingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 8
		frameWidth  = 16
		frameHeight = 32
		frameOX     = 0
		frameOY     = 96
	)
	i := clock % frameNum
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func upRestingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 8
		frameWidth  = 16
		frameHeight = 32
		frameOX     = 0
		frameOY     = 128
	)
	i := clock % frameNum
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func downRightAttackingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 4
		frameWidth  = 40
		frameHeight = 56
		frameOX     = 0
		frameOY     = 160
	)
	i := frameNum - (clock+4)/5
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func downAttackingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 4
		frameWidth  = 40
		frameHeight = 56
		frameOX     = 0
		frameOY     = 216
	)
	i := frameNum - (clock+4)/5
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func rightAttackingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 4
		frameWidth  = 40
		frameHeight = 56
		frameOX     = 0
		frameOY     = 272
	)
	i := frameNum - (clock+4)/5
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func upRightAttackingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 4
		frameWidth  = 40
		frameHeight = 56
		frameOX     = 0
		frameOY     = 328
	)
	i := frameNum - (clock+4)/5
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func upAttackingFrame(clock int) *ebiten.Image {
	const (
		frameNum    = 4
		frameWidth  = 40
		frameHeight = 56
		frameOX     = 0
		frameOY     = 384
	)
	i := frameNum - (clock+4)/5
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}

func deadFrame() *ebiten.Image {
	const (
		frameWidth  = 26
		frameHeight = 32
		frameOX     = 0
		frameOY     = 440
	)
	i := 1
	sx := frameOX + i*frameWidth
	sy := frameOY

	return runnerImage.SubImage(
		image.Rect(sx, sy, sx+frameWidth, sy+frameHeight),
	).(*ebiten.Image)
}