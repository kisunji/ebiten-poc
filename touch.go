package main

import "github.com/hajimehoshi/ebiten/v2"

func leftTouched() bool {
	for _, id := range ebiten.TouchIDs() {
		x, _ := ebiten.TouchPosition(id)
		if x < screenWidth/3 {
			return true
		}
	}
	return false
}

func rightTouched() bool {
	for _, id := range ebiten.TouchIDs() {
		x, _ := ebiten.TouchPosition(id)
		if x >= screenWidth/3 {
			return true
		}
	}
	return false
}

func downTouched() bool {
	for _, id := range ebiten.TouchIDs() {
		x, _ := ebiten.TouchPosition(id)
		if x < screenHeight/3 {
			return true
		}
	}
	return false
}

func upTouched() bool {
	for _, id := range ebiten.TouchIDs() {
		x, _ := ebiten.TouchPosition(id)
		if x >= screenHeight/3 {
			return true
		}
	}
	return false
}

