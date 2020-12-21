package game

import "github.com/hajimehoshi/ebiten/v2"

func leftTouched() bool {
	for _, id := range ebiten.TouchIDs() {
		x, _ := ebiten.TouchPosition(id)
		if x < ScreenWidth/3 {
			return true
		}
	}
	return false
}

func rightTouched() bool {
	for _, id := range ebiten.TouchIDs() {
		x, _ := ebiten.TouchPosition(id)
		if x >= 2*ScreenWidth/3 {
			return true
		}
	}
	return false
}

func downTouched() bool {
	for _, id := range ebiten.TouchIDs() {
		_, y := ebiten.TouchPosition(id)
		if y >= 2*ScreenHeight/3 {
			return true
		}
	}
	return false
}

func upTouched() bool {
	for _, id := range ebiten.TouchIDs() {
		_, y := ebiten.TouchPosition(id)
		if y < ScreenHeight/3 {
			return true
		}
	}
	return false
}
